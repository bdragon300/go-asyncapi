package doc

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

const (
	defaultStdoutWidth              = 80
	sideBySideColumnSeparator       = " | "
	sideBySideLineNumberPrefixWidth = 6
)

type Cmd struct {
	Merge      *MergeCmd `arg:"subcommand:merge" help:"Merge multiple AsyncAPI documents into one."`
	YAMLIndent int       `arg:"--yaml-indent" help:"Output YAML document indentation width" placeholder:"SPACES"`
}

type documentTree struct {
	types.RawNode
	DocumentURL jsonpointer.JSONPointer `yaml:"-" json:"-"`
}

// CollectRefs collects pointers to all object RawNodes in the document that have a "$ref" key.
func (d documentTree) CollectRefs() []*types.RawNode {
	rootKeys, componentsKeys := d.mergeableKeys()
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

func (d documentTree) MergeableNodes() []*types.RawNode {
	rootKeys, componentsKeys := d.mergeableKeys()
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

func (d documentTree) mergeableKeys() ([]string, []string) {
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

type changeLogEntry struct {
	Source, Destination *jsonpointer.JSONPointer
}

func CliDoc(cmd *Cmd, globalConfig common2.ToolConfig) error {
	cmdConfig, err := cliConfig(globalConfig, cmd)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if cmd.Merge != nil {
		return cliMerge(cmd.Merge, cmdConfig)
	}
	return fmt.Errorf("%w: unknown doc subcommand", common2.ErrWrongCliArgs)
}

func cliConfig(globalConfig common2.ToolConfig, cmd *Cmd) (common2.ToolConfig, error) {
	res := globalConfig

	res.Doc.YAMLIndent = common2.Coalesce(cmd.YAMLIndent, globalConfig.Doc.YAMLIndent)
	res.Doc.Merge.Strategy = common2.Coalesce(cmd.Merge.Strategy, globalConfig.Doc.Merge.Strategy)
	res.Doc.Merge.DisableRewriting = common2.Coalesce(cmd.Merge.DisableRewriting, globalConfig.Doc.Merge.DisableRewriting)

	return res, nil
}

// collectRefs returns a list of pointer to all RawNodes (n itself and all nested objects) that have a "$ref" key.
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
