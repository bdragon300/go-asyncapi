package asyncapi

import (
	"encoding/json"
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type RawBindings struct {
	Ref            string                                                             `json:"$ref,omitzero" yaml:"$ref"`
	ProtocolValues types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"-" yaml:"-"`
}

type Bindings struct {
	RawBindings
}

func (b *Bindings) Compile(ctx *compile.Context) error {
	obj, err := b.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (b *Bindings) build(ctx *compile.Context, bindingsKey string) (common.Artifact, error) {
	if b.Ref != "" {
		return registerRef(ctx, b.Ref, bindingsKey, nil), nil
	}

	res := render.Bindings{OriginalName: bindingsKey}

	for _, e := range b.ProtocolValues.Entries() {
		contents := make(map[string]any)
		switch e.Value.Selector {
		case 0:
			ctx.Logger.Trace("Bindings", "proto", e.Key, "format", "json")
			if err := json.Unmarshal(e.Value.V0, &contents); err != nil {
				return nil, types.CompileError{Path: ctx.CurrentRefPointer(), Proto: e.Key, Err: fmt.Errorf("json unmarshal: %w", err)}
			}
		case 1:
			ctx.Logger.Trace("Bindings", "proto", e.Key, "format", "yaml")
			if err := e.Value.V1.Decode(&contents); err != nil {
				return nil, types.CompileError{Path: ctx.CurrentRefPointer(), Proto: e.Key, Err: fmt.Errorf("yaml unmarshal: %w", err)}
			}
		default:
			panic(fmt.Errorf("invalid selector value %d, this is a bug", e.Value.Selector))
		}

		res.Values.Set(e.Key, contents)
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
