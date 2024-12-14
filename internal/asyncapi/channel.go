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
	obj, err := c.build(ctx, ctx.Stack.Top().PathItem, ctx.Stack.Top().Flags)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	//if v, ok := obj.(*render.Channel); ok {
	//	ctx.Logger.Trace("Objects", "object", obj)
	//	for _, protoObj := range v.ProtoChannels {
	//		ctx.PutObject(protoObj)
	//	}
	//}
	return nil
}

func (c Channel) build(ctx *common.CompileContext, channelKey string, flags map[common.SchemaTag]string) (common.Renderable, error) {
	ignore := c.XIgnore ||
		(!ctx.CompileOpts.GeneratePublishers && !ctx.CompileOpts.GenerateSubscribers) // ||
		//!ctx.CompileOpts.ChannelOpts.IsAllowedName(channelKey)
	_, isComponent := flags[common.SchemaTagComponent]
	if ignore {
		ctx.Logger.Debug("Channel denoted to be ignored")
		return &render.Channel{Dummy: true}, nil
	}
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		// Always draw the promises that are located in the `channels` section
		prm := lang.NewUserPromise(c.Ref, channelKey, lo.Ternary(isComponent, nil, lo.ToPtr(true)))
		ctx.PutPromise(prm)
		return prm, nil
	}

	chName, _ := lo.Coalesce(c.XGoName, channelKey)
	// Render only the channels defined directly in `channels` document section, not in `components`
	res := &render.Channel{
		OriginalName:   chName,
		TypeNamePrefix: ctx.GenerateObjName(chName, ""),
		SpecKey:        channelKey,
		IsComponent:    isComponent,
	}

	// Channel parameters
	if c.Parameters.Len() > 0 {
		ctx.Logger.Trace("Channel parameters")
		ctx.Logger.NextCallLevel()
		res.ParametersType = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(chName, "Parameters"),
				HasDefinition: true,
			},
		}
		for _, paramName := range c.Parameters.Keys() {
			ctx.Logger.Trace("Channel parameter", "name", paramName)
			ref := ctx.PathStackRef("parameters", paramName)
			prm := lang.NewInternalGolangTypeAssignCbPromise(ref, func(obj any) common.GolangType {
				return obj.(*render.Parameter).Type
			})
			ctx.PutPromise(prm)
			res.ParametersType.Fields = append(res.ParametersType.Fields, lang.GoStructField{
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
		res.SpecServerNames = *c.Servers
		prm := lang.NewListCbPromise[*render.Server](func(item common.CompileObject, path []string) bool {
			srv, ok := item.Renderable.(*render.Server)
			if !ok {
				return false
			}
			return lo.Contains(*c.Servers, srv.OriginalName)
		})
		res.ServersPromise = prm
		ctx.PutListPromise(prm)
	} else {
		ctx.Logger.Trace("Channel for all servers")
		prm := lang.NewListCbPromise[*render.Server](func(item common.CompileObject, path []string) bool {
			_, ok := item.Renderable.(*render.Server)
			return ok
		})
		res.ServersPromise = prm
		ctx.PutListPromise(prm)
	}

	// Channel/operation bindings
	var hasBindings bool
	if c.Bindings != nil {
		ctx.Logger.Trace("Found channel bindings")
		hasBindings = true

		ref := ctx.PathStackRef("bindings")
		res.BindingsChannelPromise = lang.NewInternalPromise[*render.Bindings](ref)
		ctx.PutPromise(res.BindingsChannelPromise)
	}

	genPub := c.Publish != nil && ctx.CompileOpts.GeneratePublishers
	genSub := c.Subscribe != nil && ctx.CompileOpts.GenerateSubscribers
	if genPub && !c.Publish.XIgnore {
		res.IsPublisher = true
		if c.Publish.Bindings != nil {
			ctx.Logger.Trace("Found publish operation bindings")
			hasBindings = true

			ref := ctx.PathStackRef("publish", "bindings")
			res.BindingsPublishPromise = lang.NewInternalPromise[*render.Bindings](ref)
			ctx.PutPromise(res.BindingsPublishPromise)
		}
		if c.Publish.Message != nil {
			ctx.Logger.Trace("Found publish operation message")
			ref := ctx.PathStackRef("publish", "message")
			res.PublisherMessageTypePromise = lang.NewInternalPromise[*render.Message](ref)
			ctx.PutPromise(res.PublisherMessageTypePromise)
		}
	}
	if genSub && !c.Subscribe.XIgnore {
		res.IsSubscriber = true
		if c.Subscribe.Bindings != nil {
			ctx.Logger.Trace("Found subscribe operation bindings")
			hasBindings = true

			ref := ctx.PathStackRef("subscribe", "bindings")
			res.BindingsSubscribePromise = lang.NewInternalPromise[*render.Bindings](ref)
			ctx.PutPromise(res.BindingsSubscribePromise)
		}
		if c.Subscribe.Message != nil {
			ctx.Logger.Trace("Channel subscribe operation message")
			ref := ctx.PathStackRef("subscribe", "message")
			res.SubscriberMessageTypePromise = lang.NewInternalPromise[*render.Message](ref)
			ctx.PutPromise(res.SubscriberMessageTypePromise)
		}
	}
	if hasBindings {
		res.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(chName, "Bindings"),
				HasDefinition: true,
			},
		}
	}

	// Build protocol-specific channels for all supported protocols
	// At this point we don't know yet which servers this channel is applied to, so we don't have the protocols list to compile.
	// Servers will be known on rendering stage (after linking), but there we will already need to have proto
	// channels to be compiled for certain protocols we want to render.
	// As a solution, here we just build the proto channels for all supported protocols
	ctx.Logger.Trace("Prebuild the channels for every supported protocol")
	var protoChannels []*render.ProtoChannel
	for proto, b := range ProtocolBuilders {
		ctx.Logger.Trace("Channel", "proto", proto)
		ctx.Logger.NextCallLevel()
		obj, err := b.BuildChannel(ctx, &c, res)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
		protoChannels = append(protoChannels, obj)
	}
	res.ProtoChannels = protoChannels

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
