package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Parameter struct {
	Enum 	  []string `json:"enum" yaml:"enum"`
	Default  string   `json:"default" yaml:"default"`
	Description string  `json:"description" yaml:"description"`
	Examples []string `json:"examples" yaml:"examples"`
	Schema      *Object `json:"schema" yaml:"schema"`     // DEPRECATED
	Location    string  `json:"location" yaml:"location"` // TODO: implement

	XGoName string `json:"x-go-name" yaml:"x-go-name"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (p Parameter) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj, err := p.build(ctx, ctx.Stack.Top().PathItem)
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
	res := &render.Parameter{OriginalName: parName}

	if p.Schema != nil {
		ctx.Logger.Trace("Parameter schema")
		prm := lang.NewGolangTypePromise(ctx.PathStackRef("schema"))
		ctx.PutPromise(prm)
		res.Type = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(parName, ""),
				Description:   p.Description,
				HasDefinition: true,
			},
			Fields: []lang.GoStructField{{Name: "Value", Type: prm}},
		}
	} else {
		ctx.Logger.Trace("Parameter without schema")
		res.Type = &lang.GoTypeAlias{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(parName, ""),
				Description:   p.Description,
				HasDefinition: true,
			},
			AliasedType: &lang.GoSimple{TypeName: "string"},
		}
		res.IsStringType = true
	}

	return res, nil
}
