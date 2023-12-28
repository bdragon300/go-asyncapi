package asyncapi

import (
	"encoding/json"
	"path"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

type protoChannelCompilerFunc func(ctx *common.CompileContext, channel *Channel, name string) (common.Renderer, error)

var ProtoChannelCompiler = map[string]protoChannelCompilerFunc{}

type Channel struct {
	Description string                                                             `json:"description" yaml:"description"`
	Servers     *[]string                                                          `json:"servers" yaml:"servers"`
	Subscribe   *Operation                                                         `json:"subscribe" yaml:"subscribe"`
	Publish     *Operation                                                         `json:"publish" yaml:"publish"`
	Parameters  types.OrderedMap[string, Parameter]                                `json:"parameters" yaml:"parameters"`
	Bindings    types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`

	XGoName string `json:"x-go-name" yaml:"x-go-name"`
	XIgnore bool   `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c Channel) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path) // TODO: use title
	obj, err := c.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (c Channel) build(ctx *common.CompileContext, channelKey string) (common.Renderer, error) {
	if c.XIgnore {
		ctx.Logger.Debug("Channel denoted to be ignored")
		return &render.Simple{Name: "any", IsIface: true}, nil
	}
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		res := render.NewRendererPromise(c.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	chName, _ := lo.Coalesce(c.XGoName, channelKey)
	res := &render.Channel{Name: chName, AllProtocols: make(map[string]common.Renderer)}

	// Channel parameters
	if c.Parameters.Len() > 0 {
		ctx.Logger.Trace("Channel parameters")
		ctx.Logger.NextCallLevel()
		res.ParametersStruct = &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(chName, "Parameters"),
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
		}
		for _, paramName := range c.Parameters.Keys() {
			ctx.Logger.Trace("Channel parameter", "name", paramName)
			ref := path.Join(ctx.PathRef(), "parameters", paramName)
			prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
			ctx.PutPromise(prm)
			res.ParametersStruct.Fields = append(res.ParametersStruct.Fields, render.StructField{
				Name: utils.ToGolangName(paramName, true),
				Type: prm,
			})
		}
		ctx.Logger.PrevCallLevel()
	}

	// Servers which this channel is connected to
	// Empty servers field means "no servers", omitted servers field means "all servers"
	if c.Servers != nil {
		ctx.Logger.Trace("Channel applied to particular servers", "names", *c.Servers)
		res.AppliedServers = *c.Servers
		prm := render.NewListCbPromise[*render.Server](func(item common.Renderer, path []string) bool {
			_, ok := item.(*render.Server)
			// Pick only servers from `servers:` section, skip ones from `components:`
			return ok && len(path) == 2 && path[0] == "servers" && lo.Contains(*c.Servers, path[1])
		})
		ctx.PutListPromise(prm)
		res.AppliedServersPromise = prm
	} else {
		ctx.Logger.Trace("Channel applied to all servers")
		prm := render.NewListCbPromise[*render.Server](func(item common.Renderer, path []string) bool {
			_, ok := item.(*render.Server)
			return ok && len(path) > 0 && path[0] == "servers" // Pick only servers from `servers:` section, skip ones from `components:`
		})
		ctx.PutListPromise(prm)
		res.AppliedServersPromise = prm
	}

	// Channel/operation bindings
	var hasBindings bool
	if c.Bindings.Len() > 0 {
		ctx.Logger.Trace("Found channel bindings")
		hasBindings = true
	}
	if c.Publish != nil && !c.Publish.XIgnore && c.Publish.Bindings.Len() > 0 {
		ctx.Logger.Trace("Found publish operation bindings")
		hasBindings = true
	}
	if c.Subscribe != nil && !c.Subscribe.XIgnore && c.Subscribe.Bindings.Len() > 0 {
		ctx.Logger.Trace("Found subscribe operation bindings")
		hasBindings = true
	}
	if hasBindings {
		res.BindingsStruct = &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(chName, "Bindings"),
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
		}
	}

	// Build protocol-specific channels
	for proto, f := range ProtoChannelCompiler {
		ctx.Logger.Trace("Channel", "proto", proto)
		ctx.Logger.NextCallLevel()
		obj, err := f(ctx, &c, channelKey)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
		res.AllProtocols[proto] = obj
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
	Bindings     types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`
	Traits       []OperationTrait                                                   `json:"traits" yaml:"traits"`
	// FIXME: can be either a message or map of messages?
	Message *Message `json:"message" yaml:"message"`

	XIgnore bool `json:"x-ignore" yaml:"x-ignore"`
}

type OperationTrait struct {
	OperationID  string                                                             `json:"operationId" yaml:"operationId"`
	Summary      string                                                             `json:"summary" yaml:"summary"`
	Description  string                                                             `json:"description" yaml:"description"`
	Security     []SecurityRequirement                                              `json:"security" yaml:"security"`
	Tags         []Tag                                                              `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation                                             `json:"externalDocs" yaml:"externalDocs"`
	Bindings     types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`

	Ref string `json:"$ref" yaml:"$ref"`
}

type SecurityRequirement struct {
	types.OrderedMap[string, []string] // FIXME: orderedmap must be in fields as utils.OrderedMap[SecurityRequirement, []SecurityRequirement]
}
