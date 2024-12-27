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

func isTypeStruct(typ common.GolangType) bool {
	switch v := typ.(type) {
	case golangStructType:
		return v.IsStruct()
	case GolangTypeExtractor:
		t := v.InnerGolangType()
		return !lo.IsNil(t) && isTypeStruct(t)
	}
	return false
}
