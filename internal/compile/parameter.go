package compile

import (
	"fmt"
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
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
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.PutToCurrentPkg(obj)
	return nil
}

func (p Parameter) build(ctx *common.CompileContext, parameterKey string) (common.Assembler, error) {
	if p.Ref != "" {
		res := assemble.NewRefLinkAsAssembler(p.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	res := &assemble.Parameter{Name: parameterKey}

	if p.Schema != nil {
		lnk := assemble.NewRefLinkAsGolangType(path.Join(ctx.PathRef(), "schema"))
		ctx.Linker.Add(lnk)
		res.Type = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        utils.ToGolangName(parameterKey, true),
				Description: p.Description,
				Render:      true,
				PackageName: ctx.TopPackageName(),
			},
			Fields: []assemble.StructField{{Name: "Value", Type: lnk}},
		}
	} else {
		res.Type = &assemble.TypeAlias{
			BaseType: assemble.BaseType{
				Name:        utils.ToGolangName(parameterKey, true),
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
