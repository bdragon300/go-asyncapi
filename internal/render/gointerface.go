package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type GoInterface struct {
	BaseType
	Methods []GoFuncSignature
}

func (i GoInterface) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	ctx.LogStartRender("GoInterface", i.Import, i.Name, "definition", i.DirectRendering())
	defer ctx.LogFinishRender()

	if i.Description != "" {
		res = append(res, jen.Comment(i.Name+" -- "+utils.ToLowerFirstLetter(i.Description)))
	}

	code := lo.FlatMap(i.Methods, func(item GoFuncSignature, _ int) []*jen.Statement {
		return item.RenderDefinition(ctx)
	})
	res = append(res, jen.Type().Id(i.Name).Interface(utils.ToCode(code)...))
	return res
}

func (i GoInterface) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogStartRender("GoInterface", i.Import, i.Name, "usage", i.DirectRendering())
	defer ctx.LogFinishRender()

	if i.DirectRendering() {
		if i.Import != "" && i.Import != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedModule(i.Import), i.Name)}
		}
		return []*jen.Statement{jen.Id(i.Name)}
	}

	code := lo.FlatMap(i.Methods, func(item GoFuncSignature, _ int) []*jen.Statement {
		return item.RenderDefinition(ctx)
	})
	return []*jen.Statement{
		jen.Interface(utils.ToCode(code)...),
	}
}

