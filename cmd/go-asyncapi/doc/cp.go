package doc

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

type CpCmd struct {
	Locations []string `arg:"positional,required" help:"Document with optional node path or globbing pattern. If -t is omitted, the last LOCATION is considered as DESTINATION. Format: file.{yaml|yml|json}[#/path/to/node | GLOBBING_PATTERN]" placeholder:"LOCATION"`

	To               string `arg:"--to,-t" help:"Copy all LOCATION arguments into DESTINATION" placeholder:"DESTINATION"`
	FollowRefs       bool   `arg:"--follow-refs,-r" help:"Follow $refs and copy the referenced nodes recursively"`
	ShallowRefs      bool   `arg:"--shallow-refs,-s" help:"Limit the following $refs only one level deep. Requires --follow-refs"`
	Headless         bool   `arg:"--headless" help:"Exclude nodes. Makes sense with -r or -s"`
	Force            bool   `arg:"--force,-f" help:"Overwrite existing nodes on conflict"`
	Interactive      bool   `arg:"--interactive,-i" help:"Interactive mode"`
	DisableRewriting bool   `arg:"--disable-rewriting" help:"Do not rewrite $refs"`

	Indent int    `arg:"--indent" help:"Output document indentation width" placeholder:"SPACES"`
	Format string `arg:"--format" help:"Output format. Possible values: yaml, json" placeholder:"FORMAT"`
}

