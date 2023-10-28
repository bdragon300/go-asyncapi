package assemble

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Interface struct {
	BaseType
	Methods []FuncSignature
}

func (i Interface) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if i.Description != "" {
		res = append(res, jen.Comment(i.Name+" -- "+utils.ToLowerFirstLetter(i.Description)))
	}

	code := lo.FlatMap(i.Methods, func(item FuncSignature, index int) []*jen.Statement {
		return item.AssembleDefinition(ctx)
	})
	res = append(res, jen.Type().Id(i.Name).Interface(utils.ToCode(code)...))
	return res
}

func (i Interface) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if i.AllowRender() {
		if i.Package != "" && i.Package != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(i.Package)), i.Name)}
		}
		return []*jen.Statement{jen.Id(i.Name)}
	}

	code := lo.FlatMap(i.Methods, func(item FuncSignature, index int) []*jen.Statement {
		return item.AssembleDefinition(ctx)
	})
	return []*jen.Statement{
		jen.Interface(utils.ToCode(code)...),
	}
}

