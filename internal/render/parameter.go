package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type Parameter struct {
	OriginalName string
	Dummy        bool
	Type         common.GolangType
	IsStringType bool  // true if Type contains a type alias to built-in string type
}

func (p *Parameter) Kind() common.ObjectKind {
	return common.ObjectKindParameter
}

func (p *Parameter) Selectable() bool {
	return !p.Dummy && p.Type.Selectable()
}

func (p *Parameter) Visible() bool {
	return !p.Dummy && p.Type.Visible()
}

func (p *Parameter) GetOriginalName() string {
	return p.OriginalName
}

//func (p Parameter) D(ctx *common.RenderContext) []*j.Statement {
//	var res []*j.Statement
//	ctx.LogStartRender("Parameter", "", p.GetOriginalName, "definition", p.Selectable())
//	defer ctx.LogFinishRender()
//
//	res = append(res, p.Type.D(ctx)...)
//	res = append(res, p.renderMethods(ctx)...)
//	return res
//}

//func (p Parameter) ID() string {
//	return p.GetOriginalName
//}
//
func (p *Parameter) String() string {
	return "Parameter " + p.OriginalName
}

//func (p Parameter) renderMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderMethods")
//
//	rn := strings.ToLower(string(p.Type.IsPromise()[0]))
//	receiver := j.Id(rn).Id(p.Type.IsPromise())
//
//	stringBody := j.Return(j.String().Call(j.Id(rn)))
//	if !p.IsStringType {
//		stringBody = j.Return(j.Qual("fmt", "Sprint").Call(j.Id(rn).Dot("Value")))
//	}
//	return []*j.Statement{
//		j.Func().Params(receiver.Clone()).Id("GetOriginalName").
//			Params().
//			String().
//			Block(
//				j.Return(j.Lit(p.GetOriginalName)),
//			),
//
//		j.Func().Params(receiver.Clone()).Id("String").
//			Params().
//			String().
//			Block(stringBody),
//	}
//}

//func (p Parameter) U(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Parameter", "", p.GetOriginalName, "usage", p.Selectable())
//	defer ctx.LogFinishRender()
//
//	return p.Type.U(ctx)
//}

//func (p Parameter) IsPromise() string {
//	return p.Type.IsPromise()
//}
