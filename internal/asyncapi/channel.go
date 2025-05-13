package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/specurl"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/utils"
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
	Address string `json:"address" yaml:"address"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c Channel) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj, err := c.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (c Channel) build(ctx *common.CompileContext, channelKey string) (common.Renderer, error) {
	_, isComponent := ctx.Stack.Top().Flags[common.SchemaTagComponent]
	ignore := c.XIgnore ||
		(!ctx.CompileOpts.GeneratePublishers && !ctx.CompileOpts.GenerateSubscribers) ||
		!ctx.CompileOpts.ChannelOpts.IsAllowedName(channelKey)
	if ignore {
		ctx.Logger.Debug("Channel denoted to be ignored")
		return &render.Channel{Dummy: true}, nil
	}
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		prm := render.NewRendererPromise(c.Ref, common.PromiseOriginUser)
		// Set a channel to be rendered if we reference it from `channels` document section
		prm.DirectRender = !isComponent
		ctx.PutPromise(prm)
		return prm, nil
	}

	chName, _ := lo.Coalesce(c.XGoName, channelKey)
	// Render only the channels defined directly in `channels` document section, not in `components`
	res := &render.Channel{
		Name:                chName,
		Address:             c.Address,
		GolangName:          ctx.GenerateObjName(chName, ""),
		RawName:             channelKey,
		AllProtoChannels:    make(map[string]common.Renderer),
		DirectRender:        !isComponent,
		FallbackMessageType: &render.GoSimple{Name: "any", IsIface: true},
	}

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
			ref := ctx.PathStackRef("parameters", paramName)
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
		prms := lo.FilterMap(ctx.Storage.ActiveServers(), func(item string, _ int) (*render.Promise[*render.Server], bool) {
			if lo.Contains(*c.Servers, item) {
				ref := specurl.BuildRef("servers", item)
				prm := render.NewPromise[*render.Server](ref, common.PromiseOriginInternal)
				ctx.PutPromise(prm)
				return prm, true
			}
			return nil, false
		})
		res.ServersPromises = prms
	} else {
		ctx.Logger.Trace("Channel for all servers")
		prms := lo.Map(ctx.Storage.ActiveServers(), func(item string, _ int) *render.Promise[*render.Server] {
			ref := specurl.BuildRef("servers", item)
			prm := render.NewPromise[*render.Server](ref, common.PromiseOriginInternal)
			ctx.PutPromise(prm)
			return prm
		})

		res.ServersPromises = prms
	}

	// Channel/operation bindings
	var hasBindings bool
	if c.Bindings != nil {
		ctx.Logger.Trace("Found channel bindings")
		hasBindings = true

		ref := ctx.PathStackRef("bindings")
		res.BindingsChannelPromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(res.BindingsChannelPromise)
	}

	genPub := c.Publish != nil && ctx.CompileOpts.GeneratePublishers
	genSub := c.Subscribe != nil && ctx.CompileOpts.GenerateSubscribers
	if genPub && !c.Publish.XIgnore {
		res.Publisher = true
		if c.Publish.Bindings != nil {
			ctx.Logger.Trace("Found publish operation bindings")
			hasBindings = true

			ref := ctx.PathStackRef("publish", "bindings")
			res.BindingsPublishPromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
			ctx.PutPromise(res.BindingsPublishPromise)
		}
		if c.Publish.Message != nil {
			ctx.Logger.Trace("Found publish operation message")
			ref := ctx.PathStackRef("publish", "message")
			res.PubMessagePromise = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(res.PubMessagePromise)
		}
	}
	if genSub && !c.Subscribe.XIgnore {
		res.Subscriber = true
		if c.Subscribe.Bindings != nil {
			ctx.Logger.Trace("Found subscribe operation bindings")
			hasBindings = true

			ref := ctx.PathStackRef("subscribe", "bindings")
			res.BindingsSubscribePromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
			ctx.PutPromise(res.BindingsSubscribePromise)
		}
		if c.Subscribe.Message != nil {
			ctx.Logger.Trace("Channel subscribe operation message")
			ref := ctx.PathStackRef("subscribe", "message")
			res.SubMessagePromise = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(res.SubMessagePromise)
		}
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
	ctx.Logger.Trace("Prebuild the channels for every supported protocol")
	for proto, b := range ProtocolBuilders {
		ctx.Logger.Trace("Channel", "proto", proto)
		ctx.Logger.NextCallLevel()
		obj, err := b.BuildChannel(ctx, &c, res)
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

func (c Operation) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	return nil
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
