package assemble

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Interface struct {
	BaseType
	Methods []FunctionSignature
}

func (i Interface) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if i.Description != "" {
		res = append(res, jen.Comment(i.Name+" -- "+utils.ToLowerFirstLetter(i.Description)))
	}

	code := lo.FlatMap(i.Methods, func(item FunctionSignature, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	res = append(res, jen.Type().Id(i.Name).Interface(utils.ToJenCode(code)...))
	return res
}

func (i Interface) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if i.AllowRender() {
		if i.Package != "" && i.Package != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(i.Package)), i.Name)}
		}
		return []*jen.Statement{jen.Id(i.Name)}
	}

	code := lo.FlatMap(i.Methods, func(item FunctionSignature, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	return []*jen.Statement{
		jen.Interface(utils.ToJenCode(code)...),
	}
}

func (i Interface) CanBePointer() bool {
	return false
}

type FunctionSignature struct {
	Name   string
	Args   []FuncParam
	Return []FuncParam
}

func (i FunctionSignature) assembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	stmt := jen.Id(i.Name)
	code := lo.FlatMap(i.Args, func(item FuncParam, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	stmt = stmt.Params(utils.ToJenCode(code)...)
	code = lo.FlatMap(i.Return, func(item FuncParam, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	if len(code) > 1 {
		stmt = stmt.Params(utils.ToJenCode(code)...)
	} else {
		stmt = stmt.Add(utils.ToJenCode(code)...)
	}
	return []*jen.Statement{stmt}
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
	stmt = stmt.Add(utils.ToJenCode(n.Type.AssembleUsage(ctx))...)
	return []*jen.Statement{stmt}
}
