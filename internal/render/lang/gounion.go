package lang

import (
	"fmt"
	"reflect"

	"github.com/bdragon300/go-asyncapi/internal/common"

	"github.com/samber/lo"
)

// UnionStruct represents a union struct, a special case of Go struct.
//
// Union struct is a struct that can be one of the several types.
// This struct is used to be generated the Go code from polymorphic jsonschema parts, such as $allOf, $oneOf, $anyOf.
// So, the data that matches such schema can be unmarshalled to the union type and addressed from the user code and
// be marshalled back.
type UnionStruct struct {
	GoStruct
}

// UnionStruct return the Go code of union struct definition.
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
		item.OriginalName = item.Type.Name() // FIXME: check if this is will be correct
		return item
	})
	if reflect.DeepEqual(strct.Fields, s.Fields) { // TODO: move this check to unit tests
		panic("Must not happen")
	}
	return &strct
}

func (s *UnionStruct) String() string {
	if s.Import != "" {
		return fmt.Sprintf("UnionStruct(%s.%s)", s.Import, s.OriginalName)
	}
	return "UnionStruct(" + s.OriginalName + ")"
}

func (s *UnionStruct) GoTemplate() string {
	return "code/lang/gounion"
}

func isTypeStruct(typ common.GolangType) bool {
	switch v := typ.(type) {
	case golangStructType:
		return v.IsStruct()
	case GolangWrappedType:
		t := v.UnwrapGolangType()
		return !lo.IsNil(t) && isTypeStruct(t)
	}
	return false
}
