package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoTypeAlias struct {
	BaseType
	AliasedType common.GolangType
}

func (p *GoTypeAlias) D() string {
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
	p.SetDefinitionInfo(common.GetContext().CurrentDefinitionInfo())
	return renderTemplate("lang/gotypealias/definition", p)
}

func (p *GoTypeAlias) U() string {
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
	return renderTemplate("lang/gotypealias/usage", p)
}

func (p *GoTypeAlias) UnwrapGolangType() (common.GolangType, bool) {
	if v, ok := p.AliasedType.(GolangTypeWrapperType); ok {
		return v.UnwrapGolangType()
	}
	return p.AliasedType, p.AliasedType != nil
}

func (p *GoTypeAlias) IsPointer() bool {
	if v, ok := any(p.AliasedType).(GolangPointerType); ok {
		return v.IsPointer()
	}
	return false
}

func (p *GoTypeAlias) IsStruct() bool {
	if v, ok := any(p.AliasedType).(golangStructType); ok {
		return v.IsStruct()
	}
	return false
}

func (p *GoTypeAlias) String() string {
	if p.Import != "" {
		return "GoTypeAlias /" + p.Import + "." + p.Name
	}
	return "GoTypeAlias " + p.Name
}