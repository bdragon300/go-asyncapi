package asyncapi

import (
	"errors"
	"fmt"
	"strconv"

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
	Action       OperationAction        `json:"action,omitzero" yaml:"action"`
	Channel      *StandaloneRef         `json:"channel,omitzero" yaml:"channel"`
	Title        string                 `json:"title,omitzero" yaml:"title"`
	Summary      string                 `json:"summary,omitzero" yaml:"summary"`
	Description  string                 `json:"description,omitzero" yaml:"description"`
	Security     []SecurityScheme       `json:"security,omitzero" yaml:"security"`
	Tags         []Tag                  `json:"tags,omitzero" yaml:"tags"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitzero" yaml:"externalDocs"`
	Bindings     *Bindings              `json:"bindings,omitzero" yaml:"bindings"`
	Traits       []OperationTrait       `json:"traits,omitzero" yaml:"traits"`
	Messages     *[]StandaloneRef       `json:"messages,omitzero" yaml:"messages"`
	Reply        *OperationReply        `json:"reply,omitzero" yaml:"reply"`

	XIgnore bool `json:"x-ignore,omitzero" yaml:"x-ignore"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
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
	if o.XIgnore {
		ctx.Logger.Debug("Operation denoted to be ignored")
		return &render.Operation{Dummy: true}, nil
	}

	_, isSelectable := flags[common.SchemaTagSelectable]
	if o.Ref != "" {
		// Make an operation selectable if it defined in `operations` section
		return registerRef(ctx, o.Ref, operationKey, lo.Ternary(isSelectable, lo.ToPtr(true), nil)), nil
	}

	res := &render.Operation{
		OriginalName:      operationKey,
		IsSelectable:      isSelectable,
		IsPublisher:       o.Action == OperationActionSend && ctx.CompileOpts.GeneratePublishers,
		IsSubscriber:      o.Action == OperationActionReceive && ctx.CompileOpts.GenerateSubscribers,
		IsReplyPublisher:  o.Reply != nil && o.Action == OperationActionReceive && ctx.CompileOpts.GeneratePublishers,
		IsReplySubscriber: o.Reply != nil && o.Action == OperationActionSend && ctx.CompileOpts.GenerateSubscribers,
	}

	if o.Channel == nil {
		return nil, types.CompileError{Err: errors.New("channel field is empty"), Path: ctx.CurrentRefPointer()}
	}

	ctx.Logger.Trace("Bound channel", "ref", o.Channel.Ref)
	prm := lang.NewPromise[*render.Channel](o.Channel.Ref, nil)
	ctx.PutPromise(prm)
	res.ChannelPromise = prm

	if o.Bindings != nil {
		ctx.Logger.Trace("Found operation bindings")

		ref := ctx.CurrentRefPointer("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref, nil)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Security
	if len(o.Security) > 0 {
		ctx.Logger.Trace("Server security schemes", "count", len(o.Security))
		for ind := range o.Security {
			ref := ctx.CurrentRefPointer("security", strconv.Itoa(ind))
			secPrm := lang.NewPromise[*render.SecurityScheme](ref, nil)
			ctx.PutPromise(secPrm)
			res.SecuritySchemePromises = append(res.SecuritySchemePromises, secPrm)
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

		ref := ctx.CurrentRefPointer("reply")
		res.OperationReplyPromise = lang.NewPromise[*render.OperationReply](ref, nil)
		ctx.PutPromise(res.OperationReplyPromise)
	}

	return res, nil
}

type OperationTrait struct {
	Title        string                 `json:"title,omitzero" yaml:"title"`
	Summary      string                 `json:"summary,omitzero" yaml:"summary"`
	Description  string                 `json:"description,omitzero" yaml:"description"`
	Security     []SecurityScheme       `json:"security,omitzero" yaml:"security"`
	Tags         []Tag                  `json:"tags,omitzero" yaml:"tags"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitzero" yaml:"externalDocs"`
	Bindings     *Bindings              `json:"bindings,omitzero" yaml:"bindings"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

type OperationReply struct {
	Address  *OperationReplyAddress `json:"address,omitzero" yaml:"address"`
	Channel  *StandaloneRef         `json:"channel,omitzero" yaml:"channel"`
	Messages *[]StandaloneRef       `json:"messages,omitzero" yaml:"messages"`

	XIgnore bool `json:"x-ignore,omitzero" yaml:"x-ignore"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
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

		ref := ctx.CurrentRefPointer("address")
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
	Location    string `json:"location,omitzero" yaml:"location"`
	Description string `json:"description,omitzero" yaml:"description"`

	XIgnore bool `json:"x-ignore,omitzero" yaml:"x-ignore"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
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
		return nil, types.CompileError{Err: fmt.Errorf("parse runtime expression: %w", err), Path: ctx.CurrentRefPointer()}
	}

	res := &render.OperationReplyAddress{
		OriginalName: operationKey,
		Description:  o.Description,
		BaseRuntimeExpression: lang.BaseRuntimeExpression{
			OriginalExpression: o.Location,
			StructFieldKind:    structField,
			LocationPath:       locationPath,
		},
	}

	return res, nil
}
