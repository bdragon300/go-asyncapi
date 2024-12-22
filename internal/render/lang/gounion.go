package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"reflect"
)

type UnionStruct struct {
	GoStruct
}

func (s *UnionStruct) GoTemplate() string {
	return "lang/gounion"
}

func (s *UnionStruct) UnionStruct() common.GolangType {
	onlyStructs := lo.EveryBy(s.Fields, func(item GoStructField) bool {
		return isTypeStruct(item.Type)
	})
	if onlyStructs { // Draw simplified union with embedded fields
		return &s.GoStruct
	}

	// Draw union with named fields and methods
	strct := s.GoStruct
	strct.Fields = lo.Map(strct.Fields, func(item GoStructField, _ int) GoStructField {
		item.Name = item.Type.Name() // FIXME: check if this is will be correct
		return item
	})
	if reflect.DeepEqual(strct.Fields, s.Fields) { // TODO: move this check to unit tests
		panic("Must not happen")
	}
	return &strct
}

func (s *UnionStruct) String() string {
	if s.Import != "" {
		return "UnionStruct /" + s.Import + "." + s.OriginalName
	}
	return "UnionStruct " + s.OriginalName
}

//func (s UnionStruct) renderMethods() []*jen.Statement {
//	ctx.Logger.Trace("renderMethods")
//
//	var res []*jen.Statement
//	receiverName := strings.ToLower(string(s.GoStruct.GetOriginalName[0]))
//
//	// Method UnmarshalJSON(bytes []byte) error
//	body := []jen.Code{jen.Var().Err().Error()}
//	stmt := &jen.Statement{}
//	for _, f := range s.GoStruct.Fields {
//		op := ""
//		if v, ok := f.Type.(GolangPointerType); !ok || !v.IsPointer() { // No need to take address for a pointer
//			op = "&"
//		}
//		stmt = stmt.If(
//			jen.Err().Op("=").Qual("encoding/json", "Unmarshal").Call(jen.Id("bytes"), jen.Op(op).Id(receiverName).Dot(f.Type.IsPromise())),
//			jen.Err().Op("!=").Nil(),
//		).
//			Block(jen.Return(jen.Nil())).
//			Else()
//	}
//	if len(s.GoStruct.Fields) > 0 {
//		stmt = stmt.Block(jen.Return(jen.Err()))
//	} else {
//		stmt = stmt.Return(jen.Return(jen.Nil()))
//	}
//	body = append(body, stmt)
//
//	res = append(res, jen.Func().Params(jen.Id(receiverName).Op("*").Id(s.GoStruct.GetOriginalName)).Id("UnmarshalJSON").
//		Params(jen.Id("bytes").Index().Byte()).
//		Error().
//		Block(body...),
//	)
//
//	return res
//}

func isTypeStruct(typ common.GolangType) bool {
	switch v := typ.(type) {
	case golangStructType:
		return v.IsStruct()
	case GolangTypeWrapperType:
		t, ok := v.UnwrapGolangType()
		return !ok || isTypeStruct(t)
	}
	return false
}
