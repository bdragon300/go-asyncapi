package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoMap struct {
	BaseType
	KeyType   common.GolangType
	ValueType common.GolangType
}

func (m GoMap) D() string {
	//ctx.LogStartRender("GoMap", m.Import, m.Name, "definition", m.Selectable())
	//defer ctx.LogFinishRender()
	//
	//var res []*jen.Statement
	//if m.Description != "" {
	//	res = append(res, jen.Comment(m.Name+" -- "+utils.ToLowerFirstLetter(m.Description)))
	//}
	//
	//stmt := jen.Type().Id(m.Name)
	//keyType := utils.ToCode(m.KeyType.U())
	//valueType := utils.ToCode(m.ValueType.U())
	//res = append(res, stmt.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...))
	//
	//return res
	panic("not implemented")
}

func (m GoMap) U() string {
	//ctx.LogStartRender("GoMap", m.Import, m.Name, "usage", m.Selectable())
	//defer ctx.LogFinishRender()
	//
	//if m.HasDefinition {
	//	if m.Import != "" && m.Import != context.Context.CurrentPackage {
	//		return []*jen.Statement{jen.Qual(context.Context.GeneratedModule(m.Import), m.Name)}
	//	}
	//	return []*jen.Statement{jen.Id(m.Name)}
	//}
	//
	//keyType := utils.ToCode(m.KeyType.U())
	//valueType := utils.ToCode(m.ValueType.U())
	//return []*jen.Statement{jen.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...)}
	panic("not implemented")
}
