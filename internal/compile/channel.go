package compile

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type channelProtoBuilderFunc func(ctx *common.CompileContext, channel *Channel, name string) (common.Assembler, error)

var ProtoChannelBuilders = map[string]channelProtoBuilderFunc{}

type Channel struct {
	Description string                                                             `json:"description" yaml:"description"`
	Servers     *[]string                                                          `json:"servers" yaml:"servers"`
	Subscribe   *Operation                                                         `json:"subscribe" yaml:"subscribe"`
	Publish     *Operation                                                         `json:"publish" yaml:"publish"`
	Parameters  utils.OrderedMap[string, Parameter]                                `json:"parameters" yaml:"parameters"`
	Bindings    utils.OrderedMap[string, utils.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c Channel) Compile(ctx *common.CompileContext) error {
	ctx.SetObjName(ctx.Stack.Top().Path) // TODO: use title
	obj, err := c.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.CurrentPackage().Put(ctx, obj)
	return nil
}

func (c Channel) build(ctx *common.CompileContext, channelKey string) (common.Assembler, error) {
	if c.Ref != "" {
		res := assemble.NewRefLinkAsAssembler(c.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}
	res := &assemble.Channel{Name: channelKey, AllProtocols: make(map[string]common.Assembler)}
	// Empty servers field means "no servers", omitted servers field means "all servers"
	if c.Servers != nil && len(*c.Servers) > 0 {
		for _, srv := range *c.Servers {
			lnk := assemble.NewRefLink[*assemble.Server]("#/servers/" + srv)
			ctx.Linker.Add(lnk)
			res.AppliedServerLinks = append(res.AppliedServerLinks, lnk)
			res.AppliedServers = append(res.AppliedServers, srv)
		}
	} else if c.Servers == nil {
		lnk := assemble.NewListCbLink[*assemble.Server](func(item common.Assembler, path []string) bool {
			_, ok := item.(*assemble.Server)
			return ok && len(path) > 0 && path[0] == "servers" // Pick only servers from `servers:` section, skip ones from `components:`
		})
		ctx.Linker.AddMany(lnk)
		res.AppliedToAllServersLinks = lnk
	}

	for pName, pBuild := range ProtoChannelBuilders {
		obj, err := pBuild(ctx, &c, channelKey)
		if err != nil {
			return nil, fmt.Errorf("Unable to build %s protocol: %w", pName, err)
		}
		res.AllProtocols[pName] = obj
	}
	return res, nil
}

type Operation struct {
	OperationID  string                                                             `json:"operationId" yaml:"operationId"`
	Summary      string                                                             `json:"summary" yaml:"summary"`
	Description  string                                                             `json:"description" yaml:"description"`
	Security     []SecurityRequirement                                              `json:"security" yaml:"security"`
	Tags         []Tag                                                              `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation                                             `json:"externalDocs" yaml:"externalDocs"`
	Bindings     utils.OrderedMap[string, utils.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`
	Traits       []OperationTrait                                                   `json:"traits" yaml:"traits"`
	Message      *Message                                                           `json:"message" yaml:"message"`
}

type Parameter struct {
	Description string  `json:"description" yaml:"description"`
	Schema      *Object `json:"schema" yaml:"schema"`     // TODO: implement
	Location    string  `json:"location" yaml:"location"` // TODO: implement

	Ref string `json:"$ref" yaml:"$ref"`
}

type OperationTrait struct {
	OperationID  string                                                             `json:"operationId" yaml:"operationId"`
	Summary      string                                                             `json:"summary" yaml:"summary"`
	Description  string                                                             `json:"description" yaml:"description"`
	Security     []SecurityRequirement                                              `json:"security" yaml:"security"`
	Tags         []Tag                                                              `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation                                             `json:"externalDocs" yaml:"externalDocs"`
	Bindings     utils.OrderedMap[string, utils.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`

	Ref string `json:"$ref" yaml:"$ref"`
}

type SecurityRequirement struct {
	utils.OrderedMap[string, []string]
}
