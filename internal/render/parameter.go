package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type Parameter struct {
	Name       string
	Dummy      bool
	Type         common.GolangType
	IsStringType bool  // true if Type contains a type alias to built-in string type
}

func (p Parameter) Kind() common.ObjectKind {
	return common.ObjectKindParameter
}

func (p Parameter) Selectable() bool {
	return !p.Dummy && p.Type.Selectable()
}

//func (p Parameter) D(ctx *common.RenderContext) []*j.Statement {
//	var res []*j.Statement
//	ctx.LogStartRender("Parameter", "", p.Name, "definition", p.Selectable())
//	defer ctx.LogFinishRender()
//
//	res = append(res, p.Type.D(ctx)...)
//	res = append(res, p.renderMethods(ctx)...)
//	return res
//}

//func (p Parameter) ID() string {
//	return p.Name
//}
//
//func (p Parameter) String() string {
//	return "Parameter " + p.Name
//}

//func (p Parameter) renderMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderMethods")
//
//	rn := strings.ToLower(string(p.Type.TypeName()[0]))
//	receiver := j.Id(rn).Id(p.Type.TypeName())
//
//	stringBody := j.Return(j.String().Call(j.Id(rn)))
//	if !p.IsStringType {
//		stringBody = j.Return(j.Qual("fmt", "Sprint").Call(j.Id(rn).Dot("Value")))
//	}
//	return []*j.Statement{
//		j.Func().Params(receiver.Clone()).Id("Name").
//			Params().
//			String().
//			Block(
//				j.Return(j.Lit(p.Name)),
//			),
//
//		j.Func().Params(receiver.Clone()).Id("String").
//			Params().
//			String().
//			Block(stringBody),
//	}
//}

//func (p Parameter) U(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Parameter", "", p.Name, "usage", p.Selectable())
//	defer ctx.LogFinishRender()
//
//	return p.Type.U(ctx)
//}

//func (p Parameter) TypeName() string {
//	return p.Type.TypeName()
//}
