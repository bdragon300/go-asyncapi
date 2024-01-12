package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/dave/jennifer/jen"
)

type GoMap struct {
	BaseType
	KeyType   common.GolangType
	ValueType common.GolangType
}

func (m GoMap) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoMap", m.Import, m.Name, "definition", m.DirectRendering())
	defer ctx.LogReturn()

	var res []*jen.Statement
	if m.Description != "" {
		res = append(res, jen.Comment(m.Name+" -- "+utils.ToLowerFirstLetter(m.Description)))
	}

	stmt := jen.Type().Id(m.Name)
	keyType := utils.ToCode(m.KeyType.RenderUsage(ctx))
	valueType := utils.ToCode(m.ValueType.RenderUsage(ctx))
	res = append(res, stmt.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...))

	return res
}

func (m GoMap) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoMap", m.Import, m.Name, "usage", m.DirectRendering())
	defer ctx.LogReturn()

	if m.DirectRender {
		if m.Import != "" && m.Import != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedModule(m.Import), m.Name)}
		}
		return []*jen.Statement{jen.Id(m.Name)}
	}

	keyType := utils.ToCode(m.KeyType.RenderUsage(ctx))
	valueType := utils.ToCode(m.ValueType.RenderUsage(ctx))
	return []*jen.Statement{jen.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...)}
}
