package compile

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type protoChannelCompilerFunc func(ctx *common.CompileContext, channel *Channel, name string) (common.Assembler, error)

var ProtoChannelCompiler = map[string]protoChannelCompilerFunc{}

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
	ctx.SetTopObjName(ctx.Stack.Top().Path) // TODO: use title
	obj, err := c.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.PutToCurrentPkg(obj)
	return nil
}

func (c Channel) build(ctx *common.CompileContext, channelKey string) (common.Assembler, error) {
	if c.Ref != "" {
		res := assemble.NewRefLinkAsAssembler(c.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	res := &assemble.Channel{Name: channelKey, AllProtocols: make(map[string]common.Assembler)}

	// Channel parameters
	if c.Parameters.Len() > 0 {
		res.ParametersStruct = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    utils.ToGolangName(channelKey, true) + "Parameters",
				Render:  true,
				Package: ctx.TopPackageName(),
			},
		}
		for _, paramName := range c.Parameters.Keys() {
			ref := path.Join(ctx.PathRef(), "parameters", paramName)
			lnk := assemble.NewRefLinkAsGolangType(ref)
			ctx.Linker.Add(lnk)
			res.ParametersStruct.Fields = append(res.ParametersStruct.Fields, assemble.StructField{
				Name: utils.ToGolangName(paramName, true),
				Type: lnk,
			})
		}
	}

	// Servers which this channel is connected to
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

	// Channel/operation bindings
	if c.Bindings.Len() > 0 || c.Publish != nil && c.Publish.Bindings.Len() > 0 || c.Subscribe != nil && c.Subscribe.Bindings.Len() > 0 {
		res.BindingsStruct = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    ctx.GenerateObjName("", "Bindings"),
				Render:  true,
				Package: ctx.TopPackageName(),
			},
		}
	}

	if c.Parameters.Len() > 0 {
		res.ParametersStruct = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    ctx.GenerateObjName("", "Parameters"),
				Render:  true,
				Package: ctx.TopPackageName(),
			},
			Fields: nil,
		}
		for _, paramName := range c.Parameters.Keys() {
			ref := path.Join(ctx.PathRef(), "parameters", paramName)
			lnk := assemble.NewRefLinkAsGolangType(ref)
			ctx.Linker.Add(lnk)
			res.ParametersStruct.Fields = append(res.ParametersStruct.Fields, assemble.StructField{
				Name: utils.ToGolangName(paramName, true),
				Type: lnk,
			})
		}
	}

	// Build protocol-specific channels
	for pName, pBuild := range ProtoChannelCompiler {
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
