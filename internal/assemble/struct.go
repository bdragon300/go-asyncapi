package assemble

import (
	"fmt"
	"path"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	BaseType
	Fields []StructField
}

func (s Struct) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement

	if s.Description != "" {
		res = append(res, jen.Comment(s.Name+" -- "+utils.ToLowerFirstLetter(s.Description)))
	}

	code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	res = append(res, jen.Type().Id(s.Name).Struct(utils.ToCode(code)...))

	return res
}

func (s Struct) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if s.AllowRender() {
		if s.Package != "" && s.Package != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(s.Package)), s.Name)}
		}
		return []*jen.Statement{jen.Id(s.Name)}
	}

	code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})

	return []*jen.Statement{jen.Struct(utils.ToCode(code)...)}
}

func (s Struct) NewFuncUsage(ctx *common.AssembleContext) []*jen.Statement {
	if s.Package != "" && s.Package != ctx.CurrentPackage {
		return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(s.Package)), s.NewFuncName())}
	}
	return []*jen.Statement{jen.Id(s.NewFuncName())}
}

func (s Struct) NewFuncName() string {
	return "New" + s.Name
}

func (s Struct) MustGetField(name string) StructField {
	f, ok := lo.Find(s.Fields, func(item StructField) bool {
		return item.Name == name
	})
	if !ok {
		panic(fmt.Sprintf("Field %s.%s not found", s.Name, name))
	}
	return f
}

// StructField defines the data required to generate a field in Go.
type StructField struct {
	Name         string
	Description  string
	Type         common.GolangType
	ForcePointer bool
	Tags         map[string]string
}

func (f StructField) assembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if f.Description != "" {
		res = append(res, jen.Comment(f.Name+" -- "+utils.ToLowerFirstLetter(f.Description)))
	}

	stmt := jen.Id(f.Name)

	drawPtr := f.ForcePointer
	if _, ok := f.Type.(*NullableType); ok {
		drawPtr = true // Nullable fields require to be pointer on any case
	}
	if drawPtr {
		stmt = stmt.Op("*")
	}
	items := utils.ToCode(f.Type.AssembleUsage(ctx))
	res = append(res, stmt.Add(items...))

	return res
}
