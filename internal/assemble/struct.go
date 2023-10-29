package assemble

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
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
		if s.PackageName != "" && s.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, s.PackageName), s.Name)}
		}
		return []*jen.Statement{jen.Id(s.Name)}
	}

	code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
		return item.assembleDefinition(ctx)
	})

	return []*jen.Statement{jen.Struct(utils.ToCode(code)...)}
}

func (s Struct) NewFuncName() string {
	return "New" + s.Name
}

func (s Struct) ReceiverName() string {
	return strings.ToLower(string(s.Name[0]))
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
		drawPtr = false // Prevent render double pointed field
	}
	if drawPtr {
		stmt = stmt.Op("*")
	}
	items := utils.ToCode(f.Type.AssembleUsage(ctx))
	res = append(res, stmt.Add(items...))

	return res
}

type UnionStruct struct {
	Struct
}

func (s UnionStruct) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	hasNonStructs := lo.ContainsBy(s.Fields, func(item StructField) bool {
		return !isTypeStruct(item.Type)
	})
	if hasNonStructs { // Draw union with named fields and methods
		strct := s.Struct
		strct.Fields = lo.Map(strct.Fields, func(item StructField, index int) StructField {
			item.Name = item.Type.TypeName()
			return item
		})
		if reflect.DeepEqual(strct.Fields, s.Fields) { // TODO: move this check to unit tests
			panic("Must not happen")
		}
		res = strct.AssembleDefinition(ctx)
		res = append(res, s.assembleMethods(ctx)...)
	} else { // Draw simplified union with embedded fields
		res = s.Struct.AssembleDefinition(ctx)
	}
	return res
}

func (s UnionStruct) assembleMethods(_ *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	receiverName := strings.ToLower(string(s.Struct.Name[0]))

	// Method UnmarshalJSON(bytes []byte) error
	body := []jen.Code{jen.Var().Err().Error()}
	stmt := &jen.Statement{}
	for _, f := range s.Struct.Fields {
		op := ""
		if !f.ForcePointer {
			op = "&"
		}
		stmt = stmt.If(
			jen.Err().Op("=").Qual("encoding/json", "Unmarshal").Call(jen.Id("bytes"), jen.Op(op).Id(receiverName).Dot(f.Type.TypeName())),
			jen.Err().Op("!=").Nil(),
		).
			Block(jen.Return(jen.Nil())).
			Else()
	}
	if len(s.Struct.Fields) > 0 {
		stmt = stmt.Block(jen.Return(jen.Err()))
	} else {
		stmt = stmt.Return(jen.Return(jen.Nil()))
	}
	body = append(body, stmt)

	res = append(res, jen.Func().Params(jen.Id(receiverName).Op("*").Id(s.Struct.Name)).Id("UnmarshalJSON").
		Params(jen.Id("bytes").Index().Byte()).
		Error().
		Block(body...),
	)

	return res
}

type structInitAssembler interface {
	AssembleInit(ctx *common.AssembleContext) []*jen.Statement
}

type StructInit struct {
	Type   common.GolangType
	Values utils.OrderedMap[string, any]
}

func (s StructInit) AssembleInit(ctx *common.AssembleContext) []*jen.Statement {
	stmt := &jen.Statement{}
	if s.Type != nil {
		stmt.Add(utils.ToCode(s.Type.AssembleUsage(ctx))...)
	}

	dict := make(jen.Dict)
	for _, e := range s.Values.Entries() {
		switch v := e.Value.(type) {
		case common.Assembler:
			dict[jen.Id(e.Key)] = jen.Add(utils.ToCode(v.AssembleUsage(ctx))...)
		case structInitAssembler:
			dict[jen.Id(e.Key)] = jen.Add(utils.ToCode(v.AssembleInit(ctx))...)
		default:
			dict[jen.Id(e.Key)] = jen.Lit(v)
		}
	}

	return []*jen.Statement{stmt.Values(dict)}
}

func isTypeStruct(typ common.GolangType) bool {
	switch v := typ.(type) {
	case *Struct, *UnionStruct:
		return true
	case *LinkAsGolangType:
		return isTypeStruct(v.Target())
	case *NullableType:
		_, ok := v.Type.(*Struct)
		return ok
	}
	return false
}