func cliCp(cmd *CpCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	if cmd.ShallowRefs && !cmd.FollowRefs {
		return fmt.Errorf("%w: --shallow-refs requires --follow-refs", common2.ErrInvalidCLIArgument)
	}

	sourcePatterns, destPattern, err := parseCliPatterns(cmd.Locations, cmd.To)
	if err != nil {
		return fmt.Errorf("%w: %w", common2.ErrInvalidCLIArgument, err)
	}

	locator := common2.GetLocator(cmdConfig)
	absOutputPath := &jsonpointer.JSONPointer{FSPath: absLocation(destPattern.JSONPointer)}
	outputContents := common2.NewDocumentTree(absOutputPath)
	if _, err = os.Stat(destPattern.FSPath); err == nil {
		if outputContents, err = loadDocument(destPattern.JSONPointer, locator); err != nil {
			return fmt.Errorf("load %q: %w", destPattern.FSPath, err)
		}
	}

	var changeLog []changeLogEntry
	var relocateesCount int
	var relocatees []relocatedNode
	for _, pattern := range sourcePatterns {
		logger.Debug("Loading document", "path", pattern.Location())
		inputContents, err := loadDocument(pattern.JSONPointer, locator)
		if err != nil {
			return fmt.Errorf("load document %s: %w", pattern.Location(), err)
		}

		logger.Trace("Searching for nodes matching the pattern", "pattern", pattern)
		heads := findNodesByPattern(inputContents.RawNode, pattern)
		logger.Trace("Found nodes", "count", len(heads), "pattern", pattern)
		if !cmdConfig.Doc.Cp.Headless {
			relocatees = lo.Map(heads, func(n *types.RawNode, _ int) relocatedNode {
				logger.Debug("Found node", "path", n, "pattern", pattern)
				return relocatedNode{node: n, isDependency: false, isDirectDependency: false}
			})
		}
		if cmdConfig.Doc.Cp.FollowRefs {
			logger.Trace("Collecting dependencies for head nodes", "count", len(heads), "followRefs", cmdConfig.Doc.Cp.FollowRefs, "shallowRefs", cmdConfig.Doc.Cp.ShallowRefs)
			deps := lo.FlatMap(heads, func(n *types.RawNode, _ int) []relocatedNode {
				r := collectDependencies(n, []*common2.DocumentTree{inputContents}, locator, !cmdConfig.Doc.Cp.ShallowRefs)
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
			force:        cmdConfig.Doc.Cp.Force,
			interactive:  cmdConfig.Doc.Cp.Interactive,
			formatIndent: cmdConfig.Doc.Cp.Indent,
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

	if !cmdConfig.Doc.Cp.DisableRewriting {
		logger.Debug("Rewriting $refs in the destination document", "path", destPattern.FSPath)
		rewriteRefs(outputContents, locator, changeLog)
	}

	logger.Info("Writing file", "file", outputContents.AbsOriginDocumentPath())
	buf := bytes.NewBuffer(nil)
	enc, err := getDocumentEncoder(buf, cmdConfig.Doc.Cp.Format, cmdConfig.Doc.Cp.Indent)
	if err != nil {
		return fmt.Errorf("get encoder: %w", err)
	}

	if err = enc.Encode(outputContents); err != nil {
		return fmt.Errorf("marshal document: %w", err)
	}
	if err = os.WriteFile(outputContents.AbsOriginDocumentPath().Location(), buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write unmerged file %q: %w", outputContents.AbsOriginDocumentPath(), err)
	}

	return nil
}

func ensureMandatoryNodes(outputContents *common2.DocumentTree) []changeLogEntry {
	logger := log.GetLogger("")

	logger.Trace("Adding a new node", "path", outputContents.AbsOriginDocumentPath().Join("info"))
	infoNode := types.NewEmptyRawNode(types.RawNodeKindObject, []string{"info"}, outputContents.AbsOriginDocumentPath())
	infoNode.Set("title", types.NewScalarRawNode([]string{"info", "title"}, "Untitled API", outputContents.AbsOriginDocumentPath()))
	infoNode.Set("version", types.NewScalarRawNode([]string{"info", "version"}, "1.0.0", outputContents.AbsOriginDocumentPath()))
	outputContents.SetAtStart("info", infoNode)

	logger.Trace("Adding a new node", "path", outputContents.AbsOriginDocumentPath().Join("asyncapi"))
	outputContents.SetAtStart("asyncapi", types.NewScalarRawNode([]string{"asyncapi"}, "3.0.0", outputContents.AbsOriginDocumentPath()))

	// Double-check that all mandatory root keys are present
	mandatoryExistingKeys := lo.PickByKeys(maps.Collect(outputContents.Entries()), lo.ToAnySlice(asyncapiMandatoryRootPaths()))
	if len(mandatoryExistingKeys) != len(asyncapiMandatoryRootPaths()) {
		panic("some mandatory AsyncAPI root keys are missing, this is a bug")
	}

	return []changeLogEntry{
		{
			From: nil,
			To:   lo.ToPtr(outputContents.AbsOriginDocumentPath().Join("asyncapi")),
			Move: false,
		},
		{
			From: nil,
			To:   lo.ToPtr(outputContents.AbsOriginDocumentPath().Join("info")),
			Move: false,
		},
	}
}

func parseCliPatterns(args []string, to string) ([]cliPattern, cliPattern, error) {
	var patterns []cliPattern

	for _, a := range args {
		pt, err := parseCliPattern(a)
		if err != nil {
			return nil, cliPattern{}, err
		}
		if pt.Location() == "" {
			return nil, cliPattern{}, fmt.Errorf("argument must contain a document path, got %q", a)
		}
		patterns = append(patterns, pt)
	}
	if to != "" {
		toPt, err := parseCliPattern(to)
		if err != nil {
			return nil, cliPattern{}, fmt.Errorf("parse %q: %w", to, err)
		}
		patterns = append(patterns, toPt)
	}

	switch {
	case len(patterns) == 0:
		return nil, cliPattern{}, fmt.Errorf("empty arguments")
	case len(patterns) == 1:
		return nil, cliPattern{}, fmt.Errorf("missed source arguments")
	case patterns[len(patterns)-1].URI != nil:
		return nil, cliPattern{}, fmt.Errorf("writing to documents by URL is not supported, got %q", args[len(args)-1])
	}

	return patterns[:len(patterns)-1], patterns[len(patterns)-1], nil
}

type relocatedNode struct {
	node               *types.RawNode
	isDependency       bool
	isDirectDependency bool
}

type copyNodeFlags struct {
	force        bool
	interactive  bool
	formatIndent int
}

func relocateNodes(sourceDoc, destDoc *common2.DocumentTree, relocatees []relocatedNode, destPattern cliPattern, flags copyNodeFlags) ([]changeLogEntry, error) {
	logger := log.GetLogger("")
	if sourceDoc == nil || destDoc == nil {
		panic("sourceDoc or destDoc is nil, this is a bug")
	}

	var changeLog []changeLogEntry
	headsCount := lo.CountBy(relocatees, func(r relocatedNode) bool { return !r.isDependency && !r.isDirectDependency })
	for _, r := range relocatees {
		reason := "selected"
		switch {
		case r.isDirectDependency:
			reason = "a direct dependency"
		case r.isDependency:
			reason = "an indirect dependency"
		}

		// Resolve and normalize the destination container path:
		// - "/foo/dest/": error if destination does not exist
		// - "/foo/dest":
		//   - if destination does not exist: if there is only one head relocatee, relocate it with renaming, otherwise error
		//   - if destination exists, relocate a node as-is
		// - "/" -> root
		// - empty path -> same path as the source node
		// - if it's a dependency located directly in root section ("#/servers", "#/channels", "#/operations"), relocate it to the same path
		// - if it's a dependency located elsewhere in document, then put it to "components" section depending on its kind
		// - "/foo/dest//", "//" are invalid
		dContainerPath := destPattern.Pointer
		dKey := lo.LastOrEmpty(r.node.Path())
		var dContainer *types.RawNode
		switch {
		case r.isDependency:
			logger.Trace("Node is dependency", "path", r.node.Path(), "document", sourceDoc.AbsOriginDocumentPath())
			rootSections, _ := asyncapiEntitiesSectionPaths()
			rootSection, inRootSection := lo.Find(rootSections, func(s string) bool { return len(r.node.Path()) == 2 && r.node.Path()[0] == s })
			dContainerPath = []string{rootSection}
			if !inRootSection {
				logger.Trace("Node is dependency not in root section", "path", r.node.Path(), "document", sourceDoc.AbsOriginDocumentPath())
				componentsKey := asyncapiResolveComponentsKey(r.node.Path())
				if componentsKey == "" {
					return nil, fmt.Errorf("cannot auto determine the destination node for dependency %q, try to narrow down a pattern or to copy this node manually", r.node.AbsPointerString())
				}
				dContainerPath = []string{"components", componentsKey}
			}
		case len(destPattern.Pointer) > 0:
			logger.Trace("Destination path is non-empty", "path", destPattern.Pointer, "document", destDoc.AbsOriginDocumentPath(), "headNodes", headsCount)
			dContainerPath = lo.TrimRight(destPattern.Pointer, []string{""})
			if len(destPattern.Pointer)-len(dContainerPath) > 1 {
				return nil, fmt.Errorf("path %q cannot end with several slashes", jsonpointer.PointerString(destPattern.Pointer...))
			}
			if destDoc.GetByPath(dContainerPath) == nil {
				logger.Trace("Destination node does not exist, evaluating new path", "path", dContainerPath, "document", destDoc.AbsOriginDocumentPath())
				switch {
				case lo.LastOrEmpty(destPattern.Pointer) == "":
					// "/foo/dest/"
					return nil, fmt.Errorf("destination node %q does not exist", destDoc.AbsOriginDocumentPath().Join(dContainerPath...))
				case headsCount > 1:
					return nil, fmt.Errorf("cannot relocate %d nodes from %q to path %q, create it first or pick only one node", headsCount, sourceDoc.AbsOriginDocumentPath(), destDoc.AbsOriginDocumentPath().Join(dContainerPath...))
				case len(dContainerPath) > 0:
					// Relocating a node with rename. If destination is empty, keep the original destination
					dKey = lo.LastOrEmpty(dContainerPath)
					dContainerPath = dContainerPath[:len(dContainerPath)-1]
				}
			}
		case len(destPattern.Pointer) == 0 && len(r.node.Path()) > 0:
			logger.Trace("Destination path is empty, using the source node path", "path", r.node.Path(), "document", sourceDoc.AbsOriginDocumentPath())
			dContainerPath = r.node.Path()[:len(r.node.Path())-1]
		}

		logger.Trace("Destination node", "path", dContainerPath, "document", destDoc.AbsOriginDocumentPath())
		dContainer = destDoc.GetByPath(dContainerPath)
		if dContainer == nil {
			// Create the destination container if it doesn't exist yet.
			logger.Trace("Container node does not exist, creating", "path", dContainerPath, "document", destDoc.AbsOriginDocumentPath())
			dContainer = types.NewEmptyRawNode(types.RawNodeKindObject, dContainerPath, destDoc.AbsOriginDocumentPath())
			if err := destDoc.SetNodeByPath(dContainer); err != nil {
				return nil, fmt.Errorf("create node %q in document %q: %w", jsonpointer.PointerString(dContainerPath...), destDoc.AbsOriginDocumentPath(), err)
			}
		}
		if dContainer.Kind() == types.RawNodeKindScalar {
			return nil, fmt.Errorf("destination node %q is scalar, it must be array or object", dContainer.AbsPointerString())
		}

		logger.Info("Relocating node", "src", jsonpointer.PointerString(r.node.Path()...), "dest", jsonpointer.PointerString(dContainerPath...), "reason", reason)
		change, err := copyNode(r.node.DeepCopy(), dContainer, dKey, destDoc.AbsOriginDocumentPath(), sourceDoc.AbsOriginDocumentPath(), flags)
		if err != nil {
			return nil, fmt.Errorf("copy node %v: %w", jsonpointer.PointerString(r.node.Path()...), err)
		}
		if change == nil {
			continue
		}

		changeLog = append(changeLog, *change)
	}

	return changeLog, nil
}

func collectDependencies(node *types.RawNode, docs []*common2.DocumentTree, locator common2.DocumentLocator, recursiveRefs bool) []relocatedNode {
	logger := log.GetLogger("")

	if node == nil {
		return nil
	}
	if node.Kind() == types.RawNodeKindScalar {
		return nil
	}

	var res []relocatedNode
	queue := collectRefs(node)
	for i := 0; i < len(queue); i++ {
		r := queue[i]
		ref, err := parseRefRawNode(r)
		if err != nil {
			logger.Warn("Failed to parse $ref, skipping", "path", r.Path(), "error", err)
			continue
		}

		// Explicit document path where this $ref pointed to. For internal $ref, it's the file itself
		referredPath := r.AbsOriginDocumentPath()
		if ref.Location() != "" {
			// Resolve location in external $ref relative to it's origin document location
			if referredPath, err = locator.ResolveURL(r.AbsOriginDocumentPath(), ref); err != nil {
				logger.Error("Failed to resolve $ref, skipping", "path", r.Path(), "value", ref, "error", err)
				continue
			}
		}

		logger.Trace("Found $ref", "path", r.Path(), "pointer", ref.String())
		referredDoc, found := lo.Find(docs, func(d *common2.DocumentTree) bool {
			return d.AbsOriginDocumentPath().Location() == absLocation(referredPath)
		})
		if !found {
			logger.Trace("$ref points to 3rd-party document, skipping", "path", r.Path(), "pointer", ref.String())
			// 3rd-party document
			continue
		}
		refNode := referredDoc.GetByPath(ref.Pointer)
		if refNode == nil {
			logger.Warn("Invalid $ref, skipping", "path", r, "pointer", ref.PointerString())
			continue
		}

		logger.Debug("Found dependency node", "path", r.Path(), "pointer", ref.String(), "document", referredDoc.AbsOriginDocumentPath())
		res = append(res, relocatedNode{node: refNode, isDependency: true, isDirectDependency: lo.HasPrefix(r.Path(), node.Path())})
		if recursiveRefs {
			queue = append(queue, collectRefs(refNode)...)
		}
		queue = lo.UniqBy(queue, func(n *types.RawNode) string { return n.AbsPointerString() })
	}

	return res
}

func copyNode(sNode, dContainer *types.RawNode, dKey string, dDoc, sDoc *jsonpointer.JSONPointer, flags copyNodeFlags) (*changeLogEntry, error) {
	var err error
	logger := log.GetLogger("")

	if dContainer.Kind() == types.RawNodeKindArray {
		dContainer.Append(sNode)
		return &changeLogEntry{
			From: lo.ToPtr(sDoc.Join(sNode.Path()...)),
			To:   lo.ToPtr(dDoc.Join(dContainer.Path()...).Join(strconv.Itoa(dContainer.Len() - 1))),
			Move: false,
		}, nil
	}

	dNode, conflict := dContainer.Get(dKey)
	if !conflict {
		logger.Trace("Copying node", "destination", dDoc.Join(dContainer.Path()...))
		dContainer.Set(dKey, sNode)
		return &changeLogEntry{
			From: lo.ToPtr(sDoc.Join(sNode.Path()...)),
			To:   lo.ToPtr(dDoc.Join(dContainer.Path()...).Join(dKey)),
			Move: false,
		}, nil
	}

	logger.Debug("Conflict, node already exists", "destination", dDoc.Join(dNode.Path()...), "source", sDoc.Join(sNode.Path()...))
	if dNode.Equal(sNode) {
		logger.Debug("Auto-resolving the conflict: skipping because nodes are identical")
		// Add a log entry for duplicate node to make sure that $ref to it will also be rewritten
		return &changeLogEntry{From: lo.ToPtr(sDoc.Join(sNode.Path()...)), To: lo.ToPtr(dDoc.Join(dNode.Path()...)), Move: false}, nil
	}

	var newPath *jsonpointer.JSONPointer
	switch {
	case flags.force:
		logger.Debug("Auto-resolving the conflict: force overwriting")
		newPath = lo.ToPtr(dDoc.Join(dNode.Path()...))
	case flags.interactive:
		// Ask user to resolve the conflict
		if newPath, err = promptResolveConflict(dContainer, dNode, sNode, dDoc, sDoc, flags); err != nil {
			return nil, fmt.Errorf("resolve conflict for key %q: %w", dKey, err)
		}
		if newPath == nil {
			logger.Info("Conflict resolved: ignoring the conflicting node", "path", sNode)
		} else {
			logger.Info("Conflict resolved", "path", sNode, "newPath", newPath)
		}
	default:
		return nil, fmt.Errorf("node %q already exists, use -f flag to force overwrite or -i to resolve the conflict interactively", dNode.AbsPointerString())
	}
	if newPath == nil {
		logger.Debug("Conflict resolved: skipping node")
		return nil, nil
	}

	// Apply a change
	dKey = lo.LastOrEmpty(newPath.Pointer)
	logger.Debug("Copying a node", "source", sNode.Path(), "destination", dContainer.Path())
	dContainer.Set(dKey, sNode)
	return &changeLogEntry{
		From: lo.ToPtr(sDoc.Join(sNode.Path()...)),
		To:   newPath,
		Move: false,
	}, nil
}

func promptResolveConflict(dContainer, dNode, sNode *types.RawNode, dDoc, sDoc *jsonpointer.JSONPointer, flags copyNodeFlags) (*jsonpointer.JSONPointer, error) {
	fmt.Printf("Conflict! Cannot relocate %q to %q: destination node already exists\n", sNode.AbsPointerString(), dNode.AbsPointerString())

	var action string
	for {
		if flags.interactive {
			fmt.Print("Choose option: show \033[1;4md\033[0miff/\033[1;4mi\033[0mgnore/\033[1;4mo\033[0mverwrite/\033[1;4mr\033[0mename [d/i/o/R]: ")
			inp := bufio.NewScanner(os.Stdin)
			if !inp.Scan() {
				if inp.Err() != nil {
					return nil, fmt.Errorf("read user choice: %w", inp.Err())
				}
				fmt.Println("Aborted.")
				return nil, common2.ErrInterruptedByUser // EOF, user pressed Ctrl+D
			}
			action = strings.ToLower(inp.Text())
		}

		switch action {
		case "i":
			// Ignore the new node, keep the existing one.
			return nil, nil
		case "o":
			// Overwrite the existing node with the new one.
			return lo.ToPtr(dDoc.Join(dNode.Path()...)), nil
		case "r", "":
			// Rename the new node and add it to the destination.
			k := lo.LastOrEmpty(sNode.Path())
			for i := 1; ; i++ {
				key := fmt.Sprintf("%s%d", k, i)
				if _, exists := dContainer.Get(key); !exists {
					newPath := append(dNode.Path()[:len(dNode.Path())-1], key) // nolint:gocritic
					return lo.ToPtr(dDoc.Join(newPath...)), nil
				}
			}
		}

		// Print diff in unified patch format
		dmp := diffmatchpatch.New()
		diff := dmp.DiffMain(
			formatDiffContent(dNode, dNode.Path(), flags.formatIndent),
			formatDiffContent(sNode, dNode.Path(), flags.formatIndent),
			false,
		)
		diffOutput := dmp.DiffPrettyText(diff)
		fmt.Printf("\u001B[32m+++ %s\n\u001B[0m", sDoc.String())
		fmt.Printf("\u001B[31m--- %s\n\u001B[0m", dDoc.String())
		fmt.Print(diffOutput)
	}
}

func formatDiffContent(contents *types.RawNode, nodePath []string, indentWidth int) string {
	var indentLvl int
	var b strings.Builder

	for _, p := range nodePath {
		b.WriteString(strings.Repeat(" ", indentWidth*indentLvl))
		b.WriteString(p)
		b.WriteString(":\n")
		indentLvl++
	}

	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(indentWidth)
	if err := enc.Encode(contents); err != nil {
		panic(fmt.Errorf("marshal diff content to yaml: %w", err))
	}

	lines := strings.Split(buf.String(), "\n")
	for _, line := range lines {
		b.WriteString(strings.Repeat(" ", indentWidth*indentLvl))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
