package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
)

type Array struct {
	BaseType
	ItemsType common.GolangType
	Size      int
}

func (a *Array) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("Array", a.PackageName, a.Name, "definition", a.DirectRendering())
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

func (a *Array) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("Array", a.PackageName, a.Name, "usage", a.DirectRendering())
	defer ctx.LogReturn()

	if a.DirectRender {
		if a.PackageName != "" && a.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedPackage(a.PackageName), a.Name)}
		}
		return []*jen.Statement{jen.Id(a.Name)}
	}

	items := utils.ToCode(a.ItemsType.RenderUsage(ctx))
	if a.Size > 0 {
		return []*jen.Statement{jen.Index(jen.Lit(a.Size)).Add(items...)}
	}

	return []*jen.Statement{jen.Index().Add(items...)}
}

func (a *Array) IsCollection() bool {
	return true
}
