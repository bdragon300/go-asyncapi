package render

import (
	"reflect"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type UnionStruct struct {
	GoStruct
}

func (s UnionStruct) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	ctx.LogStartRender("UnionStruct", s.Import, s.Name, "definition", s.DirectRendering())
	defer ctx.LogFinishRender()

	onlyStructs := lo.EveryBy(s.Fields, func(item GoStructField) bool {
		return isTypeStruct(item.Type)
	})
	if onlyStructs { // Draw simplified union with embedded fields
		res = s.GoStruct.RenderDefinition(ctx)
	} else { // Draw union with named fields and methods
		strct := s.GoStruct
		strct.Fields = lo.Map(strct.Fields, func(item GoStructField, _ int) GoStructField {
			item.Name = item.Type.TypeName()
			return item
		})
		if reflect.DeepEqual(strct.Fields, s.Fields) { // TODO: move this check to unit tests
			panic("Must not happen")
		}
		res = strct.RenderDefinition(ctx)
		res = append(res, s.renderMethods(ctx)...)
	}
	return res
}

func (s UnionStruct) renderMethods(ctx *common.RenderContext) []*jen.Statement {
	ctx.Logger.Trace("renderMethods")

	var res []*jen.Statement
	receiverName := strings.ToLower(string(s.GoStruct.Name[0]))

	// Method UnmarshalJSON(bytes []byte) error
	body := []jen.Code{jen.Var().Err().Error()}
	stmt := &jen.Statement{}
	for _, f := range s.GoStruct.Fields {
		op := ""
		if v, ok := f.Type.(golangPointerType); !ok || !v.IsPointer() { // No need to take address for a pointer
			op = "&"
		}
		stmt = stmt.If(
			jen.Err().Op("=").Qual("encoding/json", "Unmarshal").Call(jen.Id("bytes"), jen.Op(op).Id(receiverName).Dot(f.Type.TypeName())),
			jen.Err().Op("!=").Nil(),
		).
			Block(jen.Return(jen.Nil())).
			Else()
	}
	if len(s.GoStruct.Fields) > 0 {
		stmt = stmt.Block(jen.Return(jen.Err()))
	} else {
		stmt = stmt.Return(jen.Return(jen.Nil()))
	}
	body = append(body, stmt)

	res = append(res, jen.Func().Params(jen.Id(receiverName).Op("*").Id(s.GoStruct.Name)).Id("UnmarshalJSON").
		Params(jen.Id("bytes").Index().Byte()).
		Error().
		Block(body...),
	)

	return res
}

func isTypeStruct(typ common.GolangType) bool {
	switch v := typ.(type) {
	case golangStructType:
		return v.IsStruct()
	case golangTypeWrapperType:
		t, ok := v.WrappedGolangType()
		return !ok || isTypeStruct(t)
	}
	return false
}
