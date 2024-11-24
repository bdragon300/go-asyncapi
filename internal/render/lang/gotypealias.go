package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoTypeAlias struct {
	BaseType
	AliasedType common.GolangType
}

func (p GoTypeAlias) D() string {
	//ctx.LogStartRender("GoTypeAlias", p.Import, p.Name, "definition", p.Selectable())
	//defer ctx.LogFinishRender()
	//
	//var res []*jen.Statement
	//if p.Description != "" {
	//	res = append(res, jen.Comment(p.Name+" -- "+utils.ToLowerFirstLetter(p.Description)))
	//}
	//
	//aliasedStmt := utils.ToCode(p.AliasedType.D())
	//res = append(res, jen.Type().Id(p.Name).Add(aliasedStmt...))
	//return res
	panic("not implemented")
}

func (p GoTypeAlias) U() string {
	//ctx.LogStartRender("GoTypeAlias", p.Import, p.Name, "usage", p.IsDefinition())
	//defer ctx.LogFinishRender()
	//
	//if p.HasDefinition {
	//	if p.Import != "" && p.Import != context.Context.CurrentPackage {
	//		return []*jen.Statement{jen.Qual(context.Context.GeneratedModule(p.Import), p.Name)}
	//	}
	//	return []*jen.Statement{jen.Id(p.Name)}
	//}
	//
	//// This GoTypeAlias definition is not directly rendered anywhere, so it's name is unknown for the calling code.
	//// Just use the underlying type then
	//aliasedStmt := utils.ToCode(p.AliasedType.U())
	//return []*jen.Statement{jen.Add(aliasedStmt...)}
	panic("not implemented")
}

func (p GoTypeAlias) WrappedGolangType() (common.GolangType, bool) {
	return p.AliasedType, p.AliasedType != nil
}

func (p GoTypeAlias) IsPointer() bool {
	if v, ok := any(p.AliasedType).(GolangPointerType); ok {
		return v.IsPointer()
	}
	return false
}

func (p GoTypeAlias) IsStruct() bool {
	if v, ok := any(p.AliasedType).(golangStructType); ok {
		return v.IsStruct()
	}
	return false
}
