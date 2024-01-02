package asyncapi

import (
	"encoding/json"
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"gopkg.in/yaml.v3"
)

type RawBindings struct {
	Ref            string                                                             `json:"$ref" yaml:"$ref"`
	ProtocolValues types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"-" yaml:"-"`
	// TODO: x-tags
}

type Bindings struct {
	RawBindings
}

const (
	bindingsKindMessage = iota
	bindingsKindOperation
	bindingsKindChannel
	bindingsKindServer
)

func (b *Bindings) build(
	ctx *common.CompileContext,
	bindingsKind int,
) (*render.Bindings, error) {
	res := render.Bindings{Path: ctx.PathRef()}

	for _, e := range b.ProtocolValues.Entries() {
		ctx.Logger.Trace("Bindings", "proto", e.Key)
		builder, ok := ProtocolBuilders[e.Key]
		if !ok {
			ctx.Logger.Warn("Skip bindings protocol %q since it is not supported", "proto", e.Key)
			continue
		}

		var vals *render.GoValue
		var jsonVals types.OrderedMap[string, string]
		var err error

		switch bindingsKind {
		case bindingsKindMessage:
			vals, jsonVals, err = builder.BuildMessageBindings(ctx, e.Value)
		case bindingsKindOperation:
			vals, jsonVals, err = builder.BuildOperationBindings(ctx, e.Value)
		case bindingsKindChannel:
			vals, jsonVals, err = builder.BuildChannelBindings(ctx, e.Value)
		case bindingsKindServer:
			vals, jsonVals, err = builder.BuildServerBindings(ctx, e.Value)
		}
		if err != nil {
			return nil, types.CompileError{Err: fmt.Errorf("bindings parse: %w", err), Path: ctx.PathRef(), Proto: e.Key}
		}
		if vals != nil {
			res.Values.Set(e.Key, vals)
		}
		if jsonVals.Len() > 0 {
			res.JSONValues.Set(e.Key, jsonVals)
		}
	}

	return &res, nil
}

func (b *Bindings) UnmarshalJSON(value []byte) error {
	if err := json.Unmarshal(value, &b.RawBindings); err != nil {
		return err
	}
	if err := json.Unmarshal(value, &b.RawBindings.ProtocolValues); err != nil {
		return err
	}
	b.ProtocolValues.Delete("$ref")
	return nil
}

func (b *Bindings) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&b.RawBindings); err != nil {
		return err
	}
	if err := value.Decode(&b.RawBindings.ProtocolValues); err != nil {
		return err
	}
	b.ProtocolValues.Delete("$ref")
	return nil
}

type MessageBindings struct {
	Bindings
}

func (b *MessageBindings) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := b.build(ctx, bindingsKindMessage)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

type OperationBinding struct {
	Bindings
}

func (b *OperationBinding) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := b.build(ctx, bindingsKindOperation)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

type ChannelBindings struct {
	Bindings
}

func (b *ChannelBindings) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := b.build(ctx, bindingsKindChannel)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

type ServerBindings struct {
	Bindings
}

func (b *ServerBindings) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := b.build(ctx, bindingsKindServer)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}
