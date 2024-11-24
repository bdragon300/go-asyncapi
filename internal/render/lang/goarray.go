package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
)

type GoArray struct {
	BaseType
	ItemsType common.GolangType
	Size      int
}

func (a GoArray) D() string {
	//ctx.LogStartRender("GoArray", a.Import, a.Name, "definition", a.Selectable())
	//defer ctx.LogFinishRender()
	//
	//var b strings.Builder
	//if a.Description != "" {
	//	res = append(res, jen.Comment(a.Name+" -- "+utils.ToLowerFirstLetter(a.Description)))
	//}
	//
	//stmt := jen.Type().Id(a.Name)
	//if a.Size > 0 {
	//	stmt = stmt.Index(jen.Lit(a.Size))
	//} else {
	//	stmt = stmt.Index()
	//}
	//items := utils.ToCode(a.ItemsType.U())
	//res = append(res, stmt.Add(items...))
	//
	//return res
	panic("not implemented")
}

func (a GoArray) U() string {
	//ctx.LogStartRender("GoArray", a.Import, a.Name, "usage", a.Selectable())
	//defer ctx.LogFinishRender()
	//
	//if a.HasDefinition {
	//	if a.Import != "" && a.Import != context.Context.CurrentPackage {
	//		return []*jen.Statement{jen.Qual(context.Context.GeneratedModule(a.Import), a.Name)}
	//	}
	//	return []*jen.Statement{jen.Id(a.Name)}
	//}
	//
	//items := utils.ToCode(a.ItemsType.U())
	//if a.Size > 0 {
	//	return []*jen.Statement{jen.Index(jen.Lit(a.Size)).Add(items...)}
	//}
	//
	//return []*jen.Statement{jen.Index().Add(items...)}
	panic("not implemented")
}
