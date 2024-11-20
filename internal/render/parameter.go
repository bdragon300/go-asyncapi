package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type Parameter struct {
	Name       string
	Dummy      bool
	Type       common.GolangType
	PureString bool
}

func (p Parameter) DirectRendering() bool {
	return !p.Dummy && p.Type.DirectRendering()
}

//func (p Parameter) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
//	var res []*j.Statement
//	ctx.LogStartRender("Parameter", "", p.Name, "definition", p.DirectRendering())
//	defer ctx.LogFinishRender()
//
//	res = append(res, p.Type.RenderDefinition(ctx)...)
//	res = append(res, p.renderMethods(ctx)...)
//	return res
//}

func (p Parameter) ID() string {
	return p.Name
}

func (p Parameter) String() string {
	return "Parameter " + p.Name
}

//func (p Parameter) renderMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderMethods")
//
//	rn := strings.ToLower(string(p.Type.TypeName()[0]))
//	receiver := j.Id(rn).Id(p.Type.TypeName())
//
//	stringBody := j.Return(j.String().Call(j.Id(rn)))
//	if !p.PureString {
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

//func (p Parameter) RenderUsage(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Parameter", "", p.Name, "usage", p.DirectRendering())
//	defer ctx.LogFinishRender()
//
//	return p.Type.RenderUsage(ctx)
//}

func (p Parameter) TypeName() string {
	return p.Type.TypeName()
}
