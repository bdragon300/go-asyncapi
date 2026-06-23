package doc

import (
	"bytes"
	"fmt"
	"os"
	"slices"

	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type MvCmd struct {
	Locations []string `arg:"positional,required" help:"Nodes to move. If -t is omitted, the last LOCATION is considered as DESTINATION. Format: file.{yaml|yml|json}[#/path/to/node | GLOBBING_PATTERN]" placeholder:"LOCATION"`

	To               string `arg:"--to,-t" help:"Move all LOCATION arguments into DESTINATION" placeholder:"DESTINATION"`
	FollowRefs       bool   `arg:"--follow-refs,-r" help:"Follow $refs and move the referenced nodes recursively"`
	ShallowRefs      bool   `arg:"--shallow-refs,-s" help:"Limit the following $refs only one level deep. Requires --follow-refs"`
	Headless         bool   `arg:"--headless" help:"Exclude nodes. Makes sense with -r or -s"`
	Force            bool   `arg:"--force,-f" help:"Overwrite existing nodes on conflict"`
	Interactive      bool   `arg:"--interactive,-i" help:"Interactive mode"`
	DisableRewriting bool   `arg:"--disable-rewriting" help:"Do not rewrite $refs"`
}

func cliMv(cmd *MvCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	if cmd.ShallowRefs && !cmd.FollowRefs {
		return fmt.Errorf("%w: --shallow-refs requires --follow-refs", common2.ErrInvalidCLIArgument)
	}

	sourcePatterns, destPattern, err := parseCliPatterns(cmd.Locations, cmd.To)
	if err != nil {
		return fmt.Errorf("%w: %w", common2.ErrInvalidCLIArgument, err)
	}

	locator := common2.GetLocator(cmdConfig)

	// Keep every loaded document in a cache keyed by its absolute location. A document that is both a source and the
	// destination (or referenced by several patterns) must be represented by a single tree, so we move nodes out of it
	// and write it back consistently.
	docs := make(map[string]*documentTree)

	absOutputPath := &jsonpointer.JSONPointer{FSPath: absLocation(destPattern.JSONPointer)}
	outputContents := newDocumentTree(absOutputPath)
	if _, err = os.Stat(destPattern.FSPath); err == nil {
		if outputContents, err = loadDocumentCached(destPattern.JSONPointer, docs, locator); err != nil {
			return fmt.Errorf("load %q: %w", destPattern.FSPath, err)
		}
	} else {
		docs[absLocation(destPattern.JSONPointer)] = outputContents
	}

	// Relocation step. Identical to `doc cp`: matched nodes (and, optionally, their dependencies) are copied into the
	// destination document, producing a changelog of From->To node movements.
	var changeLog []changeLogEntry
	var relocateesCount int
	var relocatees []relocatedNode
	var locationsToRemove []string
	for _, pattern := range sourcePatterns {
		logger.Debug("Loading document", "path", pattern)
		inputContents, err := loadDocumentCached(pattern.JSONPointer, docs, locator)
		if err != nil {
			return fmt.Errorf("load document %s: %w", pattern.Location(), err)
		}

		// If input document is moved entirely (pattern has only file name) then remove it at the end
		// (except the corner case when input and output are the same document).
		if len(pattern.Pointer) == 0 && inputContents.AbsOriginDocumentPath().Location() != outputContents.AbsOriginDocumentPath().Location() {
			locationsToRemove = append(locationsToRemove, inputContents.AbsOriginDocumentPath().Location())
		}

		logger.Trace("Searching for nodes matching the pattern", "pattern", pattern)
		matchedNodes := findNodes(inputContents.RawNode, pattern)
		logger.Trace("Found nodes", "count", len(matchedNodes), "pattern", pattern)
		if !cmdConfig.Doc.Mv.Headless {
			relocatees = lo.Map(matchedNodes, func(n *types.RawNode, _ int) relocatedNode {
				logger.Debug("Found node", "path", n.AbsPointerString(), "pattern", pattern)
				return relocatedNode{node: n, isDependency: false, isDirectDependency: false}
			})
		}
		if cmdConfig.Doc.Mv.FollowRefs {
			logger.Trace("Collecting dependencies for matched nodes", "count", len(matchedNodes), "followRefs", cmdConfig.Doc.Mv.FollowRefs, "shallowRefs", cmdConfig.Doc.Mv.ShallowRefs)
			deps := lo.FlatMap(matchedNodes, func(n *types.RawNode, _ int) []relocatedNode {
				r := collectDependencies(n, []*documentTree{inputContents}, locator, !cmdConfig.Doc.Mv.ShallowRefs)
				logger.Debug("Found dependencies for node", "path", n.AbsPointerString(), "count", len(r))
				return r
			})
			deps = lo.UniqBy(deps, func(n relocatedNode) string { return n.node.AbsPointerString() })
			relocatees = append(relocatees, deps...)
		}
		if len(relocatees) == 0 {
			logger.Warn("No nodes matched the pattern, skipping", "pattern", pattern)
			continue
		}
		relocateesCount += len(relocatees)

		// Deduplicate nodes, since some nodes may have the same dependencies
		oldLen := len(relocatees)
		relocatees = lo.UniqBy(relocatees, func(n relocatedNode) string { return n.node.AbsPointerString() })
		logger.Debug("Deduplicated nodes", "count", len(relocatees)-oldLen)

		logger.Info("Relocating nodes", "source", inputContents.AbsOriginDocumentPath(), "destination", outputContents.AbsOriginDocumentPath())
		flags := copyNodeFlags{
			force:        cmdConfig.Doc.Mv.Force,
			interactive:  cmdConfig.Doc.Mv.Interactive,
			quiet:        cmdConfig.Quiet,
			formatIndent: cmdConfig.Doc.Indent,
		}
		chlog, err := relocateNodes(inputContents, outputContents, relocatees, destPattern, flags)
		if err != nil {
			return fmt.Errorf("relocate nodes: %w", err)
		}
		changeLog = append(changeLog, chlog...)
	}

	// Return error if no nodes were relocated.
	if relocateesCount == 0 {
		return fmt.Errorf("%w: no nodes were picked to relocate", common2.ErrBadResult)
	} else if len(changeLog) == 0 {
		logger.Warn("No changes were made")
		return nil
	}

	if !outputContents.Has("asyncapi") {
		logger.Debug("Adding mandatory AsyncAPI entities to the destination document", "path", destPattern.FSPath)
		changeLog = append(changeLog, ensureMandatoryNodes(outputContents)...)
	}

	logger.Trace("Applying moves to the documents", "count", len(changeLog))
	if changeLog, err = applyMoves(docs, changeLog); err != nil {
		return fmt.Errorf("apply moves: %w", err)
	}

	if !cmdConfig.Doc.Mv.DisableRewriting {
		for loc, d := range docs {
			if slices.Contains(locationsToRemove, d.AbsOriginDocumentPath().Location()) {
				logger.Trace("Skipping $ref rewriting for a document that is going to be removed", "path", loc)
				continue
			}
			logger.Debug("Rewriting $refs in the document", "path", loc)
			rewriteRefs(d, locator, changeLog)
		}
	}

	for loc, d := range docs {
		origin := d.AbsOriginDocumentPath()
		if origin.URI != nil {
			logger.Warn("Document was loaded from a URL, cannot overwrite it, skipping", "url", loc)
			continue
		}
		if slices.Contains(locationsToRemove, d.AbsOriginDocumentPath().Location()) {
			logger.Info("Removing document", "file", origin)
			if err := os.Remove(origin.Location()); err != nil {
				logger.Warn("Cannot remove the document, skipping", "file", origin.Location(), "error", err)
			}
			continue
		}

		logger.Info("Writing file", "file", origin)
		buf := bytes.NewBuffer(nil)
		enc, err := getDocumentEncoder(buf, cmdConfig)
		if err != nil {
			return fmt.Errorf("get encoder: %w", err)
		}
		if err = enc.Encode(d); err != nil {
			return fmt.Errorf("marshal document %q: %w", origin, err)
		}
		if err = os.WriteFile(origin.Location(), buf.Bytes(), 0o644); err != nil {
			logger.Warn("Cannot write the document, skipping", "file", origin.Location(), "error", err)
			continue
		}
	}

	return nil
}

func applyMoves(docs map[string]*documentTree, changeLog []changeLogEntry) ([]changeLogEntry, error) {
	logger := log.GetLogger("")
	rootSections, componentsSections := asyncapiEntitiesSectionPaths()
	entitySections := append(
		lo.Chunk(rootSections, 1), // [1,2,3] -> [[1],[2],[3]]
		lo.Map(componentsSections, func(s string, _ int) []string { return []string{"components", s} })...,
	)
	mandatorySections := lo.Chunk(asyncapiMandatoryRootPaths(), 1) // [1,2,3] -> [[1],[2],[3]]

	for i := 0; i < len(changeLog); i++ {
		entry := &changeLog[i]

		// Skip entries that don't describe a relocated source node (e.g. the mandatory nodes added to the destination).
		if entry.From == nil || entry.To == nil || len(entry.From.Pointer) == 0 {
			continue
		}

		moveIntoSelfOrChild := absLocation(entry.From) == absLocation(entry.To) && lo.HasPrefix(entry.To.Pointer, entry.From.Pointer)
		if moveIntoSelfOrChild {
			// Source moves into its own subtree, which is not allowed.
			if len(entry.From.Pointer) < len(entry.To.Pointer) {
				return nil, fmt.Errorf("cannot move node %q into its child node %q", entry.From.String(), entry.To.String())
			}
			// Source and destination are the same node, skip
			logger.Warn("Source and destination nodes are the same, skipping", "node", entry.From.String())
			continue
		}

		srcDoc, ok := docs[absLocation(entry.From)]
		if !ok {
			logger.Warn("Source document is not loaded, cannot move the node, skipping", "node", entry.From.String())
			continue
		}

		// If node is in mandatory sections ("#/asyncapi", "#/info"), then skip it to remove
		isMandatoryNode := lo.SomeBy(mandatorySections, func(p []string) bool { return lo.HasPrefix(entry.From.Pointer, p) })
		if isMandatoryNode {
			logger.Warn("Mandatory node cannot be moved, copy it instead", "node", entry.From.String())
			continue
		}

		entry.Move = true

		// If node is entity located in root sections ("#/servers", "#/channels", "#/messages") or it's a component
		// or it's a root section itself, then remove it
		rootEntityOrComponent := lo.SomeBy(entitySections, func(s []string) bool {
			return lo.HasPrefix(entry.From.Pointer, s) && len(entry.From.Pointer)-len(s) == 1
		})
		if rootEntityOrComponent || len(entry.From.Pointer) == 1 {
			logger.Debug("Removing the node", "node", entry.From.String())
			if _, err := srcDoc.DeleteByPath(entry.From.Pointer); err != nil {
				return nil, fmt.Errorf("remove node %q: %w", entry.From.String(), err)
			}
			continue
		}

		logger.Debug("Replacing the node with a $ref to its new location", "node", entry.From.String(), "ref", entry.To.String())
		dstDoc := docs[absLocation(entry.To)]
		ref := jsonpointer.JSONPointer{Pointer: entry.From.Pointer}
		newRef := rewriteRef(&ref, dstDoc.AbsOriginDocumentPath(), srcDoc.AbsOriginDocumentPath(), srcDoc, changeLog)

		refNode := types.NewEmptyRawNode(types.RawNodeKindObject, entry.From.Pointer, srcDoc.AbsOriginDocumentPath())
		refPath := append(slices.Clone(entry.From.Pointer), "$ref")
		refNode.Set("$ref", types.NewScalarRawNode(refPath, newRef.String(), srcDoc.AbsOriginDocumentPath()))
		if err := srcDoc.SetNodeByPath(refNode); err != nil {
			return nil, fmt.Errorf("replace node %q with a $ref: %w", entry.From.String(), err)
		}
		changeLog = append(changeLog, changeLogEntry{From: nil, To: entry.From, Move: true})
	}

	return changeLog, nil
}

func loadDocumentCached(p *jsonpointer.JSONPointer, docs map[string]*documentTree, locator common2.DocumentLocator) (*documentTree, error) {
	key := absLocation(p)
	if d, ok := docs[key]; ok {
		return d, nil
	}
	d, err := loadDocument(p, locator)
	if err != nil {
		return nil, err
	}
	docs[key] = d
	return d, nil
}
