package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"reflect"
)

type UnionStruct struct {
	GoStruct
}

func (s UnionStruct) D() string {
	//var res []*jen.Statement
	//ctx.LogStartRender("UnionStruct", s.Import, s.Name, "definition", s.IsDefinition())
	//defer ctx.LogFinishRender()
	//
	//onlyStructs := lo.EveryBy(s.Fields, func(item GoStructField) bool {
	//	return isTypeStruct(item.Type)
	//})
	//if onlyStructs { // Draw simplified union with embedded fields
	//	res = s.GoStruct.D()
	//} else { // Draw union with named fields and methods
	//	strct := s.GoStruct
	//	strct.Fields = lo.Map(strct.Fields, func(item GoStructField, _ int) GoStructField {
	//		item.Name = item.Type.TypeName()
	//		return item
	//	})
	//	if reflect.DeepEqual(strct.Fields, s.Fields) { // TODO: move this check to unit tests
	//		panic("Must not happen")
	//	}
	//	res = strct.D()
	//	res = append(res, s.renderMethods()...)
	//}
	//return res
	return renderTemplate("lang/union/definition", &s)
}

func (s UnionStruct) UnionStruct() common.GolangType {
	onlyStructs := lo.EveryBy(s.Fields, func(item GoStructField) bool {
		return isTypeStruct(item.Type)
	})
	if onlyStructs { // Draw simplified union with embedded fields
		return &s.GoStruct
	}

	// Draw union with named fields and methods
	strct := s.GoStruct
	strct.Fields = lo.Map(strct.Fields, func(item GoStructField, _ int) GoStructField {
		item.Name = item.Type.TypeName()
		return item
	})
	if reflect.DeepEqual(strct.Fields, s.Fields) { // TODO: move this check to unit tests
		panic("Must not happen")
	}
	return &strct
}

func (s UnionStruct) String() string {
	if s.Import != "" {
		return "UnionStruct /" + s.Import + "." + s.Name
	}
	return "UnionStruct " + s.Name
}

//func (s UnionStruct) renderMethods() []*jen.Statement {
//	ctx.Logger.Trace("renderMethods")
//
//	var res []*jen.Statement
//	receiverName := strings.ToLower(string(s.GoStruct.Name[0]))
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
//			jen.Err().Op("=").Qual("encoding/json", "Unmarshal").Call(jen.Id("bytes"), jen.Op(op).Id(receiverName).Dot(f.Type.TypeName())),
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
//	res = append(res, jen.Func().Params(jen.Id(receiverName).Op("*").Id(s.GoStruct.Name)).Id("UnmarshalJSON").
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
