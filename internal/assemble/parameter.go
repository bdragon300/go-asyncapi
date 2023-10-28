package assemble

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

func (p Parameter) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	res = append(res, p.Type.AssembleDefinition(ctx)...)
	res = append(res, p.assembleMethods()...)
	return res
}

func (p Parameter) assembleMethods() []*j.Statement {
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

func (p Parameter) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Type.AssembleUsage(ctx)
}

func (p Parameter) TypeName() string {
	return p.Type.TypeName()
}
