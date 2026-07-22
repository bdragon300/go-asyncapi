package doc

import (
	"bytes"
	"fmt"
	"os"
	"slices"

	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type MvCmd struct {
	Locations []string `arg:"positional,required" help:"Nodes to move. If -t is omitted, the last LOCATION is considered as DESTINATION. Format: file.{yaml|yml|json}[#/path/to/node | GLOBBING_PATTERN]" placeholder:"LOCATION"`

	AutoCreate       bool   `arg:"--auto-create,-a" help:"Automatically create a destination node if missing"`
	Link             bool   `arg:"--link,-l" help:"Insert a $ref into the source location after moving. Does not apply to nodes evaluated recursively"`
	Recursive        bool   `arg:"--recursive,-r" help:"Recursively move dependencies by following the $refs."`
	RecursiveShallow bool   `arg:"--recursive-shallow,-S" help:"Like -r, but move only the first level of dependencies"`
	Headless         bool   `arg:"--headless,-H" help:"Move only dependencies, excluding the matched nodes"`
	Force            bool   `arg:"--force,-f" help:"Overwrite existing nodes on conflict"`
	Interactive      bool   `arg:"--interactive,-i" help:"Resolve conflicts interactively"`
	To               string `arg:"--to,-t" help:"Destination location. Once specified, all LOCATION arguments are moved into DESTINATION" placeholder:"DESTINATION"`
	DisableRewriting bool   `arg:"--disable-rewriting" help:"Do not rewrite $refs"`

	Indent int `arg:"--indent" help:"Output document indentation width" placeholder:"SPACES"`
}

func cliMv(cmd *MvCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	sourcePatterns, destPattern, err := parseCliPatterns(cmd.Locations, cmd.To)
	if err != nil {
		return fmt.Errorf("%w: %w", common2.ErrInvalidCLIArgument, err)
	}

	locator := common2.GetLocator(cmdConfig)

	// Keep every loaded document in a cache keyed by its absolute location. A document that is both a source and the
	// destination (or referenced by several patterns) must be represented by a single tree, so we move nodes out of it
	// and write it back consistently.
	docs := make(map[string]*common2.DocumentTree)

	absOutputPath := &jsonpointer.JSONPointer{FSPath: absLocation(destPattern.JSONPointer)}
	outputContents := common2.NewDocumentTree(absOutputPath)
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
	var headNodes []*types.RawNode
	for _, pattern := range sourcePatterns {
		logger.Debug("Loading document", "path", pattern.Location())
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
		heads := findNodesByPattern(inputContents.RawNode, pattern)
		logger.Trace("Found nodes", "count", len(heads), "pattern", pattern)
		if !cmdConfig.Doc.Mv.Headless {
			relocatees = lo.Map(heads, func(n *types.RawNode, _ int) relocatedNode {
				logger.Debug("Found node", "path", n, "pattern", pattern)
				return relocatedNode{node: n, isDependency: false, isDirectDependency: false}
			})
		}
		if cmdConfig.Doc.Mv.Recursive || cmdConfig.Doc.Mv.RecursiveShallow {
			logger.Trace("Collecting dependencies for head nodes", "count", len(heads), "recursive", cmdConfig.Doc.Mv.Recursive, "recursiveShallow", cmdConfig.Doc.Mv.RecursiveShallow)
			deps := lo.FlatMap(heads, func(n *types.RawNode, _ int) []relocatedNode {
				r := collectDependencies(n, []*common2.DocumentTree{inputContents}, locator, !cmdConfig.Doc.Mv.RecursiveShallow)
				logger.Debug("Found dependencies for node", "path", n, "count", len(r))
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
			formatIndent: cmdConfig.Doc.Mv.Indent,
		}
		chlog, err := relocateNodes(inputContents, outputContents, relocatees, destPattern, flags, cmdConfig.Doc.Mv.AutoCreate)
		if err != nil {
			return fmt.Errorf("relocate nodes: %w", err)
		}
		changeLog = append(changeLog, chlog...)
		headNodes = append(headNodes, heads...)
	}
	headPaths := lo.UniqMap(headNodes, func(n *types.RawNode, _ int) string { return n.AbsPointerString() })

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
	if changeLog, err = applyMoves(docs, changeLog, headPaths, cmdConfig.Doc.Mv.Link); err != nil {
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
		ef, _, format := compiler.GuessDocumentFormat(origin.Location(), cmdConfig.Doc.Mv.Indent)
		if format == "" {
			return fmt.Errorf("cannot determine the output document format: %q", origin.Location())
		}
		if err = ef(buf).Encode(d); err != nil {
			return fmt.Errorf("marshal document %q: %w", origin, err)
		}
		if err = os.WriteFile(origin.Location(), buf.Bytes(), 0o644); err != nil {
			logger.Warn("Cannot write the document, skipping", "file", origin.Location(), "error", err)
			continue
		}
	}

	return nil
}

func applyMoves(docs map[string]*common2.DocumentTree, changeLog []changeLogEntry, heads []string, createHeadLinks bool) ([]changeLogEntry, error) {
	logger := log.GetLogger("")
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

		// Replace the head node with a $ref to its new location if user has requested it.
		toCreateRef := createHeadLinks && lo.ContainsBy(heads, func(h string) bool { return h == entry.From.String() })
		if !toCreateRef {
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

func loadDocumentCached(p *jsonpointer.JSONPointer, docs map[string]*common2.DocumentTree, locator common2.DocumentLocator) (*common2.DocumentTree, error) {
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
