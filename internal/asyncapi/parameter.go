package asyncapi

import (
	"path"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Parameter struct {
	Description string  `json:"description" yaml:"description"`
	Schema      *Object `json:"schema" yaml:"schema"`     // TODO: implement
	Location    string  `json:"location" yaml:"location"` // TODO: implement

	XGoType *types.Union2[string, xGoType] `json:"x-go-type" yaml:"x-go-type"` // FIXME: what does it mean for parameter?
	XGoName string                         `json:"x-go-name" yaml:"x-go-name"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (p Parameter) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().PathItem)
	obj, err := p.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (p Parameter) build(ctx *common.CompileContext, parameterKey string) (common.Renderer, error) {
	ignore := !ctx.CompileOpts.ChannelOpts.Enable
	if ignore {
		ctx.Logger.Debug("Parameter denoted to be ignored along with all channels")
		return &render.Parameter{Dummy: true}, nil
	}
	if p.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", p.Ref)
		res := render.NewRendererPromise(p.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	if p.XGoType != nil {
		t := buildXGoType(p.XGoType)
		ctx.Logger.Trace("Parameter is a custom type", "type", t.String())
		return t, nil
	}

	parName, _ := lo.Coalesce(p.XGoName, parameterKey)
	res := &render.Parameter{Name: parName}

	if p.Schema != nil {
		ctx.Logger.Trace("Parameter schema")
		prm := render.NewGolangTypePromise(path.Join(ctx.PathRef(), "schema"), common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		res.Type = &render.GoStruct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(parName, ""),
				Description:  p.Description,
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
			Fields: []render.GoStructField{{Name: "Value", Type: prm}},
		}
	} else {
		ctx.Logger.Trace("Parameter has no schema")
		res.Type = &render.GoTypeAlias{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(parName, ""),
				Description:  p.Description,
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
			AliasedType: &render.GoSimple{Name: "string"},
		}
		res.PureString = true
	}

	return res, nil
}
