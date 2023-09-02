package compile

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type channelProtoBuilderFunc func(ctx *common.CompileContext, name string) (assemble.ChannelParts, error)

type Channel struct {
	Description string                              `json:"description" yaml:"description"`
	Servers     *[]string                           `json:"servers" yaml:"servers"`
	Subscribe   *Operation                          `json:"subscribe" yaml:"subscribe"`
	Publish     *Operation                          `json:"publish" yaml:"publish"`
	Parameters  utils.OrderedMap[string, Parameter] `json:"parameters" yaml:"parameters"`
	Bindings    utils.OrderedMap[string, any]       `json:"bindings" yaml:"bindings"` // TODO: replace any to common protocols object

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c Channel) Compile(ctx *common.CompileContext) error {
	obj, err := c.build(ctx, ctx.Top().Path)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.CurrentPackage().Put(ctx, obj)
	return nil
}

func (c Channel) build(ctx *common.CompileContext, name string) (common.Assembler, error) {
	if c.Ref != "" {
		res := assemble.NewRefLinkAsAssembler(common.ChannelsPackageKind, c.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}
	res := &assemble.Channel{Name: name, SupportedProtocols: make(map[string]assemble.ChannelParts)}
	// Empty servers field means "no servers", omitted servers field means "all servers"
	if c.Servers != nil && len(*c.Servers) > 0 {
		for _, srv := range *c.Servers {
			lnk := assemble.NewRefLink[*assemble.Server](common.ServersPackageKind, "#/servers/"+srv)
			ctx.Linker.Add(lnk)
			res.AppliedServerLinks = append(res.AppliedServerLinks, lnk)
			res.AppliedServers = append(res.AppliedServers, srv)
		}
	} else if c.Servers == nil {
		lnk := assemble.NewListCbLink[*assemble.Server](common.ServersPackageKind, func(item any, path []string) bool {
			_, ok := item.(*assemble.Server)
			return ok && len(path) > 0 && path[0] == "servers" // Pick only servers from `servers:` section, skip ones from `components:`
		})
		ctx.Linker.AddMany(lnk)
		res.AppliedToAllServersLinks = lnk
	}

	for pName, pBuild := range c.supportedProtocols() {
		obj, err := pBuild(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("Unable to build %s protocol: %w", pName, err)
		}
		res.SupportedProtocols[pName] = obj
	}
	return res, nil
}

func (c Channel) supportedProtocols() map[string]channelProtoBuilderFunc {
	return map[string]channelProtoBuilderFunc{
		"kafka":        c.buildKafka,
		"kafka-secure": c.buildKafka,
	}
}

type Operation struct {
	OperationID  string                        `json:"operationId" yaml:"operationId"`
	Summary      string                        `json:"summary" yaml:"summary"`
	Description  string                        `json:"description" yaml:"description"`
	Security     []SecurityRequirement         `json:"security" yaml:"security"`
	Tags         []Tag                         `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation        `json:"externalDocs" yaml:"externalDocs"`
	Bindings     utils.OrderedMap[string, any] `json:"bindings" yaml:"bindings"` // TODO: replace any to common protocols object
	Traits       []OperationTrait              `json:"traits" yaml:"traits"`
	Message      *Message                      `json:"message" yaml:"message"`
}

type Parameter struct {
	Description string  `json:"description" yaml:"description"`
	Schema      *Object `json:"schema" yaml:"schema"`
	Location    string  `json:"location" yaml:"location"`

	Ref string `json:"$ref" yaml:"$ref"`
}

type OperationTrait struct {
	OperationID  string                        `json:"operationId" yaml:"operationId"`
	Summary      string                        `json:"summary" yaml:"summary"`
	Description  string                        `json:"description" yaml:"description"`
	Security     []SecurityRequirement         `json:"security" yaml:"security"`
	Tags         []Tag                         `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation        `json:"externalDocs" yaml:"externalDocs"`
	Bindings     utils.OrderedMap[string, any] `json:"bindings" yaml:"bindings"` // TODO: replace any to common protocols object

	Ref string `json:"$ref" yaml:"$ref"`
}

type SecurityRequirement struct {
	utils.OrderedMap[string, []string]
}
