package render

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type GoFuncSignature struct {
	Name   string
	Args   []GoFuncParam
	Return []GoFuncParam
}

func (f GoFuncSignature) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoFuncSignature", "", f.Name, "definition", false)
	defer ctx.LogReturn()

	stmt := jen.Id(f.Name)
	code := lo.FlatMap(f.Args, func(item GoFuncParam, index int) []*jen.Statement {
		return item.renderDefinition(ctx)
	})
	stmt = stmt.Params(utils.ToCode(code)...)
	code = lo.FlatMap(f.Return, func(item GoFuncParam, index int) []*jen.Statement {
		return item.renderDefinition(ctx)
	})
	if len(code) > 1 {
		stmt = stmt.Params(utils.ToCode(code)...)
	} else {
		stmt = stmt.Add(utils.ToCode(code)...)
	}
	return []*jen.Statement{stmt}
}

func (f GoFuncSignature) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoFuncSignature", "", f.Name, "usage", false)
	defer ctx.LogReturn()
	return []*jen.Statement{jen.Id(f.Name)}
}

func (f GoFuncSignature) ID() string {
	return f.Name
}

func (f GoFuncSignature) DirectRendering() bool {
	return false
}

func (f GoFuncSignature) TypeName() string {
	return f.Name
}

func (f GoFuncSignature) String() string {
	ret := ""
	switch {
	case len(f.Return) == 1:
		ret = f.Return[0].String()
	case len(f.Return) > 1:
		ret = strings.Join(lo.Map(f.Return, func(item GoFuncParam, _ int) string { return item.String() }), ", ")
	}
	args := strings.Join(lo.Map(f.Return, func(item GoFuncParam, _ int) string { return item.String() }), ", ")
	return fmt.Sprintf("GoFuncSignature %s(%s)%s", f.Name, args, ret)
}

type GoFuncParam struct {
	Name     string
	Type     common.GolangType
	Pointer  bool
	Variadic bool
}

func (n GoFuncParam) renderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("GoFuncParam", "", n.Name, "definition", false)
	defer ctx.LogReturn()

	stmt := &jen.Statement{}
	if n.Name != "" {
		stmt = stmt.Id(n.Name)
	}
	if n.Variadic {
		stmt = stmt.Op("...")
	}
	if n.Pointer {
		stmt = stmt.Op("*")
	}
	stmt = stmt.Add(utils.ToCode(n.Type.RenderUsage(ctx))...)
	return []*jen.Statement{stmt}
}

func (n GoFuncParam) String() string {
	return "GoFuncParam " + n.Name + " of " + n.Type.String()
}
