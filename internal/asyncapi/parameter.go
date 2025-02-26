package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Parameter struct {
	Enum        []string `json:"enum" yaml:"enum"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description" yaml:"description"`
	Examples    []string `json:"examples" yaml:"examples"`
	Location    string   `json:"location" yaml:"location"` // TODO: implement

	XGoName string `json:"x-go-name" yaml:"x-go-name"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (p Parameter) Compile(ctx *common.CompileContext) error {
	obj, err := p.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (p Parameter) build(ctx *common.CompileContext, parameterKey string) (common.Renderable, error) {
	if p.Ref != "" {
		return registerRef(ctx, p.Ref, parameterKey, nil), nil
	}

	parName, _ := lo.Coalesce(p.XGoName, parameterKey)
	res := &render.Parameter{
		OriginalName: parName,
		Type: &lang.GoTypeDefinition{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(parName, ""),
				Description:   p.Description,
				HasDefinition: true,
			},
			RedefinedType: &lang.GoSimple{TypeName: "string"},
		},
	}

	return res, nil
}
