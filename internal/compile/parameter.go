package compile

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
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

func (p Parameter) build(ctx *common.CompileContext, parameterKey string) (common.Assembler, error) {
	if p.Ref != "" {
		ctx.LogDebug("Ref", "$ref", p.Ref)
		res := assemble.NewRefLinkAsAssembler(p.Ref, common.LinkOriginUser)
		ctx.Linker.Add(res)
		return res, nil
	}

	res := &assemble.Parameter{Name: parameterKey}

	if p.Schema != nil {
		ctx.LogDebug("Parameter schema")
		lnk := assemble.NewRefLinkAsGolangType(path.Join(ctx.PathRef(), "schema"), common.LinkOriginInternal)
		ctx.Linker.Add(lnk)
		res.Type = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(parameterKey, ""),
				Description: p.Description,
				Render:      true,
				PackageName: ctx.TopPackageName(),
			},
			Fields: []assemble.StructField{{Name: "Value", Type: lnk}},
		}
	} else {
		ctx.LogDebug("Parameter has no schema")
		res.Type = &assemble.TypeAlias{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(parameterKey, ""),
				Description: p.Description,
				Render:      true,
				PackageName: ctx.TopPackageName(),
			},
			AliasedType: &assemble.Simple{Name: "string"},
		}
		res.PureString = true
	}

	return res, nil
}
