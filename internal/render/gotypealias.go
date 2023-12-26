package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
)

type TypeAlias struct {
	BaseType
	AliasedType common.GolangType
}

func (p *TypeAlias) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("TypeAlias", p.PackageName, p.Name, "definition", p.DirectRendering())
	defer ctx.LogReturn()

	var res []*jen.Statement
	if p.Description != "" {
		res = append(res, jen.Comment(p.Name+" -- "+utils.ToLowerFirstLetter(p.Description)))
	}

	aliasedStmt := utils.ToCode(p.AliasedType.RenderDefinition(ctx))
	res = append(res, jen.Type().Id(p.Name).Add(aliasedStmt...))
	return res
}

func (p *TypeAlias) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("TypeAlias", p.PackageName, p.Name, "usage", p.DirectRendering())
	defer ctx.LogReturn()

	if p.DirectRender {
		if p.PackageName != "" && p.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedPackage(p.PackageName), p.Name)}
		}
		return []*jen.Statement{jen.Id(p.Name)}
	}

	// This TypeAlias definition is not directly rendered anywhere, so it's name is unknown for the calling code.
	// Just use the underlying type then
	aliasedStmt := utils.ToCode(p.AliasedType.RenderUsage(ctx))
	return []*jen.Statement{jen.Add(aliasedStmt...)}
}

func (p *TypeAlias) WrappedGolangType() (common.GolangType, bool) {
	return p.AliasedType, p.AliasedType != nil
}
