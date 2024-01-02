package asyncapi

import (
	"errors"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
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
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := c.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (c CorrelationID) build(ctx *common.CompileContext, correlationIDKey string) (common.Renderer, error) {
	if c.XIgnore {
		ctx.Logger.Debug("CorrelationID denoted to be ignored")
		return &render.GoSimple{Name: "any", IsIface: true}, nil
	}
	// TODO: move this ref code from everywhere to single place?
	if c.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", c.Ref)
		res := render.NewRendererPromise(c.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	parts := strings.SplitN(c.Location, "#", 2)
	if len(parts) < 2 {
		return nil, types.CompileError{Err: errors.New("no fragment part in location"), Path: ctx.PathRef()}
	}

	var structField string
	switch {
	case strings.HasSuffix(parts[0], "header"):
		structField = "Headers"
	case strings.HasSuffix(parts[0], "payload"):
		structField = "Payload"
	default:
		return nil, types.CompileError{
			Err:  errors.New("location source must point only to header or payload"),
			Path: ctx.PathRef(),
		}
	}

	if !strings.HasPrefix(parts[1], "/") {
		return nil, types.CompileError{Err: errors.New("fragment part must start with a slash"), Path: ctx.PathRef()}
	}
	if parts[1] == "/" {
		return nil, types.CompileError{Err: errors.New("location must not point to root of message/header"), Path: ctx.PathRef()}
	}

	pathParts := strings.Split(parts[1], "/")[1:] // TODO: implement rfc6901 symbols encoding
	ctx.Logger.Trace("CorrelationID object", "messageField", structField, "path", pathParts)

	return &render.CorrelationID{
		Name:        correlationIDKey,
		Description: c.Description,
		StructField: structField,
		Path:        pathParts,
	}, nil
}
