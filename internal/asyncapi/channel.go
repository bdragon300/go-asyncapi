package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

type Channel struct {
	Address string `json:"address" yaml:"address"`
	// Being referenced from a channel makes the message selectable and sets its generated name
	Messages     types.OrderedMap[string, Message]   `json:"messages" yaml:"messages" cgen:"selectable"`
	Title        string                              `json:"title" yaml:"title"`
	Summary      string                              `json:"summary" yaml:"summary"`
	Description  string                              `json:"description" yaml:"description"`
	Servers      []StandaloneRef                     `json:"servers" yaml:"servers"`
	Parameters   types.OrderedMap[string, Parameter] `json:"parameters" yaml:"parameters"`
	Tags         []Tag                               `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation              `json:"externalDocs" yaml:"externalDocs"`
	Bindings     *ChannelBindings                    `json:"bindings" yaml:"bindings"`

	XGoName string `json:"x-go-name" yaml:"x-go-name"`
	XIgnore bool   `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c Channel) Compile(ctx *compile.Context) error {
	obj, err := c.build(ctx, ctx.Stack.Top().Key, ctx.Stack.Top().Flags)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

type protocolObject interface {
	Protocol() string
}

func (c Channel) build(ctx *compile.Context, channelKey string, flags map[common.SchemaTag]string) (common.Artifact, error) {
	_, isSelectable := flags[common.SchemaTagSelectable]
	ignore := c.XIgnore || (!ctx.CompileOpts.GeneratePublishers && !ctx.CompileOpts.GenerateSubscribers)
	if ignore {
		ctx.Logger.Debug("Channel denoted to be ignored")
		return &render.Channel{Dummy: true}, nil
	}

	if c.Ref != "" {
		return registerRef(ctx, c.Ref, channelKey, lo.Ternary(isSelectable, lo.ToPtr(true), nil)), nil
	}

	chName, _ := lo.Coalesce(c.XGoName, channelKey)
	res := &render.Channel{
		OriginalName: chName,
		Address:      c.Address,
		IsSelectable: isSelectable,
		IsPublisher:  ctx.CompileOpts.GeneratePublishers,
		IsSubscriber: ctx.CompileOpts.GenerateSubscribers,
	}

	// Servers which this channel is bound with
	if len(c.Servers) > 0 {
		ctx.Logger.Trace("Channel servers", "refs", c.Servers)
		for _, srvRef := range c.Servers {
			prm := lang.NewPromise[*render.Server](srvRef.Ref, nil)
			res.ServersPromises = append(res.ServersPromises, prm)
			ctx.PutPromise(prm)
		}
	} else {
		ctx.Logger.Trace("Channel for all servers")
	}
	prm := lang.NewListCbPromise[common.Artifact](func(item common.Artifact) bool {
		path := item.Pointer().Pointer
		if len(path) < 2 || len(path) >= 2 && path[0] != "servers" {
			return false
		}
		return item.Kind() == common.ArtifactKindServer && item.Visible()
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
			ref := ctx.CurrentPositionRef("parameters", paramName)
			prmType := lang.NewGolangTypePromise(ref, func(obj common.Artifact) common.GolangType {
				return obj.(*render.Parameter).Type
			})
			ctx.PutPromise(prmType)
			res.ParametersType.Fields = append(res.ParametersType.Fields, lang.GoStructField{
				Name: utils.ToGolangName(paramName, true),
				Type: prmType,
			})

			prm := lang.NewRef(ref, paramName, nil)
			res.ParametersPromises = append(res.ParametersPromises, prm)
			ctx.PutPromise(prm)
		}
		ctx.Logger.PrevCallLevel()
	}

	for _, msgName := range c.Messages.Keys() {
		ctx.Logger.Trace("Channel message", "name", msgName)
		ref := ctx.CurrentPositionRef("messages", msgName)
		// Do not consider the name which a message $ref is registered with, keeping the original message name in code.
		prm2 := lang.NewRef(ref, "", nil)
		ctx.PutPromise(prm2)
		res.MessagesRefs = append(res.MessagesRefs, prm2)
	}

	// All known Operations
	prmOp := lang.NewListCbPromise[common.Artifact](func(item common.Artifact) bool {
		path := item.Pointer().Pointer
		if len(path) < 2 || len(path) >= 2 && path[0] != "operations" {
			return false
		}
		return item.Kind() == common.ArtifactKindOperation && item.Visible()
	}, nil)
	res.AllActiveOperationsPromise = prmOp
	ctx.PutListPromise(prmOp)

	// Bindings
	if c.Bindings != nil {
		ctx.Logger.Trace("Found channel bindings")

		ref := ctx.CurrentPositionRef("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref, nil)
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
	for _, b := range ctx.ProtocolBuilders {
		builder := b.(protocolObject)
		ctx.Logger.Trace("Channel", "proto", builder.Protocol())
		protoCh := &render.ProtoChannel{
			Channel:  res,
			Protocol: b.Protocol(),
		}
		res.ProtoChannels = append(res.ProtoChannels, protoCh)
	}

	return res, nil
}
