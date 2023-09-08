package assemble

import (
	"fmt"
	"path"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

type NewFunc struct {
	BaseType
	Struct             *Struct
	NewFuncArgs        []FuncParam
	NewFuncAllocFields []string // Struct field names to set in allocation expr, their order must follow the NewFuncArgs order
}

func (n NewFunc) AssembleDefinitionWithBody(ctx *common.AssembleContext, body []jen.Code) []*jen.Statement {
	funcArgs := lo.Map(n.NewFuncArgs, func(arg FuncParam, index int) jen.Code {
		return jen.Add(utils.ToJenCode(arg.assembleDefinition(ctx))...)
	})

	return []*jen.Statement{
		jen.Func().Id(n.Name).
			Params(funcArgs...).
			Params(jen.Op("*").Add(utils.ToJenCode(n.Struct.AssembleUsage(ctx))...)).
			Block(body...),
	}
}

func (n NewFunc) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	if len(n.NewFuncAllocFields) != len(n.NewFuncArgs) {
		panic(fmt.Sprintf("New func args count %d doesn't equal to struct fields count to set %d", len(n.NewFuncArgs), len(n.NewFuncAllocFields)))
	}

	allocFields := lo.SliceToMap(lo.Zip2(n.NewFuncAllocFields, n.NewFuncArgs), func(item lo.Tuple2[string, FuncParam]) (string, jen.Code) {
		return item.A, jen.Id(item.B.Name)
	})
	body := []jen.Code{jen.Return(jen.Op("&").Add(utils.ToJenCode(n.Struct.AssembleAllocation(ctx, allocFields))...))}
	return n.AssembleDefinitionWithBody(ctx, body)
}

func (n NewFunc) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if n.Package != "" && n.Package != ctx.CurrentPackage {
		return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(n.Package)), n.Name)}
	}
	return []*jen.Statement{jen.Id(n.Name)}
}

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	BaseType
	Fields []StructField

	// Render config
	Nullable bool
}

func (s Struct) CanBePointer() bool {
	return !s.Nullable
}

func (s Struct) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement

	if s.Description != "" {
		res = append(res, jen.Comment(s.Name+" -- "+utils.ToLowerFirstLetter(s.Description)))
	}

	code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})
	res = append(res, jen.Type().Id(s.Name).Struct(utils.ToJenCode(code)...))

	return res
}

func (s Struct) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	stmt := &jen.Statement{}
	if s.Nullable {
		stmt = stmt.Op("*")
	}
	if s.AllowRender() {
		if s.Package != "" && s.Package != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(s.Package)), s.Name)}
		}
		return []*jen.Statement{stmt.Id(s.Name)}
	}

	code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})

	return []*jen.Statement{stmt.Struct(utils.ToJenCode(code)...)}
}

func (s Struct) AssembleAllocation(ctx *common.AssembleContext, values map[string]jen.Code) []*jen.Statement {
	stmt := &jen.Statement{}
	if s.AllowRender() {
		if s.Package != "" && s.Package != ctx.CurrentPackage {
			stmt = stmt.Qual(path.Join(ctx.ImportBase, string(s.Package)), s.Name)
		} else {
			stmt = stmt.Id(s.Name)
		}
	} else {
		code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
			return item.assembleDefinition(ctx)
		})
		stmt = stmt.Struct(utils.ToJenCode(code)...)
	}

	vals := lo.MapKeys(values, func(value jen.Code, key string) jen.Code { return jen.Id(key) })
	stmt = stmt.Values(jen.Dict(vals))
	return []*jen.Statement{stmt}
}

func (s Struct) GetField(name string) (StructField, bool) {
	return lo.Find(s.Fields, func(item StructField) bool {
		return item.Name == name
	})
}

// StructField defines the data required to generate a field in Go.
type StructField struct {
	Name          string
	Description   string
	Type          common.GolangType
	RequiredValue bool // TODO: maybe create assemble.Pointer?
	Tags          map[string]string
}

func (f StructField) assembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if f.Description != "" {
		res = append(res, jen.Comment(f.Name+" -- "+utils.ToLowerFirstLetter(f.Description)))
	}

	stmt := jen.Id(f.Name)
	if f.Type.CanBePointer() && f.RequiredValue {
		stmt = stmt.Op("*")
	}
	items := utils.ToJenCode(f.Type.AssembleUsage(ctx))
	res = append(res, stmt.Add(items...))

	return res
}
