package assemble

import (
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type FuncSignature struct {
	Name   string
	Args   []FuncParam
	Return []FuncParam
}

func (f FuncSignature) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	stmt := jen.Id(f.Name)
	code := lo.FlatMap(f.Args, func(item FuncParam, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	stmt = stmt.Params(utils.ToCode(code)...)
	code = lo.FlatMap(f.Return, func(item FuncParam, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	if len(code) > 1 {
		stmt = stmt.Params(utils.ToCode(code)...)
	} else {
		stmt = stmt.Add(utils.ToCode(code)...)
	}
	return []*jen.Statement{stmt}
}

func (f FuncSignature) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	return []*jen.Statement{jen.Id(f.Name)}
}

type Func struct {
	FuncSignature
	Receiver        common.GolangType
	PointerReceiver bool
	Package         string // optional import path from any generated package
	BodyAssembler   func(ctx *common.AssembleContext, f *Func) []*jen.Statement
}

func (f Func) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	stmt := jen.Func()
	if f.Receiver != nil {
		r := jen.Id(f.ReceiverName())
		if f.PointerReceiver {
			r = r.Op("*")
		}
		r = r.Add(utils.ToCode(f.Receiver.AssembleUsage(ctx))...)
		stmt = stmt.Params(r)
	} else {
		stmt = stmt.Func()
	}
	stmt = stmt.Add(utils.ToCode(f.FuncSignature.AssembleDefinition(ctx))...)
	if f.BodyAssembler != nil {
		stmt = stmt.Block(utils.ToCode(f.BodyAssembler(ctx, &f))...)
	}
	return []*jen.Statement{stmt}
}

func (f Func) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if f.Package != "" && f.Package != ctx.CurrentPackage && f.Receiver == nil {
		return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, f.Package), f.Name)}
	}
	return []*jen.Statement{jen.Id(f.Name)}
}

func (f Func) AllowRender() bool {
	return true
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

func (n FuncParam) assembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
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
	stmt = stmt.Add(utils.ToCode(n.Type.AssembleUsage(ctx))...)
	return []*jen.Statement{stmt}
}
