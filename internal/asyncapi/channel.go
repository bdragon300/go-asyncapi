package asyncapi

import (
	"path"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

type Channel struct {
	Description string                              `json:"description" yaml:"description"`
	Servers     *[]string                           `json:"servers" yaml:"servers"`
	Subscribe   *Operation                          `json:"subscribe" yaml:"subscribe"`
	Publish     *Operation                          `json:"publish" yaml:"publish"`
	Parameters  types.OrderedMap[string, Parameter] `json:"parameters" yaml:"parameters"`
	Bindings    *ChannelBindings                    `json:"bindings" yaml:"bindings"`

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
		return &render.GoSimple{Name: "any", IsIface: true}, nil
	}
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		res := render.NewRendererPromise(c.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	chName, _ := lo.Coalesce(c.XGoName, channelKey)
	res := &render.Channel{Name: chName, ChannelKey: channelKey, AllProtoChannels: make(map[string]common.Renderer)}

	// Channel parameters
	if c.Parameters.Len() > 0 {
		ctx.Logger.Trace("Channel parameters")
		ctx.Logger.NextCallLevel()
		res.ParametersStruct = &render.GoStruct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(chName, "Parameters"),
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
		}
		for _, paramName := range c.Parameters.Keys() {
			ctx.Logger.Trace("Channel parameter", "name", paramName)
			ref := path.Join(ctx.PathRef(), "parameters", paramName)
			prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
			ctx.PutPromise(prm)
			res.ParametersStruct.Fields = append(res.ParametersStruct.Fields, render.GoStructField{
				Name: utils.ToGolangName(paramName, true),
				Type: prm,
			})
		}
		ctx.Logger.PrevCallLevel()
	}

	// Servers which this channel is connected to
	// Empty servers field means "no servers", omitted servers field means "all servers"
	if c.Servers != nil {
		ctx.Logger.Trace("Channel servers", "names", *c.Servers)
		res.ExplicitServerNames = *c.Servers
		prm := render.NewListCbPromise[*render.Server](func(item common.Renderer, path []string) bool {
			_, ok := item.(*render.Server)
			// Pick only servers from `servers:` section, skip ones from `components:`
			return ok && len(path) == 2 && path[0] == "servers" && lo.Contains(*c.Servers, path[1])
		})
		ctx.PutListPromise(prm)
		res.ServersPromise = prm
	} else {
		ctx.Logger.Trace("Channel for all servers")
		prm := render.NewListCbPromise[*render.Server](func(item common.Renderer, path []string) bool {
			_, ok := item.(*render.Server)
			return ok && len(path) > 0 && path[0] == "servers" // Pick only servers from `servers:` section, skip ones from `components:`
		})
		ctx.PutListPromise(prm)
		res.ServersPromise = prm
	}

	// Channel/operation bindings
	var hasBindings bool
	if c.Bindings != nil {
		ctx.Logger.Trace("Found channel bindings")
		hasBindings = true

		ref := ctx.PathRef() + "/bindings"
		res.BindingsChannelPromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(res.BindingsChannelPromise)
	}
	if c.Publish != nil && !c.Publish.XIgnore && c.Publish.Bindings != nil {
		ctx.Logger.Trace("Found publish operation bindings")
		hasBindings = true

		ref := ctx.PathRef() + "/publish/bindings"
		res.BindingsPublishPromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(res.BindingsPublishPromise)
	}
	if c.Subscribe != nil && !c.Subscribe.XIgnore && c.Subscribe.Bindings != nil {
		ctx.Logger.Trace("Found subscribe operation bindings")
		hasBindings = true

		ref := ctx.PathRef() + "/subscribe/bindings"
		res.BindingsSubscribePromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(res.BindingsSubscribePromise)
	}
	if hasBindings {
		res.BindingsStruct = &render.GoStruct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(chName, "Bindings"),
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
		}
	}

	// Build protocol-specific channels for all supported protocols
	// Here we don't know yet which servers this channel is applied to, so we don't have the protocols list to compile.
	// Servers will be known on rendering stage (after linking), but there we will already need to have proto
	// channels to be compiled for certain protocols we want to render.
	// As a solution, here we just build the proto channels for all supported protocols
	for proto, b := range ProtocolBuilders {
		ctx.Logger.Trace("Channel", "proto", proto)
		ctx.Logger.NextCallLevel()
		obj, err := b.BuildChannel(ctx, &c, channelKey, res)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
		res.AllProtoChannels[proto] = obj
	}

	return res, nil
}

type Operation struct {
	OperationID  string                 `json:"operationId" yaml:"operationId"`
	Summary      string                 `json:"summary" yaml:"summary"`
	Description  string                 `json:"description" yaml:"description"`
	Security     []SecurityRequirement  `json:"security" yaml:"security"`
	Tags         []Tag                  `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
	Bindings     *OperationBinding      `json:"bindings" yaml:"bindings"`
	Traits       []OperationTrait       `json:"traits" yaml:"traits"`
	// FIXME: can be either a message or map of messages?
	Message *Message `json:"message" yaml:"message"`

	XIgnore bool `json:"x-ignore" yaml:"x-ignore"`
}

type OperationTrait struct {
	OperationID  string                 `json:"operationId" yaml:"operationId"`
	Summary      string                 `json:"summary" yaml:"summary"`
	Description  string                 `json:"description" yaml:"description"`
	Security     []SecurityRequirement  `json:"security" yaml:"security"`
	Tags         []Tag                  `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
	Bindings     *OperationBinding      `json:"bindings" yaml:"bindings"`

	Ref string `json:"$ref" yaml:"$ref"`
}

type SecurityRequirement struct {
	types.OrderedMap[string, []string] // FIXME: orderedmap must be in fields as utils.OrderedMap[SecurityRequirement, []SecurityRequirement]
}
