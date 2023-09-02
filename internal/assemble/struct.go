package assemble

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	BaseType
	Fields []StructField

	// Render config
	Nullable bool
}

func (s *Struct) CanBePointer() bool {
	return !s.Nullable
}

func (s *Struct) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if s.Description != "" {
		res = append(res, jen.Comment(s.Name+" -- "+utils.ToLowerFirstLetter(s.Description)))
	}

	var code []jen.Code
	for _, f := range s.Fields {
		items := utils.CastSliceItems[*jen.Statement, jen.Code](f.assembleDefinition(ctx))
		code = append(code, items...)
	}
	res = append(res, jen.Type().Id(s.Name).Struct(code...))

	return res
}

func (s *Struct) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
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

	var code []jen.Code
	for _, f := range s.Fields {
		items := utils.CastSliceItems[*jen.Statement, jen.Code](f.assembleDefinition(ctx))
		code = append(code, items...)
	}

	return []*jen.Statement{stmt.Struct(code...)}
}

// StructField defines the data required to generate a field in Go.
type StructField struct {
	Name          string
	Description   string
	Type          common.GolangType
	RequiredValue bool // TODO: maybe create assemble.Pointer?
	Tags          map[string]string
}

func (f *StructField) assembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if f.Description != "" {
		res = append(res, jen.Comment(f.Name+" -- "+utils.ToLowerFirstLetter(f.Description)))
	}

	stmt := jen.Id(f.Name)
	if f.Type.CanBePointer() && f.RequiredValue {
		stmt = stmt.Op("*")
	}
	items := utils.CastSliceItems[*jen.Statement, jen.Code](f.Type.AssembleUsage(ctx))
	res = append(res, stmt.Add(items...))

	return res
}
