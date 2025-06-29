package asyncapi

import (
	"encoding/json"
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type RawBindings struct {
	Ref            string                                                             `json:"$ref" yaml:"$ref"`
	ProtocolValues types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"-" yaml:"-"`
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

type bindingsProtoBuilder interface {
	BuildMessageBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildOperationBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildChannelBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildServerBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)
}

func (b *Bindings) build(
	ctx *compile.Context,
	bindingsKind int,
	bindingsKey string,
) (common.Artifact, error) {
	if b.Ref != "" {
		return registerRef(ctx, b.Ref, bindingsKey, nil), nil
	}

	res := render.Bindings{OriginalName: bindingsKey}
	for _, e := range b.ProtocolValues.Entries() {
		ctx.Logger.Trace("Bindings", "proto", e.Key)
		bld, ok := ctx.GetProtocolBuilder(e.Key)
		if !ok {
			ctx.Logger.Warn("Skip bindings protocol %q since it is not supported", "proto", e.Key)
			continue
		}
		builder := bld.(bindingsProtoBuilder)

		var vals *lang.GoValue
		var jsonVals types.OrderedMap[string, string]
		var err error

		switch bindingsKind {
		case bindingsKindMessage:
			ctx.Logger.Trace("Message bindings", "proto", e.Key)
			vals, jsonVals, err = builder.BuildMessageBindings(ctx, e.Value)
		case bindingsKindOperation:
			ctx.Logger.Trace("Operation bindings", "proto", e.Key)
			vals, jsonVals, err = builder.BuildOperationBindings(ctx, e.Value)
		case bindingsKindChannel:
			ctx.Logger.Trace("Channel bindings", "proto", e.Key)
			vals, jsonVals, err = builder.BuildChannelBindings(ctx, e.Value)
		case bindingsKindServer:
			ctx.Logger.Trace("Server bindings", "proto", e.Key)
			vals, jsonVals, err = builder.BuildServerBindings(ctx, e.Value)
		}
		if err != nil {
			return nil, types.CompileError{Err: fmt.Errorf("bindings build: %w", err), Path: ctx.CurrentPositionRef(), Proto: e.Key}
		}
		if vals != nil {
			ctx.Logger.Trace("Have bindings values", "proto", e.Key, "value", vals)
			res.Values.Set(e.Key, vals)
		}
		if jsonVals.Len() > 0 {
			ctx.Logger.Trace("Have bindings jsonschema values", "proto", e.Key, "keys", jsonVals.Keys())
			res.JSONValues.Set(e.Key, jsonVals)
		}
	}

	return &res, nil
}

func (b *Bindings) UnmarshalJSON(value []byte) error {
	if err := json.Unmarshal(value, &b.RawBindings); err != nil {
		return err
	}
	if err := json.Unmarshal(value, &b.ProtocolValues); err != nil {
		return err
	}
	b.ProtocolValues.Delete("$ref")
	return nil
}

func (b *Bindings) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode(&b.RawBindings); err != nil {
		return err
	}
	if err := value.Decode(&b.ProtocolValues); err != nil {
		return err
	}
	b.ProtocolValues.Delete("$ref")
	return nil
}

type MessageBindings struct {
	Bindings
}

func (b *MessageBindings) Compile(ctx *compile.Context) error {
	obj, err := b.build(ctx, bindingsKindMessage, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

type OperationBinding struct {
	Bindings
}

func (b *OperationBinding) Compile(ctx *compile.Context) error {
	obj, err := b.build(ctx, bindingsKindOperation, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

type ChannelBindings struct {
	Bindings
}

func (b *ChannelBindings) Compile(ctx *compile.Context) error {
	obj, err := b.build(ctx, bindingsKindChannel, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

type ServerBindings struct {
	Bindings
}

func (b *ServerBindings) Compile(ctx *compile.Context) error {
	obj, err := b.build(ctx, bindingsKindServer, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}
