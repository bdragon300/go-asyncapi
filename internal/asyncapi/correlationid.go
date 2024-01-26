package asyncapi

import (
	"errors"
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
	ctx.SetTopObjName(ctx.Stack.Top().PathItem)
	obj, err := c.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (c CorrelationID) build(ctx *common.CompileContext, correlationIDKey string) (common.Renderer, error) {
	ignore := c.XIgnore || !ctx.CompileOpts.MessageOpts.Enable
	if ignore {
		ctx.Logger.Debug("CorrelationID denoted to be ignored")
		return &render.CorrelationID{}, nil
	}
	// TODO: move this ref code from everywhere to single place?
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		res := render.NewRendererPromise(c.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	locationParts := strings.SplitN(c.Location, "#", 2)
	if len(locationParts) < 2 {
		return nil, types.CompileError{Err: errors.New("no fragment part in location"), Path: ctx.PathRef()}
	}

	var structField string
	switch {
	case strings.HasSuffix(locationParts[0], "header"):
		structField = "Headers"
	case strings.HasSuffix(locationParts[0], "payload"):
		structField = "Payload"
	default:
		return nil, types.CompileError{
			Err:  errors.New("location source must point only to header or payload"),
			Path: ctx.PathRef(),
		}
	}

	if !strings.HasPrefix(locationParts[1], "/") {
		return nil, types.CompileError{Err: errors.New("fragment part must start with a slash"), Path: ctx.PathRef()}
	}
	if locationParts[1] == "/" {
		return nil, types.CompileError{Err: errors.New("location must not point to root of message/header"), Path: ctx.PathRef()}
	}

	locationPath := strings.Split(locationParts[1], "/")[1:] // TODO: implement rfc6901 symbols encoding
	ctx.Logger.Trace("CorrelationID object", "messageField", structField, "path", locationPath)

	return &render.CorrelationID{
		Name:         correlationIDKey,
		Description:  c.Description,
		StructField:  structField,
		LocationPath: locationPath,
	}, nil
}
