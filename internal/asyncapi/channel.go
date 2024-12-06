package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
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

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c Channel) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	chans, err := c.buildChannels(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	for _, ch := range chans {
		ctx.PutObject(ch)
	}
	return nil
}

func (c Channel) buildChannels(ctx *common.CompileContext, channelKey string) ([]common.Renderable, error) {
	var res []common.Renderable

	ignore := c.XIgnore ||
		(!ctx.CompileOpts.GeneratePublishers && !ctx.CompileOpts.GenerateSubscribers) // ||
		//!ctx.CompileOpts.ChannelOpts.IsAllowedName(channelKey)
	if ignore {
		ctx.Logger.Debug("Channel denoted to be ignored")
		res = append(res, &render.ProtoChannel{Channel: &render.Channel{Dummy: true}})
		return res, nil
	}
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		prm := lang.NewRenderablePromise(c.Ref, common.PromiseOriginUser)
		// Set a channel to be rendered if we reference it from `channels` document section
		ctx.PutPromise(prm)
		res = append(res, prm)
		return res, nil
	}

	chName, _ := lo.Coalesce(c.XGoName, channelKey)
	// Render only the channels defined directly in `channels` document section, not in `components`
	baseChan := &render.Channel{
		Name:                chName,
		TypeNamePrefix:      ctx.GenerateObjName(chName, ""),
		SpecKey:             channelKey,
	}

	// Channel parameters
	if c.Parameters.Len() > 0 {
		ctx.Logger.Trace("Channel parameters")
		ctx.Logger.NextCallLevel()
		baseChan.ParametersType = &lang.GoStruct{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(chName, "Parameters"),
				HasDefinition: true,
			},
		}
		for _, paramName := range c.Parameters.Keys() {
			ctx.Logger.Trace("Channel parameter", "name", paramName)
			ref := ctx.PathStackRef("parameters", paramName)
			prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
			ctx.PutPromise(prm)
			baseChan.ParametersType.Fields = append(baseChan.ParametersType.Fields, lang.GoStructField{
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
		baseChan.SpecServerNames = *c.Servers
		prm := lang.NewListCbPromise[*render.Server](func(item common.Renderable, path []string) bool {
			srv, ok := item.(*render.Server)
			if !ok {
				return false
			}
			return lo.Contains(*c.Servers, srv.Name)
		})
		baseChan.ServersPromise = prm
	} else {
		ctx.Logger.Trace("Channel for all servers")
		prm := lang.NewListCbPromise[*render.Server](func(item common.Renderable, path []string) bool {
			_, ok := item.(*render.Server)
			return ok
		})
		baseChan.ServersPromise = prm
	}

	// Channel/operation bindings
	var hasBindings bool
	if c.Bindings != nil {
		ctx.Logger.Trace("Found channel bindings")
		hasBindings = true

		ref := ctx.PathStackRef("bindings")
		baseChan.BindingsChannelPromise = lang.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(baseChan.BindingsChannelPromise)
	}

	genPub := c.Publish != nil && ctx.CompileOpts.GeneratePublishers
	genSub := c.Subscribe != nil && ctx.CompileOpts.GenerateSubscribers
	if genPub && !c.Publish.XIgnore {
		baseChan.IsPublisher = true
		if c.Publish.Bindings != nil {
			ctx.Logger.Trace("Found publish operation bindings")
			hasBindings = true

			ref := ctx.PathStackRef("publish", "bindings")
			baseChan.BindingsPublishPromise = lang.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
			ctx.PutPromise(baseChan.BindingsPublishPromise)
		}
		if c.Publish.Message != nil {
			ctx.Logger.Trace("Found publish operation message")
			ref := ctx.PathStackRef("publish", "message")
			baseChan.PublisherMessageTypePromise = lang.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(baseChan.PublisherMessageTypePromise)
		}
	}
	if genSub && !c.Subscribe.XIgnore {
		baseChan.IsSubscriber = true
		if c.Subscribe.Bindings != nil {
			ctx.Logger.Trace("Found subscribe operation bindings")
			hasBindings = true

			ref := ctx.PathStackRef("subscribe", "bindings")
			baseChan.BindingsSubscribePromise = lang.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
			ctx.PutPromise(baseChan.BindingsSubscribePromise)
		}
		if c.Subscribe.Message != nil {
			ctx.Logger.Trace("Channel subscribe operation message")
			ref := ctx.PathStackRef("subscribe", "message")
			baseChan.SubscribeMessageTypePromise = lang.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(baseChan.SubscribeMessageTypePromise)
		}
	}
	if hasBindings {
		baseChan.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(chName, "Bindings"),
				HasDefinition: true,
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
		obj, err := b.BuildChannel(ctx, &c, baseChan)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
		res = append(res, obj)
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
