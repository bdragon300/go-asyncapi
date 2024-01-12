package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/dave/jennifer/jen"
)

type GoArray struct {
	BaseType
	ItemsType common.GolangType
	Size      int
}

func (a GoArray) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoArray", a.Import, a.Name, "definition", a.DirectRendering())
	defer ctx.LogReturn()

	var res []*jen.Statement
	if a.Description != "" {
		res = append(res, jen.Comment(a.Name+" -- "+utils.ToLowerFirstLetter(a.Description)))
	}

	stmt := jen.Type().Id(a.Name)
	if a.Size > 0 {
		stmt = stmt.Index(jen.Lit(a.Size))
	} else {
		stmt = stmt.Index()
	}
	items := utils.ToCode(a.ItemsType.RenderUsage(ctx))
	res = append(res, stmt.Add(items...))

	return res
}

func (a GoArray) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoArray", a.Import, a.Name, "usage", a.DirectRendering())
	defer ctx.LogReturn()

	if a.DirectRender {
		if a.Import != "" && a.Import != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedModule(a.Import), a.Name)}
		}
		return []*jen.Statement{jen.Id(a.Name)}
	}

	items := utils.ToCode(a.ItemsType.RenderUsage(ctx))
	if a.Size > 0 {
		return []*jen.Statement{jen.Index(jen.Lit(a.Size)).Add(items...)}
	}

	return []*jen.Statement{jen.Index().Add(items...)}
}
