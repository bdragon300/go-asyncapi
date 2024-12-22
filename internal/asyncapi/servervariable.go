package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

type ServerVariable struct {
	Enum        []string `json:"enum" yaml:"enum"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description" yaml:"description"`
	Examples    []string `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (sv ServerVariable) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj, err := sv.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (sv ServerVariable) build(ctx *common.CompileContext, serverVariableKey string) (common.Renderable, error) {
	if sv.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", sv.Ref)
		res := lang.NewRef(sv.Ref, serverVariableKey, nil)
		ctx.PutPromise(res)
		return res, nil
	}

	res := &render.ServerVariable{
		OriginalName: serverVariableKey,
		Enum:         sv.Enum,
		Default:      sv.Default,
		Description:  sv.Description,
	}

	return res, nil
}
