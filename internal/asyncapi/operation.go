package asyncapi

import (
	"errors"
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type OperationAction string

const (
	OperationActionSend    OperationAction = "send"
	OperationActionReceive OperationAction = "receive"
)

type Operation struct {
	Action      OperationAction `json:"action" yaml:"action"`
	Channel     *StandaloneRef  `json:"channel" yaml:"channel"`
	Title       string          `json:"title" yaml:"title"`
	Summary     string          `json:"summary" yaml:"summary"`
	Description string          `json:"description" yaml:"description"`
	// Security     SecurityScheme  `json:"security" yaml:"security"`
	Tags         []Tag                  `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
	Bindings     *OperationBinding      `json:"bindings" yaml:"bindings"`
	Traits       []OperationTrait       `json:"traits" yaml:"traits"`
	Messages     *[]StandaloneRef       `json:"messages" yaml:"messages"`
	Reply        *OperationReply        `json:"reply" yaml:"reply"`

	XIgnore bool `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (o Operation) Compile(ctx *compile.Context) error {
	obj, err := o.build(ctx, ctx.Stack.Top().Key, ctx.Stack.Top().Flags)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (o Operation) build(ctx *compile.Context, operationKey string, flags map[common.SchemaTag]string) (common.Artifact, error) {
	ignore := o.XIgnore ||
		o.Action == OperationActionSend && !ctx.CompileOpts.GeneratePublishers ||
		o.Action == OperationActionReceive && !ctx.CompileOpts.GenerateSubscribers
	if ignore {
		ctx.Logger.Debug("Operation denoted to be ignored")
		return &render.Operation{Dummy: true}, nil
	}

	_, isSelectable := flags[common.SchemaTagSelectable]
	if o.Ref != "" {
		// Make an operation selectable if it defined in `operations` section
		return registerRef(ctx, o.Ref, operationKey, lo.Ternary(isSelectable, lo.ToPtr(true), nil)), nil
	}

	res := &render.Operation{
		OriginalName:    operationKey,
		IsSelectable:    isSelectable,
		IsPublisher:     o.Action == OperationActionSend && ctx.CompileOpts.GeneratePublishers,
		IsSubscriber:    o.Action == OperationActionReceive && ctx.CompileOpts.GenerateSubscribers,
		AllowTwoWayCode: ctx.CompileOpts.GeneratePublishers && ctx.CompileOpts.GenerateSubscribers,
	}

	if o.Channel == nil {
		return nil, types.CompileError{Err: errors.New("channel field is empty"), Path: ctx.CurrentPositionRef()}
	}

	ctx.Logger.Trace("Bound channel", "ref", o.Channel.Ref)
	prm := lang.NewPromise[*render.Channel](o.Channel.Ref, nil)
	ctx.PutPromise(prm)
	res.ChannelPromise = prm

	if o.Bindings != nil {
		ctx.Logger.Trace("Found operation bindings")

		ref := ctx.CurrentPositionRef("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref, nil)
		ctx.PutPromise(res.BindingsPromise)

		res.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(operationKey, "Bindings"),
				HasDefinition: true,
			},
		}
	}

	if o.Messages != nil {
		for _, message := range *o.Messages {
			ctx.Logger.Trace("Operation message", "ref", message.Ref)
			prm := lang.NewPromise[*render.Message](message.Ref, nil)
			ctx.PutPromise(prm)
			res.MessagesPromises = append(res.MessagesPromises, prm)
		}
	} else {
		ctx.Logger.Trace("Using all messages in the channel for this operation")
		res.UseAllChannelMessages = true
	}

	if o.Reply != nil {
		ctx.Logger.Trace("Found operation reply")

		ref := ctx.CurrentPositionRef("reply")
		res.OperationReplyPromise = lang.NewPromise[*render.OperationReply](ref, nil)
		ctx.PutPromise(res.OperationReplyPromise)
	}

	// Build protocol-specific operations for all supported protocols
	// At this point we don't have the actual protocols list to compile, because we don't know yet which servers the
	// channel is bound with -- it will be known only after linking stage.
	// So we just compile the proto operations for all supported protocols.
	ctx.Logger.Trace("Prebuild the operations for every supported protocol")
	for _, proto := range ctx.SupportedProtocols() {
		ctx.Logger.Trace("Operation", "proto", proto)
		prmCh := lang.NewPromise[*render.ProtoChannel](o.Channel.Ref, func(obj common.Artifact) *render.ProtoChannel {
			ch := obj.(*render.Channel)
			if ch.Dummy {
				return &render.ProtoChannel{Channel: ch, Protocol: proto} // Dummy channel
			}
			protoCh, found := lo.Find(ch.ProtoChannels, func(p *render.ProtoChannel) bool {
				return p.Protocol == proto
			})
			if !found {
				panic(fmt.Sprintf("ProtoChannel[%s] not found in %s. This is a bug", proto, ch))
			}
			return protoCh
		})
		ctx.PutPromise(prmCh)

		protoOp := &render.ProtoOperation{
			Operation:           res,
			ProtoChannelPromise: prmCh,
			Protocol:            proto,
		}
		res.ProtoOperations = append(res.ProtoOperations, protoOp)
	}

	return res, nil
}

type OperationTrait struct {
	Title       string `json:"title" yaml:"title"`
	Summary     string `json:"summary" yaml:"summary"`
	Description string `json:"description" yaml:"description"`
	// Security     SecurityScheme  `json:"security" yaml:"security"`
	Tags         []Tag                  `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
	Bindings     *OperationBinding      `json:"bindings" yaml:"bindings"`

	Ref string `json:"$ref" yaml:"$ref"`
}

type OperationReply struct {
	Address  *OperationReplyAddress `json:"address" yaml:"address"`
	Channel  *StandaloneRef         `json:"channel" yaml:"channel"`
	Messages *[]StandaloneRef       `json:"messages" yaml:"messages"`

	XIgnore bool `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (o OperationReply) Compile(ctx *compile.Context) error {
	obj, err := o.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (o OperationReply) build(ctx *compile.Context, operationKey string) (common.Artifact, error) {
	if o.XIgnore {
		ctx.Logger.Debug("OperationReply denoted to be ignored")
		return &render.OperationReply{Dummy: true}, nil
	}
	if o.Ref != "" {
		return registerRef(ctx, o.Ref, operationKey, nil), nil
	}

	res := &render.OperationReply{
		OriginalName: operationKey,
	}

	if o.Address != nil {
		ctx.Logger.Trace("Found operation reply address")

		ref := ctx.CurrentPositionRef("address")
		res.OperationReplyAddressPromise = lang.NewPromise[*render.OperationReplyAddress](ref, nil)
		ctx.PutPromise(res.OperationReplyAddressPromise)
	}

	if o.Channel != nil {
		ctx.Logger.Trace("Bound channel", "ref", o.Channel.Ref)
		prm := lang.NewPromise[*render.Channel](o.Channel.Ref, nil)
		ctx.PutPromise(prm)
		res.ChannelPromise = prm
	}

	if o.Messages != nil {
		for _, message := range *o.Messages {
			ctx.Logger.Trace("Operation reply message", "ref", message.Ref)
			prm := lang.NewPromise[*render.Message](message.Ref, nil)
			ctx.PutPromise(prm)
			res.MessagesPromises = append(res.MessagesPromises, prm)
		}
	} else {
		ctx.Logger.Trace("Using all messages in the channel for this operation reply (if any)")
		res.UseAllChannelMessages = true
	}

	return res, nil
}

type OperationReplyAddress struct {
	Location    string `json:"location" yaml:"location"`
	Description string `json:"description" yaml:"description"`

	XIgnore bool `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (o OperationReplyAddress) Compile(ctx *compile.Context) error {
	obj, err := o.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (o OperationReplyAddress) build(ctx *compile.Context, operationKey string) (common.Artifact, error) {
	if o.XIgnore {
		ctx.Logger.Debug("OperationReplyAddress denoted to be ignored")
		return &render.OperationReplyAddress{Dummy: true}, nil
	}
	if o.Ref != "" {
		return registerRef(ctx, o.Ref, operationKey, nil), nil
	}

	ctx.Logger.Trace("Parsing OperationReplyAddress location runtime expression", "location", o.Location)
	structField, locationPath, err := parseRuntimeExpression(o.Location)
	if err != nil {
		return nil, types.CompileError{Err: fmt.Errorf("parse runtime expression: %w", err), Path: ctx.CurrentPositionRef()}
	}

	res := &render.OperationReplyAddress{
		OriginalName: operationKey,
		Description:  o.Description,
		BaseRuntimeExpression: lang.BaseRuntimeExpression{
			StructFieldKind: structField,
			LocationPath:    locationPath,
		},
	}

	return res, nil
}
