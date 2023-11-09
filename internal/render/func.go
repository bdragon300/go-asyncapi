package render

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type FuncSignature struct {
	Name   string
	Args   []FuncParam
	Return []FuncParam
}

func (f FuncSignature) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	stmt := jen.Id(f.Name)
	code := lo.FlatMap(f.Args, func(item FuncParam, index int) []*jen.Statement {
		return item.renderDefinition(ctx)
	})
	stmt = stmt.Params(utils.ToCode(code)...)
	code = lo.FlatMap(f.Return, func(item FuncParam, index int) []*jen.Statement {
		return item.renderDefinition(ctx)
	})
	if len(code) > 1 {
		stmt = stmt.Params(utils.ToCode(code)...)
	} else {
		stmt = stmt.Add(utils.ToCode(code)...)
	}
	return []*jen.Statement{stmt}
}

func (f FuncSignature) RenderUsage(_ *common.RenderContext) []*jen.Statement {
	return []*jen.Statement{jen.Id(f.Name)}
}

func (f FuncSignature) String() string {
	ret := ""
	switch {
	case len(f.Return) == 1:
		ret = f.Return[0].String()
	case len(f.Return) > 1:
		ret = strings.Join(lo.Map(f.Return, func(item FuncParam, _ int) string { return item.String() }), ", ")
	}
	args := strings.Join(lo.Map(f.Return, func(item FuncParam, _ int) string { return item.String() }), ", ")
	return fmt.Sprintf("%s(%s)%s", f.Name, args, ret)
}

type Func struct {
	FuncSignature
	Receiver        common.GolangType
	PointerReceiver bool
	PackageName     string // optional import path from any generated package
	BodyRenderer    func(ctx *common.RenderContext, f *Func) []*jen.Statement
}

func (f Func) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	stmt := jen.Func()
	if f.Receiver != nil {
		r := jen.Id(f.ReceiverName())
		if f.PointerReceiver {
			r = r.Op("*")
		}
		r = r.Add(utils.ToCode(f.Receiver.RenderUsage(ctx))...)
		stmt = stmt.Params(r)
	} else {
		stmt = stmt.Func()
	}
	stmt = stmt.Add(utils.ToCode(f.FuncSignature.RenderDefinition(ctx))...)
	if f.BodyRenderer != nil {
		stmt = stmt.Block(utils.ToCode(f.BodyRenderer(ctx, &f))...)
	}
	return []*jen.Statement{stmt}
}

func (f Func) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	if f.PackageName != "" && f.PackageName != ctx.CurrentPackage && f.Receiver == nil {
		return []*jen.Statement{jen.Qual(ctx.GeneratedPackage(f.PackageName), f.Name)}
	}
	return []*jen.Statement{jen.Id(f.Name)}
}

func (f Func) AllowRender() bool {
	return true
}

func (f Func) String() string {
	return fmt.Sprintf("(%s) %s", f.Receiver.String(), f.FuncSignature.String())
}

func (f Func) ReceiverName() string {
	if f.Receiver == nil {
		panic("receiver has not set")
	}
	return strings.ToLower(string(f.Receiver.TypeName()[0]))
}

type FuncParam struct {
	Name     string
	Type     common.GolangType
	Pointer  bool
	Variadic bool
}

func (n FuncParam) renderDefinition(ctx *common.RenderContext) []*jen.Statement {
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

func (n FuncParam) String() string {
	return n.Type.String() + " " + n.Type.String()
}