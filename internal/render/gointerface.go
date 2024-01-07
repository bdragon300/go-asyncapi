package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type GoInterface struct {
	BaseType
	Methods []GoFuncSignature
}

func (i GoInterface) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	ctx.LogRender("GoInterface", i.PackageName, i.Name, "definition", i.DirectRendering())
	defer ctx.LogReturn()

	if i.Description != "" {
		res = append(res, jen.Comment(i.Name+" -- "+utils.ToLowerFirstLetter(i.Description)))
	}

	code := lo.FlatMap(i.Methods, func(item GoFuncSignature, index int) []*jen.Statement {
		return item.RenderDefinition(ctx)
	})
	res = append(res, jen.Type().Id(i.Name).Interface(utils.ToCode(code)...))
	return res
}

func (i GoInterface) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoInterface", i.PackageName, i.Name, "usage", i.DirectRendering())
	defer ctx.LogReturn()

	if i.DirectRendering() {
		if i.PackageName != "" && i.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedModule(i.PackageName), i.Name)}
		}
		return []*jen.Statement{jen.Id(i.Name)}
	}

	code := lo.FlatMap(i.Methods, func(item GoFuncSignature, index int) []*jen.Statement {
		return item.RenderDefinition(ctx)
	})
	return []*jen.Statement{
		jen.Interface(utils.ToCode(code)...),
	}
}

