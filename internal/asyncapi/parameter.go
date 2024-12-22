package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Parameter struct {
	Description string  `json:"description" yaml:"description"`
	Schema      *Object `json:"schema" yaml:"schema"`     // TODO: implement
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
	//ignore := !ctx.CompileOpts.ChannelOpts.Enable
	//if ignore {
	//	ctx.Logger.Debug("Parameter denoted to be ignored along with all channels")
	//	return &render.Parameter{Dummy: true}, nil
	//}
	if p.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", p.Ref)
		res := lang.NewUserPromise(p.Ref, parameterKey, nil)
		ctx.PutPromise(res)
		return res, nil
	}

	parName, _ := lo.Coalesce(p.XGoName, parameterKey)
	res := &render.Parameter{OriginalName: parName}

	if p.Schema != nil {
		ctx.Logger.Trace("Parameter schema")
		prm := lang.NewInternalGolangTypePromise(ctx.PathStackRef("schema"))
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
		ctx.Logger.Trace("Parameter has no schema")
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
