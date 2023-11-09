package render

import (
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	j "github.com/dave/jennifer/jen"
)

type Parameter struct {
	Name       string
	Type       common.GolangType
	PureString bool
}

func (p Parameter) AllowRender() bool {
	return p.Type.AllowRender()
}

func (p Parameter) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	res = append(res, p.Type.RenderDefinition(ctx)...)
	res = append(res, p.renderMethods()...)
	return res
}

func (p Parameter) String() string {
	return "Parameter " + p.Name
}

func (p Parameter) renderMethods() []*j.Statement {
	rn := strings.ToLower(string(p.Type.TypeName()[0]))
	receiver := j.Id(rn).Id(p.Type.TypeName())

	stringBody := j.Return(j.String().Call(j.Id(rn)))
	if !p.PureString {
		stringBody = j.Return(j.Qual("fmt", "Sprint").Call(j.Id(rn).Dot("Value")))
	}
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Name)),
			),

		j.Func().Params(receiver.Clone()).Id("String").
			Params().
			String().
			Block(stringBody),
	}
}

func (p Parameter) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	return p.Type.RenderUsage(ctx)
}

func (p Parameter) TypeName() string {
	return p.Type.TypeName()
}