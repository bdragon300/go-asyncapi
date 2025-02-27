package asyncapi

import (
	"errors"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type CorrelationID struct {
	Description string `json:"description" yaml:"description"`
	Location    string `json:"location" yaml:"location"`

	XIgnore bool `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
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

	locationParts := strings.SplitN(c.Location, "#", 2)
	if len(locationParts) < 2 {
		return nil, types.CompileError{Err: errors.New("no fragment part in location"), Path: ctx.CurrentPositionRef()}
	}

	var structField render.CorrelationIDStructFieldKind
	switch {
	case strings.HasSuffix(locationParts[0], "header"):
		structField = render.CorrelationIDStructFieldKindHeaders
	case strings.HasSuffix(locationParts[0], "payload"):
		structField = render.CorrelationIDStructFieldKindPayload
	default:
		return nil, types.CompileError{
			Err:  errors.New("location source must point only to header or payload"),
			Path: ctx.CurrentPositionRef(),
		}
	}

	if !strings.HasPrefix(locationParts[1], "/") {
		return nil, types.CompileError{Err: errors.New("fragment part must start with a slash"), Path: ctx.CurrentPositionRef()}
	}
	if locationParts[1] == "/" {
		return nil, types.CompileError{Err: errors.New("location must not point to root of message/header"), Path: ctx.CurrentPositionRef()}
	}

	locationPath := strings.Split(locationParts[1], "/")[1:]
	ctx.Logger.Trace("CorrelationID object", "messageField", structField, "path", locationPath)

	return &render.CorrelationID{
		OriginalName:    correlationIDKey,
		Description:     c.Description,
		StructFieldKind: structField,
		LocationPath:    locationPath,
	}, nil
}
