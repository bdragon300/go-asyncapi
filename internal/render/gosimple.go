package render

import (
	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
)

type Simple struct {
	Name            string            // type name
	IsIface         bool              // TODO: check if this field is filled correctly everywhere
	Package         string            // optional import path from any package
	TypeParamValues []common.Renderer // optional type parameter types to be filled in definition and usage
}

func (p Simple) DirectRendering() bool {
	return false
}

func (p Simple) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("SimpleType", p.Package, p.Name, "definition", p.DirectRendering())
	defer ctx.LogReturn()

	stmt := jen.Id(p.Name)
	if len(p.TypeParamValues) > 0 {
		typeParams := lo.FlatMap(p.TypeParamValues, func(item common.Renderer, index int) []jen.Code {
			return utils.ToCode(item.RenderUsage(ctx))
		})
		stmt = stmt.Types(typeParams...)
	}
	return []*jen.Statement{stmt}
}

func (p Simple) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("SimpleType", p.Package, p.Name, "usage", p.DirectRendering())
	defer ctx.LogReturn()

	stmt := &jen.Statement{}
	switch {
	case p.Package != "" && p.Package != ctx.CurrentPackage:
		stmt = stmt.Qual(p.Package, p.Name)
	default:
		stmt = stmt.Id(p.Name)
	}

	if len(p.TypeParamValues) > 0 {
		typeParams := lo.FlatMap(p.TypeParamValues, func(item common.Renderer, index int) []jen.Code {
			return utils.ToCode(item.RenderUsage(ctx))
		})
		stmt = stmt.Types(typeParams...)
	}

	return []*jen.Statement{stmt}
}

func (p Simple) TypeName() string {
	return p.Name
}

func (p Simple) String() string {
	return p.Name
}
