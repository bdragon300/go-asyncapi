package asyncapi

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type CorrelationID struct {
	Description string `json:"description" yaml:"description"`
	Location    string `json:"location" yaml:"location"`

	// Not used cause the object is not rendered

	XIgnore bool `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c CorrelationID) Compile(ctx *common.CompileContext) error {
	// TODO: move this code from everywhere to single place?
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj, err := c.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (c CorrelationID) build(ctx *common.CompileContext, correlationIDKey string) (common.Renderable, error) {
	ignore := c.XIgnore //|| !ctx.CompileOpts.MessageOpts.Enable
	if ignore {
		ctx.Logger.Debug("CorrelationID denoted to be ignored")
		return &render.CorrelationID{}, nil
	}
	// TODO: move this ref code from everywhere to single place?
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		res := lang.NewUserPromise(c.Ref, correlationIDKey, nil)
		ctx.PutPromise(res)
		return res, nil
	}

	locationParts := strings.SplitN(c.Location, "#", 2)
	if len(locationParts) < 2 {
		return nil, types.CompileError{Err: errors.New("no fragment part in location"), Path: ctx.PathStackRef()}
	}

	var structField render.CorrelationIDStructField
	switch {
	case strings.HasSuffix(locationParts[0], "header"):
		structField = render.CorrelationIDStructFieldHeaders
	case strings.HasSuffix(locationParts[0], "payload"):
		structField = render.CorrelationIDStructFieldPayload
	default:
		return nil, types.CompileError{
			Err:  errors.New("location source must point only to header or payload"),
			Path: ctx.PathStackRef(),
		}
	}

	if !strings.HasPrefix(locationParts[1], "/") {
		return nil, types.CompileError{Err: errors.New("fragment part must start with a slash"), Path: ctx.PathStackRef()}
	}
	if locationParts[1] == "/" {
		return nil, types.CompileError{Err: errors.New("location must not point to root of message/header"), Path: ctx.PathStackRef()}
	}

	locationPath := strings.Split(locationParts[1], "/")[1:]
	ctx.Logger.Trace("CorrelationID object", "messageField", structField, "path", locationPath)

	return &render.CorrelationID{
		OriginalName: correlationIDKey,
		Description:  c.Description,
		StructField:  structField,
		LocationPath: locationPath,
	}, nil
}
