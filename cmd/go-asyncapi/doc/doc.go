package doc

import (
	"fmt"
	"path/filepath"

	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type anyEncoder interface {
	Encode(v any) error
}

type Cmd struct {
	Merge   *MergeCmd   `arg:"subcommand:merge" help:"Merge multiple AsyncAPI documents into one."`
	Unmerge *UnmergeCmd `arg:"subcommand:unmerge" help:"Move objects from one AsyncAPI document to another."`
	Indent  int         `arg:"--indent" help:"Output document indentation width" placeholder:"SPACES"`
	Format  string      `arg:"--format,-f" help:"Output format. Possible values: yaml, json" placeholder:"FORMAT"`
}

func newDocumentTree(originDocument *jsonpointer.JSONPointer) *documentTree {
	return &documentTree{RawNode: *types.NewEmptyRawNode(originDocument)}
}

type documentTree struct {
	types.RawNode
}

// CollectRefs collects pointers to all object RawNodes in the document that have a "$ref" key.
func (d documentTree) CollectRefs() []*types.RawNode {
	rootKeys, componentsKeys := d.MergeableMapPaths()
	res := lo.FlatMap(rootKeys, func(key string, _ int) []*types.RawNode {
		node, _ := d.Get(key)
		return collectRefs(node)
	})

	if comp, ok := d.Get("components"); ok {
		res = append(res, lo.FlatMap(componentsKeys, func(key string, _ int) []*types.RawNode {
			node, _ := comp.Get(key)
			return collectRefs(node)
		})...)
	}

	return res
}

func (d documentTree) MergeableMaps() []*types.RawNode {
	rootKeys, componentsKeys := d.MergeableMapPaths()
	res := lo.Map(rootKeys, func(key string, _ int) *types.RawNode {
		node, _ := d.Get(key)
		return node
	})
	if comp, ok := d.Get("components"); ok {
		res = append(res, lo.Map(componentsKeys, func(key string, _ int) *types.RawNode {
			node, _ := comp.Get(key)
			return node
		})...)
	}

	return res
}

func (d documentTree) MergeableMapPaths() ([]string, []string) {
	rootKeys := []string{"servers", "channels", "operations"}
	componentsKeys := []string{
		"schemas",
		"servers", "channels", "operations", "messages",
		"securitySchemes", "serverVariables", "parameters", "correlationIds", "replies", "replyAddresses", "externalDocs", "tags",
		"operationTraits", "messageTraits",
		"serverBindings", "channelBindings", "operationBindings", "messageBindings",
	}
	return rootKeys, componentsKeys
}

func (d *documentTree) Root() *types.RawNode {
	return &d.RawNode
}

type changeLogEntry struct {
	Source, Destination *jsonpointer.JSONPointer
	Move                bool
}

func CliDoc(cmd *Cmd, globalConfig common2.ToolConfig) error {
	cmdConfig, err := cliConfig(globalConfig, cmd)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	switch {
	case cmd.Merge != nil:
		return cliMerge(cmd.Merge, cmdConfig)
	case cmd.Unmerge != nil:
		return cliUnmerge(cmd.Unmerge, cmdConfig)
	}
	return fmt.Errorf("%w: unknown doc subcommand", common2.ErrWrongCliArgs)
}

func cliConfig(globalConfig common2.ToolConfig, cmd *Cmd) (common2.ToolConfig, error) {
	res := globalConfig

	res.Doc.Indent = common2.Coalesce(cmd.Indent, globalConfig.Doc.Indent)
	res.Doc.Format = common2.Coalesce(cmd.Format, globalConfig.Doc.Format)
	res.Doc.Merge.OutputFile = common2.Coalesce(cmd.Merge.Output, globalConfig.Doc.Merge.OutputFile)
	res.Doc.Merge.Strategy = common2.Coalesce(cmd.Merge.Strategy, globalConfig.Doc.Merge.Strategy)
	res.Doc.Merge.DisableRewriting = common2.Coalesce(cmd.Merge.DisableRewriting, globalConfig.Doc.Merge.DisableRewriting)
	res.Doc.Unmerge.OutputFile = common2.Coalesce(cmd.Unmerge.Output, globalConfig.Doc.Unmerge.OutputFile)
	res.Doc.Unmerge.Scope = common2.Coalesce(cmd.Unmerge.Scope, globalConfig.Doc.Unmerge.Scope)
	res.Doc.Unmerge.Duplicate = common2.Coalesce(cmd.Unmerge.Duplicate, globalConfig.Doc.Unmerge.Duplicate)
	res.Doc.Unmerge.DisableRewriting = common2.Coalesce(cmd.Unmerge.DisableRewriting, globalConfig.Doc.Unmerge.DisableRewriting)

	return res, nil
}

// collectRefs recursively collects all n's children RawNodes (including n itself) that have a "$ref" key.
func collectRefs(n *types.RawNode) []*types.RawNode {
	if n == nil {
		return nil
	}
	if n.Kind() == types.RawNodeKindScalar {
		return nil
	}

	var res []*types.RawNode
	for _, v := range n.Entries() {
		res = append(res, collectRefs(v)...)
	}
	if n.Kind() == types.RawNodeKindObject && n.Has("$ref") {
		res = append(res, n)
	}

	return res
}

func parseRefRawNode(n *types.RawNode) (*jsonpointer.JSONPointer, error) {
	variant, _ := n.Get("$ref")
	if variant.Kind() != types.RawNodeKindScalar {
		return nil, fmt.Errorf("expected scalar node for $ref value, got %s", variant.Kind())
	}
	s, ok := variant.AsScalar().(string)
	if !ok {
		return nil, fmt.Errorf("expected string value for $ref, got %T", variant.AsScalar())
	}

	ref, err := jsonpointer.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("parse $ref value: %w", err)
	}
	return ref, nil
}

// absLocation returns the absolute path to document if p is a file path. Otherwise, if p is an URL, it returns the URL as is.
func absLocation(p *jsonpointer.JSONPointer) string {
	if p.FSPath != "" {
		return lo.Must(filepath.Abs(p.FSPath))
	}
	return p.Location()
}
