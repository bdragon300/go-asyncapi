package asyncapi

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type CorrelationID struct {
	Description string `json:"description,omitzero" yaml:"description"`
	Location    string `json:"location,omitzero" yaml:"location"`

	XIgnore bool `json:"x-ignore,omitzero" yaml:"x-ignore"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

func (c CorrelationID) Compile(ctx *compile.Context) error {
	obj, err := c.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (c CorrelationID) build(ctx *compile.Context, correlationIDKey string) (common.Artifact, error) {
	if c.XIgnore {
		ctx.Logger.Debug("CorrelationID denoted to be ignored")
		return &render.CorrelationID{Dummy: true}, nil
	}
	if c.Ref != "" {
		return registerRef(ctx, c.Ref, correlationIDKey, nil), nil
	}

	ctx.Logger.Trace("Parsing CorrelationID location runtime expression", "location", c.Location)
	structField, locationPath, err := parseRuntimeExpression(c.Location)
	if err != nil {
		return nil, types.CompileError{Err: fmt.Errorf("parse runtime expression: %w", err), Path: ctx.CurrentRefPointer()}
	}

	ctx.Logger.Trace("CorrelationID object", "structField", structField, "path", locationPath)
	return &render.CorrelationID{
		OriginalName: correlationIDKey,
		Description:  c.Description,
		BaseRuntimeExpression: lang.BaseRuntimeExpression{
			StructFieldKind: structField,
			LocationPath:    locationPath,
		},
	}, nil
}
