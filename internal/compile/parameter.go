package compile

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
)

type Parameter struct {
	Description string  `json:"description" yaml:"description"`
	Schema      *Object `json:"schema" yaml:"schema"`     // TODO: implement
	Location    string  `json:"location" yaml:"location"` // TODO: implement

	Ref string `json:"$ref" yaml:"$ref"`
}

func (p Parameter) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := p.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return err
	}
	ctx.PutToCurrentPkg(obj)
	return nil
}

func (p Parameter) build(ctx *common.CompileContext, parameterKey string) (common.Renderer, error) {
	if p.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", p.Ref)
		res := render.NewRefLinkAsRenderer(p.Ref, common.LinkOriginUser)
		ctx.Linker.Add(res)
		return res, nil
	}

	res := &render.Parameter{Name: parameterKey}

	if p.Schema != nil {
		ctx.Logger.Trace("Parameter schema")
		lnk := render.NewRefLinkAsGolangType(path.Join(ctx.PathRef(), "schema"), common.LinkOriginInternal)
		ctx.Linker.Add(lnk)
		res.Type = &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(parameterKey, ""),
				Description:  p.Description,
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
			Fields: []render.StructField{{Name: "Value", Type: lnk}},
		}
	} else {
		ctx.Logger.Trace("Parameter has no schema")
		res.Type = &render.TypeAlias{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(parameterKey, ""),
				Description:  p.Description,
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
			AliasedType: &render.Simple{Name: "string"},
		}
		res.PureString = true
	}

	return res, nil
}
