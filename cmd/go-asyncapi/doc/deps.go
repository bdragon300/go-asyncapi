package doc

import (
	"fmt"
	"slices"
	"time"

	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/samber/lo"
)

type DepsCmd struct {
	Document string `arg:"positional,required" help:"AsyncAPI document file or url" placeholder:"DOCUMENT"`

	Tree            bool `arg:"--tree,-t" help:"Show the result as a tree"`
	AllowRemoteRefs bool `arg:"--allow-remote-refs,-F" help:"Follow the $refs pointing to URLs"`

	LocatorRootDir string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`
}

func cliDeps(cmd *DepsCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")
	logger.Info("Hint: Use --quiet to suppress the logging output")

	docLocation, err := jsonpointer.Parse(cmd.Document)
	if err != nil {
		return fmt.Errorf("parse part or url %s: %w", cmd.Document, err)
	}

	logger.Debug("Loading document", "path", docLocation.Location())
	locator := common2.GetLocator(cmdConfig)
	inputContents, err := loadDocument(docLocation, locator)
	if err != nil {
		return fmt.Errorf("load document %s: %w", docLocation.Location(), err)
	}

	logger.Debug("Inspecting node", "path", inputContents.AbsPointerString())
	docs := map[string]*documentTree{absLocation(inputContents.AbsOriginDocumentPath()): inputContents}
	inspectedRoot, err := inspectNode(inputContents.RawNode, nil, docs, 0, common2.GetLocator(cmdConfig), true, cmdConfig.Doc.Deps.AllowRemoteReferences)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", docLocation.Location(), err)
	}

	logger.Trace("Building render doc tree", "location", docLocation.Location())
	renderDocTree := &renderDocTreeNode{absLocation: inputContents.AbsOriginDocumentPath()}
	buildRenderDocTree(renderDocTree, nil, inspectedRoot)

	if cmdConfig.Doc.Deps.Tree {
		logger.Trace("Rendering document topology as tree", "location", docLocation.Location())
		displayRenderDocTree(renderDocTree, inputContents.AbsOriginDocumentPath(), nil)
		return nil
	}

	renderNodes := common2.FlattenTree[*renderDocTreeNode](renderDocTree)
	logger.Trace("Rendering document topology as list", "location", docLocation.Location(), "nodes", len(renderNodes))
	for _, n := range renderNodes {
		fmt.Println(getRelativePath(inputContents.AbsOriginDocumentPath(), n.absLocation, true))
	}

	return nil
}

type renderDocTreeNode struct {
	absLocation *jsonpointer.JSONPointer
	children    []*renderDocTreeNode
	cycle       bool
}

func (r renderDocTreeNode) Children() []*renderDocTreeNode {
	return r.children
}

func buildRenderDocTree(renderParent *renderDocTreeNode, visited []string, node *entityNode) {
	getDocument := func(p *renderDocTreeNode, n *entityNode) *renderDocTreeNode {
		doc, ok := lo.Find(p.children, func(child *renderDocTreeNode) bool {
			return child.absLocation.Location() == n.node.AbsOriginDocumentPath().Location()
		})
		if !ok {
			doc = &renderDocTreeNode{absLocation: n.node.AbsOriginDocumentPath()}
			p.children = append(p.children, doc)
		}
		return doc
	}

	p := node.node.AbsOriginDocumentPath()
	visited = append(slices.Clone(visited), p.Location())

	// Unroll the $ref chain
	for node.refTo != nil {
		node = node.refTo
		if renderParent.absLocation.Location() != node.node.AbsOriginDocumentPath().Location() {
			doc := getDocument(renderParent, node)
			if slices.Contains(visited, node.node.AbsOriginDocumentPath().Location()) {
				// Circular reference detected
				doc.cycle = true
				return
			}
			buildRenderDocTree(doc, visited, node)
			return
		}
	}

	for _, child := range node.children {
		buildRenderDocTree(renderParent, visited, child)
	}
}

func displayRenderDocTree(node *renderDocTreeNode, absMainDoc *jsonpointer.JSONPointer, exhausted []bool) {
	logger := log.GetLogger("")
	logger.Trace("Render node", "location", node.absLocation.Location(), "exhausted", exhausted)

	if len(exhausted) > 0 {
		for _, lastNode := range exhausted[:len(exhausted)-1] {
			fmt.Print(lo.Ternary(lastNode, "    ", "│   "))
		}
		fmt.Print(lo.Ternary(exhausted[len(exhausted)-1], "└── ", "├── "))
	}
	fmt.Print(getRelativePath(absMainDoc, node.absLocation, true))
	if node.cycle {
		fmt.Print(" (cycle)")
	}
	fmt.Println()

	for i, child := range node.children {
		displayRenderDocTree(child, absMainDoc, append(slices.Clone(exhausted), i == len(node.children)-1))
	}
}
