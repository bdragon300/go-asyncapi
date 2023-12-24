package render

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

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

func (s Struct) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	ctx.LogRender("Struct", s.PackageName, s.Name, "definition", s.DirectRendering())
	defer ctx.LogReturn()

	if s.Description != "" {
		res = append(res, jen.Comment(s.Name+" -- "+utils.ToLowerFirstLetter(s.Description)))
	}
	code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
		return item.renderDefinition(ctx)
	})
	res = append(res, jen.Type().Id(s.Name).Struct(utils.ToCode(code)...))
	return res
}

func (s Struct) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("Struct", s.PackageName, s.Name, "usage", s.DirectRendering())
	defer ctx.LogReturn()

	if s.DirectRendering() {
		if s.PackageName != "" && s.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedPackage(s.PackageName), s.Name)}
		}
		return []*jen.Statement{jen.Id(s.Name)}
	}

	code := lo.FlatMap(s.Fields, func(item StructField, index int) []*jen.Statement {
		return item.renderDefinition(ctx)
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
	MarshalName  string
	Description  string
	Type         common.GolangType
	ForcePointer bool // TODO: remove in favor of NullableType
	TagsSource   *LinkList[*Message]
}

func (f StructField) renderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	ctx.LogRender("StructField", "", f.Name, "definition", false)
	defer ctx.LogReturn()

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
	items := utils.ToCode(f.Type.RenderUsage(ctx))
	stmt = stmt.Add(items...)

	if f.TagsSource != nil {
		tagNames := lo.Uniq(lo.FilterMap(f.TagsSource.Targets(), func(item *Message, index int) (string, bool) {
			format := getFormatByContentType(item.ContentType)
			return format, format != ""
		}))
		tags := lo.SliceToMap(tagNames, func(item string) (string, string) {
			return item, f.MarshalName
		})
		stmt.Tag(tags)
	}

	res = append(res, stmt)

	return res
}

type UnionStruct struct {
	Struct
}

func (s UnionStruct) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	ctx.LogRender("UnionStruct", s.PackageName, s.Name, "definition", s.DirectRendering())
	defer ctx.LogReturn()

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
		res = strct.RenderDefinition(ctx)
		res = append(res, s.renderMethods(ctx)...)
	} else { // Draw simplified union with embedded fields
		res = s.Struct.RenderDefinition(ctx)
	}
	return res
}

func (s UnionStruct) renderMethods(_ *common.RenderContext) []*jen.Statement {
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

type structInitRenderer interface {
	RenderInit(ctx *common.RenderContext) []*jen.Statement
}

type StructInit struct {
	Type   common.GolangType
	Values types.OrderedMap[string, any]
}

func (s StructInit) RenderInit(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("StructInit", "", "", "definition", false)
	defer ctx.LogReturn()

	stmt := &jen.Statement{}
	if s.Type != nil {
		stmt.Add(utils.ToCode(s.Type.RenderUsage(ctx))...)
	}

	dict := make(jen.Dict)
	for _, e := range s.Values.Entries() {
		switch v := e.Value.(type) {
		case common.Renderer:
			dict[jen.Id(e.Key)] = jen.Add(utils.ToCode(v.RenderUsage(ctx))...)
		case structInitRenderer:
			dict[jen.Id(e.Key)] = jen.Add(utils.ToCode(v.RenderInit(ctx))...)
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
