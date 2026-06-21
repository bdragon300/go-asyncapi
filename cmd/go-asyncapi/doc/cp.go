package doc

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/gobwas/glob"
	"github.com/samber/lo"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"
)

type CpCmd struct {
	Locations []string `arg:"positional,required" help:"Nodes to copy. If -t is omitted, the last LOCATION is considered as DESTINATION. Format: file.{yaml|yml|json}[#/path/to/node | GLOBBING_PATTERN]" placeholder:"LOCATION"`

	To               string `arg:"--to,-t" help:"Copy all LOCATION arguments into DESTINATION" placeholder:"DESTINATION"`
	Recursive        bool   `arg:"--recursive,-r" help:"Copy nodes recursively"`
	Shallow          bool   `arg:"--shallow,-s" help:"Copy nodes recursively only with direct dependencies"`
	Headless         bool   `arg:"--headless" help:"Exclude nodes. Makes sense with -r or -s"`
	Force            bool   `arg:"--force,-f" help:"Overwrite existing nodes on conflict"`
	Interactive      bool   `arg:"--interactive,-i" help:"Interactive mode"`
	DisableRewriting bool   `arg:"--disable-rewriting" help:"Do not rewrite $refs"`
}

func cliCp(cmd *CpCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	sourcePatterns, destPattern, err := parseCliPatterns(cmd.Locations, cmd.To)
	if err != nil {
		return fmt.Errorf("%w: %w", common2.ErrInvalidCLIArgument, err)
	}

	locator := common2.GetLocator(cmdConfig)
	absOutputPath := &jsonpointer.JSONPointer{FSPath: absLocation(destPattern.JSONPointer)}
	outputContents := newDocumentTree(absOutputPath)
	if _, err = os.Stat(destPattern.FSPath); err == nil {
		if outputContents, err = loadDocument(destPattern.JSONPointer, locator); err != nil {
			return fmt.Errorf("load %q: %w", destPattern.FSPath, err)
		}
	}

	var changeLog []changeLogEntry
	var relocateesCount int
	var relocatees []relocatedNode
	for _, pattern := range sourcePatterns {
		logger.Debug("Loading document", "path", pattern)
		inputContents, err := loadDocument(pattern.JSONPointer, locator)
		if err != nil {
			return fmt.Errorf("load document %s: %w", pattern.Location(), err)
		}

		logger.Trace("Searching for nodes matching the pattern", "pattern", pattern)
		matchedNodes := findNodes(inputContents.RawNode, pattern)
		logger.Trace("Found nodes", "count", len(matchedNodes), "pattern", pattern)
		if !cmdConfig.Doc.Cp.Headless {
			relocatees = lo.Map(matchedNodes, func(n *types.RawNode, _ int) relocatedNode {
				logger.Debug("Found node", "path", n.AbsPointerString(), "pattern", pattern)
				return relocatedNode{node: n, isDependency: false, isDirectDependency: false}
			})
		}
		if cmdConfig.Doc.Cp.Recursive || cmdConfig.Doc.Cp.Shallow {
			logger.Trace("Collecting dependencies for matched nodes", "count", len(matchedNodes), "recursive", cmdConfig.Doc.Cp.Recursive, "shallow", cmdConfig.Doc.Cp.Shallow)
			deps := lo.FlatMap(matchedNodes, func(n *types.RawNode, _ int) []relocatedNode {
				r := collectDependencies(n, []*documentTree{inputContents}, locator, !cmdConfig.Doc.Cp.Shallow)
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
			force:        cmdConfig.Doc.Cp.Force,
			interactive:  cmdConfig.Doc.Cp.Interactive,
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

	if !cmdConfig.Doc.Cp.DisableRewriting {
		logger.Debug("Rewriting $refs in the destination document", "path", destPattern.FSPath)
		rewriteRefs(outputContents, locator, changeLog)
	}

	logger.Info("Writing file", "file", outputContents.AbsOriginDocumentPath())
	buf := bytes.NewBuffer(nil)
	enc, err := getDocumentEncoder(buf, cmdConfig)
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

func ensureMandatoryNodes(outputContents *documentTree) []changeLogEntry {
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

type cliPattern struct {
	*jsonpointer.JSONPointer
	pattern glob.Glob
}

func (c cliPattern) MatchPath(p []string) bool {
	if len(c.Pointer) == 0 || slices.Equal(p, c.Pointer) {
		return true
	}
	return c.pattern.Match(strings.Join(p, "/"))
}

func parseCliPatterns(args []string, to string) ([]cliPattern, cliPattern, error) {
	var patterns []cliPattern

	parse := func(arg string) (cliPattern, error) {
		p, err := jsonpointer.Parse(arg)
		if err != nil {
			return cliPattern{}, fmt.Errorf("parse %q: %w", arg, err)
		}
		gl, err := glob.Compile(strings.Join(p.Pointer, "/"))
		if err != nil {
			return cliPattern{}, fmt.Errorf("compile glob pattern %q: %w", arg, err)
		}
		return cliPattern{JSONPointer: p, pattern: gl}, nil
	}

	for _, a := range args {
		pt, err := parse(a)
		if err != nil {
			return nil, cliPattern{}, err
		}
		if pt.Location() == "" {
			return nil, cliPattern{}, fmt.Errorf("argument must contain a document path, got %q", a)
		}
		patterns = append(patterns, pt)
	}
	if to != "" {
		toPt, err := parse(to)
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
	quiet        bool
	formatIndent int
}

func relocateNodes(sourceDoc, destDoc *documentTree, relocatees []relocatedNode, destPattern cliPattern, flags copyNodeFlags) ([]changeLogEntry, error) {
	logger := log.GetLogger("")
	if sourceDoc == nil || destDoc == nil {
		panic("sourceDoc or destDoc is nil, this is a bug")
	}

	var changeLog []changeLogEntry
	for _, r := range relocatees {
		reason := "selected"
		switch {
		case r.isDirectDependency:
			reason = "a direct dependency"
		case r.isDependency:
			reason = "an indirect dependency"
		}

		// Resolve and normalize the destination container path:
		// - "/foo/bar" -> "/foo/bar", create the node if it does not exists
		// - "/foo/bar/" -> "/foo/bar", error if node does not exist
		// - "/" -> root
		// - empty path -> same path as the source node
		// - if it's a dependency located directly in root section ("#/servers", "#/channels", "#/operations"), copy it to the same path
		// - if it's a dependency located elsewhere in document, then put it to "components" section depending on its kind
		// - "/foo/bar//", "//" are invalid
		dContainerPath := destPattern.Pointer
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
					return nil, fmt.Errorf("cannot determine the components section for the dependency %q", r.node.AbsPointerString())
				}
				dContainerPath = []string{"components", componentsKey}
			}
		case len(destPattern.Pointer) > 0 && lo.LastOrEmpty(destPattern.Pointer) == "":
			logger.Trace("Destination path ends with empty segment, it must exist", "path", destPattern.Pointer, "document", destDoc.AbsOriginDocumentPath())
			dContainerPath = lo.TrimRight(destPattern.Pointer, []string{""})
			if len(destPattern.Pointer)-len(dContainerPath) > 1 {
				return nil, fmt.Errorf("path %q: it cannot end with empty segments", jsonpointer.PointerString(destPattern.Pointer...))
			}
			if destDoc.GetByPath(dContainerPath) == nil {
				return nil, fmt.Errorf("destination node %q does not exist", destDoc.AbsOriginDocumentPath().Join(dContainerPath...))
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
			return nil, fmt.Errorf("%q node is scalar, it must be array or object", dContainer.AbsPointerString())
		}

		logger.Info("Relocating node", "src", jsonpointer.PointerString(r.node.Path()...), "dest", jsonpointer.PointerString(dContainerPath...), "reason", reason)
		change, err := copyNode(r.node.DeepCopy(), dContainer, destDoc.AbsOriginDocumentPath(), sourceDoc.AbsOriginDocumentPath(), flags)
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

func findNodes(node *types.RawNode, pattern cliPattern) []*types.RawNode {
	if node == nil {
		return nil
	}

	// Exclude the root node from matching, because we copying nodes by keys, and the root node doesn't have a key.
	if len(node.Path()) > 0 && pattern.MatchPath(node.Path()) {
		return []*types.RawNode{node}
	}

	var res []*types.RawNode
	if node.Kind() == types.RawNodeKindObject || node.Kind() == types.RawNodeKindArray {
		for _, e := range node.Entries() {
			res = append(res, findNodes(e, pattern)...)
		}
	}
	return res
}

func collectDependencies(node *types.RawNode, docs []*documentTree, locator common2.DocumentLocator, recursive bool) []relocatedNode {
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
		referredDoc, found := lo.Find(docs, func(d *documentTree) bool { return d.AbsOriginDocumentPath().Location() == absLocation(referredPath) })
		if !found {
			logger.Trace("$ref points to 3rd-party document, skipping", "path", r.Path(), "pointer", ref.String())
			// 3rd-party document
			continue
		}
		refNode := referredDoc.GetByPath(ref.Pointer)
		if refNode == nil {
			logger.Warn("Invalid $ref, skipping", "path", r.AbsPointerString(), "pointer", ref.PointerString())
			continue
		}

		logger.Debug("Found dependency node", "path", r.Path(), "pointer", ref.String(), "document", referredDoc.AbsOriginDocumentPath())
		res = append(res, relocatedNode{node: refNode, isDependency: true, isDirectDependency: lo.HasPrefix(r.Path(), node.Path())})
		if recursive {
			queue = append(queue, collectRefs(refNode)...)
		}
		queue = lo.UniqBy(queue, func(n *types.RawNode) string { return n.AbsPointerString() })
	}

	return res
}

func copyNode(sNode, dContainer *types.RawNode, dDoc, sDoc *jsonpointer.JSONPointer, flags copyNodeFlags) (*changeLogEntry, error) {
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

	nodeKey := lo.LastOrEmpty(sNode.Path())
	dNode, conflict := dContainer.Get(nodeKey)
	if !conflict {
		logger.Trace("Copying node", "destination", dDoc.Join(dContainer.Path()...))
		dContainer.Set(nodeKey, sNode)
		return &changeLogEntry{
			From: lo.ToPtr(sDoc.Join(sNode.Path()...)),
			To:   lo.ToPtr(dDoc.Join(dContainer.Path()...).Join(nodeKey)),
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
			return nil, fmt.Errorf("resolve conflict for key %q: %w", nodeKey, err)
		}
		if newPath == nil {
			logger.Info("Conflict resolved: ignoring the conflicting node", "path", sNode.AbsPointerString())
		} else {
			logger.Info("Conflict resolved", "path", sNode.AbsPointerString(), "newPath", newPath)
		}
	default:
		return nil, fmt.Errorf("node %q already exists, use -f flag to force rewrite or -i to resolve the conflict interactively", dNode.AbsPointerString())
	}
	if newPath == nil {
		logger.Debug("Conflict resolved: skipping node")
		return nil, nil
	}

	// Apply a change
	dKey := lo.LastOrEmpty(newPath.Pointer)
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
		// TODO: fix msg vvv
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
