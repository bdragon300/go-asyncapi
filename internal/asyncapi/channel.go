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
	Address    string                              `json:"address" yaml:"address"`
	Messages  types.OrderedMap[string, Message]    `json:"messages" yaml:"messages"`
	Title 	string                              `json:"title" yaml:"title"`
	Summary 	string                              `json:"summary" yaml:"summary"`
	Description string          `json:"description" yaml:"description"`
	Servers     []StandaloneRef `json:"servers" yaml:"servers"`
	Parameters  types.OrderedMap[string, Parameter] `json:"parameters" yaml:"parameters"`
	Tags 	  []Tag                               `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation             `json:"externalDocs" yaml:"externalDocs"`
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
	return nil
}

func (c Channel) build(ctx *common.CompileContext, channelKey string, flags map[common.SchemaTag]string) (common.Renderable, error) {
	ignore := c.XIgnore || (!ctx.CompileOpts.GeneratePublishers && !ctx.CompileOpts.GenerateSubscribers)
	if ignore {
		ctx.Logger.Debug("Channel denoted to be ignored")
		return &render.Channel{Dummy: true}, nil
	}

	_, isComponent := flags[common.SchemaTagComponent]
	if c.Ref != "" {
		// Make a promise selectable if it defined in `channels` section
		return registerRef(ctx,  c.Ref, channelKey, lo.Ternary(isComponent, nil, lo.ToPtr(true))), nil
	}

	chName, _ := lo.Coalesce(c.XGoName, channelKey)
	res := &render.Channel{
		OriginalName: chName,
		IsComponent:  isComponent,
		IsPublisher: ctx.CompileOpts.GeneratePublishers,
		IsSubscriber: ctx.CompileOpts.GenerateSubscribers,
	}

	// Servers which this channel is bound with
	if len(c.Servers) > 0 {
		ctx.Logger.Trace("Channel servers", "refs", c.Servers)
		for _, srvRef := range c.Servers {
			prm := lang.NewPromise[*render.Server](srvRef.Ref)
			res.ServersPromises = append(res.ServersPromises, prm)
			ctx.PutPromise(prm)
		}
	} else {
		ctx.Logger.Trace("Channel for all servers")
	}
	prm := lang.NewListCbPromise[common.Renderable](func(item common.CompileObject, path []string) bool {
		if len(path) < 2 || len(path) >= 2 && path[0] != "servers" {
			return false
		}
		return item.Kind() == common.ObjectKindServer && item.Visible()
	}, nil)
	res.AllActiveServersPromise = prm
	ctx.PutListPromise(prm)

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
			prm := lang.NewGolangTypeAssignCbPromise(ref, nil, func(obj any) common.GolangType {
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

	for _, msgEntry := range c.Messages.Entries() {
		msgName, ref := msgEntry.Key, msgEntry.Value
		ctx.Logger.Trace("Channel message", "name", msgName)
		refObj := lang.NewRef(ref.Ref, msgName, lo.ToPtr(false))
		ctx.PutPromise(refObj)
		res.MessagesPromises = append(res.MessagesPromises, refObj)
	}

	// All known Operations
	prmOp := lang.NewListCbPromise[common.Renderable](func(item common.CompileObject, path []string) bool {
		if len(path) < 2 || len(path) >= 2 && path[0] != "operations" {
			return false
		}
		return item.Kind() == common.ObjectKindOperation && item.Visible()
	}, nil)
	res.AllActiveOperationsPromise = prmOp
	ctx.PutListPromise(prmOp)

	// Bindings
	if c.Bindings != nil {
		ctx.Logger.Trace("Found channel bindings")

		ref := ctx.PathStackRef("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref)
		ctx.PutPromise(res.BindingsPromise)

		res.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(chName, "Bindings"),
				HasDefinition: true,
			},
		}
	}

	// Build protocol-specific channels for all supported protocols
	// At this point we don't have the actual protocols list to compile, because we don't know yet which servers this
	// channel is bound with -- it will be known only after linking stage.
	// So we just compile the proto channels for all supported protocols.
	ctx.Logger.Trace("Prebuild the channels for every supported protocol")
	for proto, b := range ProtocolBuilders {
		ctx.Logger.Trace("Channel", "proto", proto)
		ctx.Logger.NextCallLevel()
		obj, err := b.BuildChannel(ctx, &c, res)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
		res.ProtoChannels = append(res.ProtoChannels, obj)
	}

	return res, nil
}


type SecurityRequirement struct {
	types.OrderedMap[string, []string] // FIXME: orderedmap must be in fields as utils.OrderedMap[SecurityRequirement, []SecurityRequirement]
}
