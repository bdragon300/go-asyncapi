package doc

import (
	"fmt"
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
	"github.com/samber/lo"
)

const maxDescriptionLengthInHumanPath = 80

type NodesCmd struct {
	Location string `arg:"positional,required" help:"Document with optional node name or globbing pattern. Format: file.{yaml|yml|json}[#/path/to/node | GLOBBING_PATTERN]" placeholder:"LOCATION"`

	Entities           string `arg:"--entities,-e" help:"Comma-separated list of entities to show or 'help' to list all available entities and exit'" placeholder:"ENTITIES"`
	TopLevel           bool   `arg:"--top-level" help:"Show only servers, channels and operations defined in the top-level sections of the documents"`
	Components         bool   `arg:"--components" help:"Show only entities defined in components section of the documents"`
	Expand             bool   `arg:"--expand,-x" help:"Expand nodes and resolve $refs."`
	ExpandAll          bool   `arg:"--expand-all,-X" help:"Expand all nodes. The same as -x, but also shows all nested jsonschema objects and fully unfolds all $refs"`
	Tree               bool   `arg:"--tree,-t" help:"Show the result as a tree"`
	FollowExternalRefs bool   `arg:"--follow-external-refs,-f" help:"Follow the $refs pointing to other documents"`
	AllowRemoteRefs    bool   `arg:"--allow-remote-refs,-F" help:"Follow the $refs pointing to URLs. Implies --follow-external-refs."`

	EntryStyle common2.DocNodesPathStyle `arg:"--entry-style" help:"Style of the output entries. Options: human, human-no-color, json-pointer, yq" placeholder:"STYLE"`

	LocatorRootDir string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`
}

func cliNodes(cmd *NodesCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	switch cmdConfig.Doc.Nodes.EntryStyle {
	case common2.DocNodesPathStyleJSONPointer, common2.DocNodesPathStyleYq, common2.DocNodesPathStyleHuman, common2.DocNodesPathStyleHumanNoColor:
	default:
		return fmt.Errorf("%w: unknown path format %q", common2.ErrInvalidCLIArgument, cmdConfig.Doc.Nodes.EntryStyle)
	}

	if strings.EqualFold(cmdConfig.Doc.Nodes.Entities, "help") {
		fmt.Println("Available entity types: " + strings.Join(asyncapiEntities(), ", ") + ", other")
		return nil
	}

	logger.Info("Hint: Use --quiet to suppress the logging output")

	// Parse --entities
	entities, err := parseEntityFilterExpression(cmdConfig.Doc.Nodes.Entities)
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

	logger.Debug("Collecting root nodes", "pattern", pattern)
	docs := map[string]*documentTree{absLocation(inputContents.AbsOriginDocumentPath()): inputContents}
	inputNodes := preselectInspectedNodes(inputContents, pattern)
	if len(inputNodes) == 0 {
		logger.Warn("Nothing to inspect", "pattern", pattern)
		return nil
	}

	var renderNodes []*entityNode
	for _, n := range inputNodes {
		logger.Debug("Inspecting node", "path", n)
		inspected, err := inspectNode(
			n,
			nil,
			docs,
			0,
			common2.GetLocator(cmdConfig),
			cmdConfig.Doc.Nodes.FollowExternalRefs || cmdConfig.Doc.Nodes.AllowRemoteReferences,
			cmdConfig.Doc.Nodes.AllowRemoteReferences,
		)
		if err != nil {
			return fmt.Errorf("inspect %s: %w", pattern.Location(), err)
		}
		renderNodes = append(renderNodes, inspected)
	}
	renderNodes = lo.UniqBy(renderNodes, func(n *entityNode) string {
		return n.node.AbsPointerString()
	})
	logger.Trace("Rendering nodes topology", "nodes", len(renderNodes), "tree", cmdConfig.Doc.Nodes.Tree, "expand", cmdConfig.Doc.Nodes.Expand, "expandAll", cmdConfig.Doc.Nodes.ExpandAll)
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

func preselectInspectedNodes(inputContents *documentTree, pattern cliPattern) []*types.RawNode {
	logger := log.GetLogger("")

	if len(pattern.Pointer) > 0 {
		logger.Trace("Collecting nodes by pattern", "pattern", pattern)
		return findNodesByPattern(inputContents.RawNode, pattern)
	}

	logger.Trace("Collecting nodes from root and components sections")
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

func inspectNode(node *types.RawNode, visited []*types.RawNode, docs map[string]*documentTree, level int, locator common2.DocumentLocator, external, remote bool) (*entityNode, error) {
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

func displayNodesTopology(renderNodes []*entityNode, docLocation *jsonpointer.JSONPointer, entities []string, cmdConfig common2.ToolConfig) {
	logger := log.GetLogger("")

	for i := 0; i < len(renderNodes); i++ {
		logger.Trace("Building render tree", "node", renderNodes[i].node)
		renderTree := buildRenderTree(nil, renderNodes[i], cmdConfig)
		showNode := isNodeVisibleInNodesTopology(renderNodes[i], entities, cmdConfig)
		allRenderNodes := common2.FlattenTree[*renderTreeNode](renderTree)
		logger.Trace("Render tree built", "node", renderTree.unresolvedEntityNode.node, "nodes", len(allRenderNodes), "showNode", showNode)

		if showNode {
			if cmdConfig.Doc.Nodes.Tree {
				displayRenderTree(docLocation, renderTree, nil, cmdConfig)
			} else {
				relPath := getRelativePath(docLocation, renderNodes[i].node.AbsOriginDocumentPath(), false)
				fmt.Println(formatRenderTreeNode(relPath, renderNodes[i], cmdConfig))
			}
		}

		l := len(renderNodes)
		if cmdConfig.Doc.Nodes.Expand || cmdConfig.Doc.Nodes.ExpandAll {
			if cmdConfig.Doc.Nodes.Tree {
				if !cmdConfig.Doc.Nodes.ExpandAll {
					// Do not extend the output in deep recursive mode, because the tree is already expanded and all nodes are printed
					renderNodes = append(renderNodes, lo.FilterMap(allRenderNodes, func(n *renderTreeNode, _ int) (*entityNode, bool) {
						return n.entityNode, n.visible && n.refHops > 0
					})...)
				}
			} else {
				renderNodes = append(renderNodes, lo.FilterMap(allRenderNodes, func(n *renderTreeNode, _ int) (*entityNode, bool) {
					return n.entityNode, cmdConfig.Doc.Nodes.ExpandAll || n.visible
				})...)
			}
			renderNodes = lo.UniqBy(renderNodes, func(n *entityNode) string {
				return n.node.AbsPointerString()
			})
			logger.Trace("Extended rendering nodes slice", "grow", len(renderNodes)-l)
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
	visible              bool
	refHops              int
	children             []*renderTreeNode
}

func (r renderTreeNode) Children() []*renderTreeNode {
	return r.children
}

func buildRenderTree(parent *renderTreeNode, node *entityNode, cmdConfig common2.ToolConfig) *renderTreeNode {
	// Unroll the node to final node through $ref chain, if any.
	var hops int
	targetNode := node
	for ; targetNode.refTo != nil; hops++ {
		targetNode = targetNode.refTo
	}

	res := &renderTreeNode{
		unresolvedEntityNode: node,
		entityNode:           targetNode,
		refHops:              hops,
		visible:              parent == nil || parent.visible,
	}
	if res.visible && parent != nil {
		switch {
		case cmdConfig.Doc.Nodes.ExpandAll:
		case parent.refHops > 0:
			// Don't expand $refs
			res.visible = false
		case strings.HasPrefix(node.entity, ">") && parent.entityNode.entity == node.entity && node.ref == nil:
			// Hide entity that is the same as the parent entity. In particular, this prevents printing subschemas of a schema.
			res.visible = false
		}
	}
	if targetNode.node.Kind() == types.RawNodeKindScalar {
		return res
	}

	for _, child := range targetNode.children {
		ch := buildRenderTree(res, child, cmdConfig)
		res.children = append(res.children, ch)
	}

	return res
}

func isNodeVisibleInNodesTopology(node *entityNode, entities []string, cmdConfig common2.ToolConfig) bool {
	if len(entities) > 0 {
		match := lo.Contains(entities, node.entity)
		nonEntityNodeMatch := !strings.HasPrefix(node.entity, ">") && lo.Contains(entities, "")
		if !match && !nonEntityNodeMatch {
			return false
		}
	}

	if !cmdConfig.Doc.Nodes.TopLevel && !cmdConfig.Doc.Nodes.Components {
		return true
	}
	rootSections, componentsSections := asyncapiEntitiesSectionPaths()

	var parentPath []string
	if len(node.node.Path()) > 0 {
		parentPath = node.node.Path()[:len(node.node.Path())-1]
	}

	isDefinition := lo.ContainsBy(rootSections, func(s string) bool { return slices.Equal([]string{s}, parentPath) })
	isComponent := lo.ContainsBy(componentsSections, func(s string) bool { return slices.Equal([]string{"components", s}, parentPath) })

	switch {
	case cmdConfig.Doc.Nodes.TopLevel && cmdConfig.Doc.Nodes.Components:
		return isDefinition || isComponent
	case cmdConfig.Doc.Nodes.TopLevel:
		return isDefinition
	case cmdConfig.Doc.Nodes.Components:
		return isComponent
	}
	return true
}

func displayRenderTree(mainDoc *jsonpointer.JSONPointer, node *renderTreeNode, exhausted []bool, cmdConfig common2.ToolConfig) {
	logger := log.GetLogger("")
	logger.Trace("Render node", "entityNode", node.entityNode.node, "unresolvedEntityNode", node.unresolvedEntityNode.node, "visible", node.visible, "exhausted", exhausted)

	if node.visible {
		displayRenderTreeLine(mainDoc, node, exhausted, cmdConfig)
		fmt.Println()
	}

	_, lastPrintableIdx, _ := lo.FindLastIndexOf(node.children, func(child *renderTreeNode) bool {
		return child.visible
	})
	for i, child := range node.children {
		displayRenderTree(mainDoc, child, append(slices.Clone(exhausted), i == lastPrintableIdx), cmdConfig)
	}
}

func displayRenderTreeLine(mainDoc *jsonpointer.JSONPointer, node *renderTreeNode, exhausted []bool, cmdConfig common2.ToolConfig) {
	if len(exhausted) > 0 {
		for _, lastNode := range exhausted[:len(exhausted)-1] {
			fmt.Print(lo.Ternary(lastNode, "    ", "│   "))
		}
		fmt.Print(lo.Ternary(exhausted[len(exhausted)-1], "└── ", "├── "))
	}

	if node.unresolvedEntityNode.ref != nil {
		// If $ref is located in a top-level node, e.g. in "#/channels", print this node as well for readability
		if len(exhausted) == 0 {
			relPath := getRelativePath(mainDoc, node.unresolvedEntityNode.node.AbsOriginDocumentPath(), false)
			fmt.Print(formatRenderTreeNode(relPath, node.unresolvedEntityNode, cmdConfig))
			fmt.Print(": ")
		}

		fmt.Print("$ref ")

		var tags []string
		if node.refHops > 1 {
			// More than one $refs between node and $ref's target, indicate that in the output for readability
			tags = append(tags, fmt.Sprintf("%d more refs", node.refHops-1))
		}
		if node.entityNode.cycle {
			tags = append(tags, "cycle")
		}
		if len(tags) > 0 {
			fmt.Print("-(" + strings.Join(tags, ", ") + ")")
		}

		fmt.Print("-> ")
		if node.refHops == 0 {
			// $ref was not resolved into target node, so just print the $ref value
			fmt.Print("'" + node.unresolvedEntityNode.ref.String() + "'")
			return
		}
	}

	fmt.Print(formatRenderTreeNode(getRelativePath(mainDoc, node.entityNode.node.AbsOriginDocumentPath(), false), node.entityNode, cmdConfig))
}

func formatRenderTreeNode(docLocation string, node *entityNode, cmdConfig common2.ToolConfig) string {
	switch cmdConfig.Doc.Nodes.EntryStyle {
	case common2.DocNodesPathStyleYq:
		return formatRenderTreeYqEntryStyle(node.node.RawPath(), docLocation)
	case common2.DocNodesPathStyleJSONPointer:
		if docLocation != "" {
			p := lo.Must(jsonpointer.Parse(docLocation))
			return p.Join(node.node.Path()...).String()
		}
		return jsonpointer.PointerString(node.node.Path()...)
	case common2.DocNodesPathStyleHuman, common2.DocNodesPathStyleHumanNoColor:
		return formatRenderTreeHumanEntryStyle(node, docLocation, cmdConfig.Doc.Nodes.EntryStyle == common2.DocNodesPathStyleHuman)
	default:
		panic(fmt.Errorf("unknown path format %q", cmdConfig.Doc.Nodes.EntryStyle))
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
func formatRenderTreeHumanEntryStyle(node *entityNode, docLocation string, color bool) string {
	var (
		consoleReset = lo.Ternary(color, "\033[0m", "")

		consoleGray        = lo.Ternary(color, "\033[90m", "")
		consoleYellow      = lo.Ternary(color, "\033[33m", "")
		consoleGreen       = lo.Ternary(color, "\033[32m", "")
		consoleBrightWhite = lo.Ternary(color, "\033[97m", "")
		consoleBlue        = lo.Ternary(color, "\033[34m", "")
	)

	var b strings.Builder
	getProps := func(n *entityNode, addName bool, names ...string) string {
		for _, name := range names {
			if v, ok := n.node.Get(name); ok && v.AsStringSafe() != "" {
				if addName {
					return fmt.Sprintf("%s:%s", name, v.AsStringSafe())
				}
				return v.AsStringSafe()
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
		return v
	}

	var id, description string
	var tags []string

	if len(node.node.Path()) == 3 && node.node.Path()[0] == "components" {
		tags = append(tags, "component")
	}
	switch node.entity {
	case ">contact":
		description = getProps(node, true, "name", "email")
	case ">info":
		description = getProps(node, true, "title", "description")
	case ">license":
		description = getProps(node, true, "name", "url")
	case ">server":
		id = getKey(node)
		tags = append(tags, fmt.Sprintf(
			"%s://%s",
			lo.CoalesceOrEmpty(getProps(node, false, "protocol"), "?protocol"),
			lo.CoalesceOrEmpty(getProps(node, false, "host"), "?host"),
		))
		description = getProps(node, true, "title", "summary", "description")
	case ">serverVariable":
		id = getKey(node)
		description = getProps(node, true, "description")
	case ">channel":
		id = getKey(node)
		tags = append(tags, lo.CoalesceOrEmpty(getProps(node, true, "address"), "?address"))
		description = getProps(node, true, "title", "summary", "description")
	case ">operation":
		id = getKey(node)
		tags = append(tags, lo.CoalesceOrEmpty(getProps(node, true, "action"), "?action"))
		description = getProps(node, true, "title", "summary", "description")
	case ">operationTrait":
		description = getProps(node, true, "title", "summary", "description")
	case ">operationReply":
	case ">operationReplyAddress":
		tags = append(tags, lo.CoalesceOrEmpty(getProps(node, true, "location"), "?location"))
		description = getProps(node, true, "description")
	case ">parameter":
		id = getKey(node)
		description = getProps(node, true, "description")
	case ">serverBindings", ">channelBindings", ">operationBindings", ">messageBindings":
		for k := range node.node.Entries() {
			tags = append(tags, fmt.Sprintf("%v", k))
		}
	case ">message", ">messageTrait":
		id = getKey(node)
		description = getProps(node, true, "title", "summary", "description", "name")
	case ">tag":
		description = getProps(node, true, "name", "description")
	case ">externalDocumentation":
		description = getProps(node, true, "description", "url")
	case ">securityScheme":
		id = lo.CoalesceOrEmpty(getProps(node, true, "name"), getKey(node), "?name")
		tags = append(tags,
			lo.CoalesceOrEmpty(getProps(node, false, "type"), "?type"),
			lo.CoalesceOrEmpty(getProps(node, false, "in"), "?in"),
		)
		description = getProps(node, true, "description")
	case ">schema":
		id = getKey(node)
		description = getProps(node, true, "title", "description")
	}

	nodeEntity := lo.Ternary(
		strings.HasPrefix(node.entity, ">"),
		lo.PascalCase(strings.TrimPrefix(node.entity, ">")),
		fmt.Sprintf("?%sNode", lo.Capitalize(string(node.node.Kind()))),
	)
	b.WriteString(consoleGreen + nodeEntity)
	if docLocation != "" {
		b.WriteString(consoleYellow + "@" + docLocation)
	}
	if id == "" {
		b.WriteString(consoleBrightWhite + " '" + jsonpointer.PointerString(node.node.Path()...) + "'")
	} else {
		b.WriteString(consoleBrightWhite + " " + strconv.Quote(id))
	}
	if len(tags) > 0 {
		b.WriteString(consoleBlue + " [" + strings.Join(tags, ",") + "]")
	}
	if description != "" {
		b.WriteString(consoleGray + " " + lo.Ellipsis(description, maxDescriptionLengthInHumanPath))
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
