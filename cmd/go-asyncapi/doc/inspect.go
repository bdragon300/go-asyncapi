package doc

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"golang.org/x/term"
)

type InspectCmd struct {
	Location string `arg:"positional,required" help:"Document with optional node name or globbing pattern. Format: file.{yaml|yml|json}[#/path/to/node | GLOBBING_PATTERN]" placeholder:"LOCATION"`

	Entities           string `arg:"--entities,-e" help:"Comma-separated list of entities to show or 'help' to list all available entities and exit'" placeholder:"ENTITIES"`
	Main               bool   `arg:"--main" help:"Show only servers, channels and operations defined in the root sections of the document"`
	Components         bool   `arg:"--components" help:"Show only entities defined in components section of the documents"`
	Recursive          bool   `arg:"--recursive,-r" help:"Show all nested nodes recursively"`
	RecursiveAll       bool   `arg:"--recursive-all,-R" help:"Like -r, but also showing all nested JSON Schema objects and fully unfolding $ref chains"`
	List               bool   `arg:"--list,-l" help:"Show the result as a list"`
	FollowExternalRefs bool   `arg:"--follow-external-refs,-f" help:"Follow the $refs pointing to other documents"`
	AllowRemoteRefs    bool   `arg:"--allow-remote-refs,-F" help:"Follow the $refs pointing to URLs. Implies --follow-external-refs."`

	EntryStyle common2.DocInspectPathStyle `arg:"--entry-style,-s" help:"Style of the output entries. Options: human, human-no-color, json-pointer, yq" placeholder:"STYLE"`

	LocatorRootDir string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`
}

func cliInspect(cmd *InspectCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	switch cmdConfig.Doc.Inspect.EntryStyle {
	case common2.DocInspectPathStyleJSONPointer, common2.DocInspectPathStyleYq, common2.DocInspectPathStyleHuman, common2.DocInspectPathStyleHumanNoColor:
	default:
		return fmt.Errorf("%w: unknown path format %q", common2.ErrInvalidCLIArgument, cmdConfig.Doc.Inspect.EntryStyle)
	}

	if strings.EqualFold(cmdConfig.Doc.Inspect.Entities, "help") {
		fmt.Println("Available entity types: " + strings.Join(asyncapiEntities(), ", ") + ", other")
		return nil
	}

	logger.Info("Hint: Use --quiet to suppress the logging output")

	// Parse --entities
	entities, err := parseEntityFilterExpression(cmdConfig.Doc.Inspect.Entities)
	if err != nil {
		return fmt.Errorf("%w: parse entity arg: %w", common2.ErrInvalidCLIArgument, err)
	}

	// Parse locations
	pattern, err := parseCliPattern(cmd.Location)
	if err != nil {
		return fmt.Errorf("%w: %w", common2.ErrInvalidCLIArgument, err)
	}

	logger.Debug("Loading document", "path", pattern.Location())
	locator := common2.GetLocator(cmdConfig)
	inputContents, err := loadDocument(pattern.JSONPointer, locator)
	if err != nil {
		return fmt.Errorf("load document %s: %w", pattern.Location(), err)
	}

	var inputNodes []*types.RawNode
	if len(pattern.Pointer) > 0 {
		logger.Trace("Collecting nodes by pattern", "pattern", pattern)
		inputNodes = findNodesByPattern(inputContents.RawNode, pattern)
	} else {
		logger.Trace("Collecting nodes from top-level and components sections")
		inputNodes = collectTopLevelAndComponentsNodes(inputContents)
	}
	if len(inputNodes) == 0 {
		logger.Warn("Nothing to inspect", "pattern", pattern)
		return nil
	}

	var renderNodes []*entityNode
	docs := map[string]*common2.DocumentTree{absLocation(inputContents.AbsOriginDocumentPath()): inputContents}
	for _, n := range inputNodes {
		logger.Debug("Inspecting node", "path", n)
		inspected, err := inspectNode(
			n,
			nil,
			docs,
			0,
			common2.GetLocator(cmdConfig),
			cmdConfig.Doc.Inspect.FollowExternalRefs || cmdConfig.Doc.Inspect.AllowRemoteReferences,
			cmdConfig.Doc.Inspect.AllowRemoteReferences,
		)
		if err != nil {
			return fmt.Errorf("inspect %s: %w", pattern.Location(), err)
		}
		renderNodes = append(renderNodes, inspected)
	}
	renderNodes = lo.UniqBy(renderNodes, func(n *entityNode) string {
		return n.node.AbsPointerString()
	})
	logger.Trace("Rendering nodes topology", "nodes", len(renderNodes), "list", cmdConfig.Doc.Inspect.List, "recursive", cmdConfig.Doc.Inspect.Recursive, "recursiveAll", cmdConfig.Doc.Inspect.RecursiveAll)
	displayNodesTopology(renderNodes, inputContents.AbsOriginDocumentPath(), entities, cmdConfig)

	return nil
}

// parseEntityFilterExpression turns the argument with comma-separated "entity" values into a set of ">entity" keys.
// A special argument "other", which means "non-entity nodes", produces an empty string key.
func parseEntityFilterExpression(s string) ([]string, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}

	entities := asyncapiEntities()

	var res []string
	for _, e := range strings.Split(s, ",") {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if e == "other" {
			res = append(res, "")
		} else if !lo.Contains(entities, e) {
			return nil, fmt.Errorf("unknown entity %q", e)
		}
		res = append(res, ">"+e)
	}
	return res, nil
}

func collectTopLevelAndComponentsNodes(inputContents *common2.DocumentTree) []*types.RawNode {
	rootSections, componentsSections := asyncapiEntitiesSectionPaths()
	entitySections := append(
		lo.Chunk(rootSections, 1), // [1,2,3] -> [[1],[2],[3]]
		lo.Map(componentsSections, func(s string, _ int) []string { return []string{"components", s} })...,
	)
	return lo.FlatMap(entitySections, func(section []string, _ int) []*types.RawNode {
		n := inputContents.GetByPath(section)
		if n == nil {
			return nil
		}
		var res []*types.RawNode
		for _, e := range n.Entries() {
			res = append(res, e)
		}
		return res
	})
}

func inspectNode(node *types.RawNode, visited []*types.RawNode, docs map[string]*common2.DocumentTree, level int, locator common2.DocumentLocator, external, remote bool) (*entityNode, error) {
	entity := asyncapiEntityByPath(node.Path())
	res := entityNode{
		node:   node,
		entity: entity,
		cycle:  slices.Contains(visited, node),
	}
	if res.cycle || node.Kind() == types.RawNodeKindScalar {
		return &res, nil
	}

	if node.Has("$ref") {
		refNode, err := parseRefRawNode(node)
		if err != nil {
			return nil, fmt.Errorf("parse $ref: %w", err)
		}
		res.ref = refNode
		if refNode.Location() == "" || external {
			// $ref here can also point to a node that is not an entity, e.g. a message payload or a scalar node.
			// It's ok, we record this node as well
			resolvedNode, err := resolveRefNode(refNode, node, docs, locator, external, remote)
			if err != nil {
				return nil, fmt.Errorf("resolve $ref: %w", err)
			}
			res.refTo, err = inspectNode(resolvedNode, append(visited, node), docs, level+1, locator, external, remote)
			if err != nil {
				return nil, err
			}
			return &res, nil
		}
	}

	for _, e := range collectInnerEntityNodes(node) {
		child, err := inspectNode(e, append(visited, node), docs, level+1, locator, external, remote)
		if err != nil {
			return nil, err
		}
		res.children = append(res.children, child)
	}

	return &res, nil
}

func collectInnerEntityNodes(node *types.RawNode) []*types.RawNode {
	var res []*types.RawNode
	for _, child := range node.Entries() {
		if child.Kind() == types.RawNodeKindScalar {
			continue
		}

		entity := asyncapiEntityByPath(child.Path())
		if strings.HasPrefix(entity, ">") {
			res = append(res, child)
		} else {
			res = append(res, collectInnerEntityNodes(child)...)
		}
	}

	return res
}

func displayNodesTopology(roots []*entityNode, docLocation *jsonpointer.JSONPointer, entities []string, cmdConfig common2.ToolConfig) {
	logger := log.GetLogger("")

	fd := int(os.Stdout.Fd())
	termWidth, _, err := term.GetSize(fd)
	if err != nil {
		// Fallback to standard 80 columns if an error occurs (e.g., piped output)
		logger.Debug("Failed to get terminal size, disabling the text truncation", "error", err)
		termWidth = 0
	}
	logger.Debug("Terminal size", "width", termWidth)

	var renderTreeNodes []*renderTreeNode
	for _, root := range roots {
		// Show a root if none or both flags are set, otherwise check its location in the document
		show := cmdConfig.Doc.Inspect.Main == cmdConfig.Doc.Inspect.Components || isNodeIsMainOrComponent(root, cmdConfig)
		if show {
			logger.Trace("Building render tree", "node", root.node)
			renderTree := buildRenderTree(root, entities, cmdConfig)
			if renderTree == nil {
				logger.Trace("Skipping node due to filtering by entity kind", "node", root.node, "entity", root.entity, "entities", entities)
				continue
			}

			if cmdConfig.Doc.Inspect.List {
				renderTreeNodes = append(renderTreeNodes, common2.FlattenTree[*renderTreeNode](renderTree)...)
			} else {
				displayRenderTree(docLocation, renderTree, nil, termWidth, cmdConfig)
			}
		} else {
			logger.Trace("Skipping node due to filtering by location", "node", root.node, "main", cmdConfig.Doc.Inspect.Main, "components", cmdConfig.Doc.Inspect.Components)
		}
	}

	if cmdConfig.Doc.Inspect.List {
		nodes := lo.UniqBy(renderTreeNodes, func(n *renderTreeNode) string {
			return n.entityNode.node.AbsPointerString()
		})
		logger.Trace("Showing nodes as list", "nodesCount", len(nodes), "entryStyle", cmdConfig.Doc.Inspect.EntryStyle)

		for _, n := range nodes {
			entity := n.entityNode
			// Show a node if none or both flags are set, otherwise check its location in the document
			show := cmdConfig.Doc.Inspect.Main == cmdConfig.Doc.Inspect.Components || isNodeIsMainOrComponent(entity, cmdConfig)
			if !show || n.dummy {
				logger.Trace("Skipping node due to filtering", "node", entity.node, "entities", entities, "main", cmdConfig.Doc.Inspect.Main, "components", cmdConfig.Doc.Inspect.Components)
				continue
			}
			relPath := getRelativePath(docLocation, entity.node.AbsOriginDocumentPath(), false)
			fmt.Println(formatRenderTreeNode(relPath, entity, cmdConfig, false, false))
		}
	}
}

type entityNode struct {
	node     *types.RawNode
	refTo    *entityNode
	ref      *jsonpointer.JSONPointer
	children []*entityNode
	cycle    bool
	entity   string
}

type renderTreeNode struct {
	unresolvedEntityNode *entityNode
	entityNode           *entityNode
	refHops              int
	children             []*renderTreeNode
	dummy                bool
}

func (r renderTreeNode) Children() []*renderTreeNode {
	return r.children
}

func buildRenderTree(root *entityNode, entities []string, cmdConfig common2.ToolConfig) *renderTreeNode {
	res := buildRenderTree2(nil, root, entities, cmdConfig)
	if len(res) > 1 {
		panic(fmt.Errorf("expected exactly zero or one render tree root to be build, got %d; this is a bug", len(res)))
	}
	v, _ := lo.First(res)
	return v
}

func buildRenderTree2(parent *renderTreeNode, node *entityNode, entities []string, cmdConfig common2.ToolConfig) []*renderTreeNode {
	var dummy bool
	if len(entities) > 0 {
		match := lo.Contains(entities, node.entity)
		// Special case for `-e other`, which means "non-entity nodes".
		nonEntityNodeMatch := !strings.HasPrefix(node.entity, ">") && lo.Contains(entities, "")
		dummy = !match && !nonEntityNodeMatch
	}

	targetNode := node
	n := &renderTreeNode{unresolvedEntityNode: node, entityNode: targetNode, dummy: dummy}
	visible := true

	// Collapse the possible $ref chains
	var hops int
	for ; targetNode.refTo != nil; hops++ {
		targetNode = targetNode.refTo
	}

	switch {
	case cmdConfig.Doc.Inspect.RecursiveAll:
		// Show every $ref hop
		if hops > 1 {
			n.refHops = 1
			n.entityNode = node.refTo
			n.children = append(n.children, buildRenderTree2(n, node.refTo, entities, cmdConfig)...)
			if len(n.children) == 0 && dummy {
				// Eliminate dummy nodes if they don't have non-dummy children on any level deep.
				return nil
			}
			return []*renderTreeNode{n}
		}
	case parent == nil:
	case cmdConfig.Doc.Inspect.Recursive:
		// In recursive mode, show only the $refs hierarchy inside the JSON Schema objects.
		if parent.unresolvedEntityNode.entity == ">schema" && node.entity == ">schema" {
			visible = node.ref != nil
		}
	case node.ref != nil:
		// Don't show the $refs in non-recursive mode
		return nil
	}
	n.entityNode = targetNode
	n.refHops = hops

	if targetNode.node.Kind() == types.RawNodeKindScalar {
		if dummy {
			// Eliminate dummy scalar nodes
			return nil
		}
		return lo.Ternary(visible, []*renderTreeNode{n}, nil)
	}

	// Compress the entity tree to contain only the visible nodes
	visibleParent := lo.Ternary(visible, n, parent)
	var children []*renderTreeNode
	for _, child := range targetNode.children {
		ch := buildRenderTree2(visibleParent, child, entities, cmdConfig)
		children = append(children, ch...)
	}
	if len(children) == 0 && dummy {
		// Eliminate dummy nodes if they don't have non-dummy children on any level deep.
		return nil
	}
	if !visible {
		// Substitute the node with its children if the node is not visible
		return children
	}

	n.children = children
	return []*renderTreeNode{n}
}

func isNodeIsMainOrComponent(node *entityNode, cmdConfig common2.ToolConfig) bool {
	rootSections, componentsSections := asyncapiEntitiesSectionPaths()

	var parentPath []string
	if len(node.node.Path()) > 0 {
		parentPath = node.node.Path()[:len(node.node.Path())-1]
	}

	isDefinition := lo.ContainsBy(rootSections, func(s string) bool { return slices.Equal([]string{s}, parentPath) })
	isComponent := lo.ContainsBy(componentsSections, func(s string) bool { return slices.Equal([]string{"components", s}, parentPath) })

	switch {
	case cmdConfig.Doc.Inspect.Main && cmdConfig.Doc.Inspect.Components:
		return isDefinition || isComponent
	case cmdConfig.Doc.Inspect.Main:
		return isDefinition
	case cmdConfig.Doc.Inspect.Components:
		return isComponent
	}
	return true
}

func displayRenderTree(mainDoc *jsonpointer.JSONPointer, node *renderTreeNode, tails []bool, termWidth int, cmdConfig common2.ToolConfig) {
	logger := log.GetLogger("")
	logger.Trace("Render node", "entityNode", node.entityNode.node, "unresolvedEntityNode", node.unresolvedEntityNode.node, "tails", tails)

	fmt.Println(renderTreeLine(mainDoc, node, tails, termWidth, cmdConfig))

	for i, child := range node.children {
		displayRenderTree(mainDoc, child, append(slices.Clone(tails), i == len(node.children)-1), termWidth, cmdConfig)
	}
}

func renderTreeLine(mainDoc *jsonpointer.JSONPointer, node *renderTreeNode, tails []bool, termWidth int, cmdConfig common2.ToolConfig) string {
	truncateLine := func(line string, termWidth int, ansiOutput bool) string {
		// Truncate the line to fit the terminal width, if necessary
		if termWidth > 0 {
			if l, ok := utils.TruncateANSIString(line, termWidth-1); ok {
				line = l + "…"
			}
		}
		if ansiOutput {
			line += "\033[0m" // Reset color to avoid color bleeding in the terminal
		}
		return line
	}

	if len(tails) > 0 {
		for _, isTail := range tails[:len(tails)-1] {
			fmt.Print(lo.Ternary(isTail, "    ", "│   "))
		}
		fmt.Print(lo.Ternary(tails[len(tails)-1], "└── ", "├── "))
	}
	contentWidth := termWidth - len(tails)*4
	colorfulOutput := cmdConfig.Doc.Inspect.EntryStyle == common2.DocInspectPathStyleHuman

	var contents strings.Builder
	if node.dummy {
		contents.WriteString("X ")
	}
	if node.unresolvedEntityNode.ref != nil {
		contents.WriteString(formatRenderTreeNode(
			getRelativePath(mainDoc, node.unresolvedEntityNode.node.AbsOriginDocumentPath(), false),
			node.unresolvedEntityNode,
			cmdConfig,
			true,
			node.dummy,
		))
		contents.WriteString(" ")

		var tags []string
		if node.refHops > 1 {
			// More than one $refs between node and $ref's target, indicate that in the output for readability
			tags = append(tags, fmt.Sprintf("%d more refs", node.refHops-1))
		}
		// Mark the node "cycle" if there is a cycle detected between $refs
		if node.entityNode.cycle {
			tags = append(tags, "cycle")
		}
		if len(tags) > 0 {
			contents.WriteString("-(" + strings.Join(tags, ", ") + ")")
		}

		contents.WriteString("-> ")
		if node.refHops == 0 {
			// $ref was not resolved into target node, so just print the $ref value
			contents.WriteString("'" + node.unresolvedEntityNode.ref.String() + "'")
			return truncateLine(contents.String(), contentWidth, colorfulOutput)
		}
	}

	contents.WriteString(formatRenderTreeNode(
		getRelativePath(mainDoc, node.entityNode.node.AbsOriginDocumentPath(), false),
		node.entityNode,
		cmdConfig,
		false,
		node.dummy,
	))

	return truncateLine(contents.String(), contentWidth, colorfulOutput)
}

func formatRenderTreeNode(docLocation string, node *entityNode, cmdConfig common2.ToolConfig, isRef, dimmed bool) string {
	switch cmdConfig.Doc.Inspect.EntryStyle {
	case common2.DocInspectPathStyleYq:
		return formatRenderTreeYqEntryStyle(node.node.RawPath(), docLocation)
	case common2.DocInspectPathStyleJSONPointer:
		if docLocation != "" {
			p := lo.Must(jsonpointer.Parse(docLocation))
			return p.Join(node.node.Path()...).String()
		}
		return jsonpointer.PointerString(node.node.Path()...)
	case common2.DocInspectPathStyleHuman, common2.DocInspectPathStyleHumanNoColor:
		return formatRenderTreeHumanEntryStyle(node, docLocation, cmdConfig.Doc.Inspect.EntryStyle == common2.DocInspectPathStyleHuman, isRef, dimmed)
	default:
		panic(fmt.Errorf("unknown path format %q", cmdConfig.Doc.Inspect.EntryStyle))
	}
}

// formatRenderTreeYqEntryStyle renders a node path in the yq-compatible expression format, e.g. ".channels.foo" or
// ".channels.foo.messages[0]". If docLocation is not empty, the node resides in a different document than the main one,
// which a yq path expression cannot address, so the location is appended as a trailing yq comment for readability.
func formatRenderTreeYqEntryStyle(p []any, docLocation string) string {
	var b strings.Builder
	if len(p) == 0 {
		b.WriteString(".")
	}
	for _, seg := range p {
		switch v := seg.(type) {
		case int:
			b.WriteString("[" + strconv.Itoa(v) + "]")
		case string:
			isSimpleKey := true
			for i, r := range v {
				isSimpleKey = r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || (i > 0 && r >= '0' && r <= '9')
				if !isSimpleKey {
					break
				}
			}
			if isSimpleKey {
				b.WriteString("." + v)
			} else {
				b.WriteString(`.["` + v + `"]`)
			}
		default:
			panic(fmt.Errorf("unexpected path segment type %T, this is a bug", seg))
		}
	}
	if docLocation != "" {
		// yq path expressions can't reference another document, so print its location as a trailing comment
		b.WriteString(" #FILE:" + docLocation)
	}
	return b.String()
}

// formatRenderTreeHumanEntryStyle renders an entity in a human-readable format.
func formatRenderTreeHumanEntryStyle(node *entityNode, docLocation string, color, isRef, dimmed bool) string {
	var (
		consoleReset = lo.Ternary(color, "\033[0m", "")

		consoleGray         = lo.Ternary(color, "\033[90m", "")
		consoleYellow       = lo.Ternary(color, "\033[33m", "")
		consoleBrightWhite  = lo.Ternary(color, "\033[97m", "")
		consoleBlue         = lo.Ternary(color, "\033[34m", "")
		consoleCyan         = lo.Ternary(color, "\033[36m", "")
		consoleBrightYellow = lo.Ternary(color, "\033[93m", "")
		consoleDimmed       = lo.Ternary(color, "\033[2m", "")
		consoleDimmedYellow = lo.Ternary(color, "\033[2;33m", "")
	)

	if dimmed {
		consoleYellow = consoleDimmedYellow
	}

	var b strings.Builder
	getProps := func(n *entityNode, addName bool, names ...string) string {
		for _, name := range names {
			if v, ok := n.node.Get(name); ok && v.AsStringSafe() != "" {
				s := strings.Trim(strconv.QuoteToGraphic(v.AsStringSafe()), "\"")
				if addName {
					s = fmt.Sprintf("%s:%s", name, s)
				}
				return s
			}
		}
		return ""
	}
	getKey := func(n *entityNode) string {
		last, ok := lo.Last(n.node.RawPath())
		if !ok {
			return ""
		}
		v, ok := last.(string)
		if !ok {
			return ""
		}
		return strings.Trim(strconv.QuoteToGraphic(v), "\"")
	}

	var id, description string
	var tags []string

	if len(node.node.Path()) == 3 && node.node.Path()[0] == "components" {
		tags = append(tags, "component")
	}
	switch node.entity {
	case ">contact":
		description = getProps(node, false, "name", "email")
	case ">info":
		description = getProps(node, false, "title", "description")
	case ">license":
		description = getProps(node, false, "name", "url")
	case ">server":
		id = getKey(node)
		tags = append(tags,
			lo.CoalesceOrEmpty(getProps(node, true, "protocol"), "?protocol"),
			lo.CoalesceOrEmpty(getProps(node, true, "host"), "?host"),
		)
		description = getProps(node, false, "title", "summary", "description")
	case ">serverVariable":
		id = getKey(node)
		description = getProps(node, false, "description")
	case ">channel":
		id = getKey(node)
		tags = append(tags, lo.CoalesceOrEmpty(getProps(node, true, "address"), "?address"))
		description = getProps(node, false, "title", "summary", "description")
	case ">operation":
		id = getKey(node)
		tags = append(tags, lo.CoalesceOrEmpty(getProps(node, true, "action"), "?action"))
		description = getProps(node, false, "title", "summary", "description")
	case ">operationTrait":
		description = getProps(node, false, "title", "summary", "description")
	case ">operationReply":
	case ">operationReplyAddress":
		tags = append(tags, lo.CoalesceOrEmpty(getProps(node, true, "location"), "?location"))
		description = getProps(node, false, "description")
	case ">parameter":
		id = getKey(node)
		description = getProps(node, false, "description")
	case ">serverBindings", ">channelBindings", ">operationBindings", ">messageBindings":
		for k := range node.node.Entries() {
			tags = append(tags, fmt.Sprintf("%v", k))
		}
	case ">message", ">messageTrait":
		id = getKey(node)
		description = getProps(node, false, "title", "summary", "description", "name")
	case ">tag":
		description = getProps(node, false, "name", "description")
	case ">externalDocumentation":
		description = getProps(node, false, "description", "url")
	case ">securityScheme":
		id = lo.CoalesceOrEmpty(getProps(node, true, "name"), getKey(node), "?name")
		tags = append(tags,
			lo.CoalesceOrEmpty(getProps(node, true, "type"), "?type"),
			lo.CoalesceOrEmpty(getProps(node, true, "in"), "?in"),
		)
		description = getProps(node, false, "description")
	case ">schema":
		id = getKey(node)
		description = getProps(node, false, "title", "description")
	}

	if dimmed {
		b.WriteString(consoleDimmed)
	}
	nodeEntity := lo.Ternary(
		strings.HasPrefix(node.entity, ">"),
		lo.PascalCase(strings.TrimPrefix(node.entity, ">")),
		fmt.Sprintf("?%sNode", lo.Capitalize(string(node.node.Kind()))),
	)
	b.WriteString(consoleBrightYellow + nodeEntity)
	if isRef {
		b.WriteString(consoleBlue + "$ref")
	}
	b.WriteString(consoleReset + consoleBrightWhite + "(")
	if id == "" {
		b.WriteString("'" + jsonpointer.PointerString(node.node.Path()...) + "'")
	} else {
		b.WriteString(id)
	}
	if docLocation != "" {
		b.WriteString(consoleCyan + "@" + docLocation + consoleBrightWhite)
	}
	b.WriteString(")" + consoleReset)
	if !isRef {
		if len(tags) > 0 {
			for _, t := range tags {
				b.WriteString(" " + consoleYellow + "[" + t + "]" + consoleReset)
			}
		}
		if description != "" {
			b.WriteString(consoleGray + " # " + description)
		}
	}
	b.WriteString(consoleReset)

	return b.String()
}

func getRelativePath(absMainDoc, doc *jsonpointer.JSONPointer, printSelf bool) string {
	if doc.Location() == absMainDoc.Location() {
		if !printSelf {
			return ""
		}
		if doc.FSPath != "" {
			return path.Base(doc.FSPath)
		}
		return doc.Location()
	}
	if doc.FSPath != "" && absMainDoc.FSPath != "" {
		return lo.Must(filepath.Rel(filepath.Dir(absMainDoc.FSPath), doc.FSPath))
	}
	return doc.Location()
}
