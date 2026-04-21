package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

// GoEnum represents a Go values enumeration for any type.
// For simple type, like string or int, it is a set of constants:
//
//	type Foo string
//	const (
//		FooA Foo = "a"
//		FooB Foo = "b"
//	)
//
// For struct type, it is a set of variables + init function, which initializes them with struct literals:
//
//	type Bar struct {
//		Name string
//	}
//	var (
//		BarA Bar
//		BarB Bar
//	)
//	func init() {
//		c := `{Name: "a"}`
//		json.Unmarshal([]byte(c), &BarA)
//		c = `{Name: "b"}`
//		json.Unmarshal([]byte(c), &BarB)
//	}
type GoEnum struct {
	BaseJSONPointed
	WrappedType common.GolangType
	// PrimitiveEnums contains unmarshalled enum values of primitive Go types (string, int, etc.) to add for WrappedType.
	// These values are supposed to be rendered as constants. Key is the constant name suffix.
	PrimitiveEnums types.OrderedMap[string, any]
	// ComplexEnums contains unmarshalled enum values of complex types (json object and array) to add for WrappedType.
	// These values are supposed to be rendered as initialization code in init() function. Key is the constant name suffix.
	ComplexEnums types.OrderedMap[string, any]
}

func (e *GoEnum) Name() string {
	return e.WrappedType.Name()
}

func (e *GoEnum) Kind() common.ArtifactKind {
	return e.WrappedType.Kind()
}

func (e *GoEnum) Selectable() bool {
	return e.WrappedType.Selectable()
}

func (e *GoEnum) Visible() bool {
	return e.WrappedType.Visible()
}

func (e *GoEnum) CanBeAddressed() bool {
	return e.WrappedType.CanBeAddressed()
}

func (e *GoEnum) String() string {
	return fmt.Sprintf("GoEnum -> %s", e.WrappedType.String())
}

func (e *GoEnum) GoTemplate() string {
	return "code/lang/goenum"
}

func (e *GoEnum) UnwrapGolangType() common.GolangType {
	if v, ok := e.WrappedType.(GolangWrappedType); ok {
		return v.UnwrapGolangType()
	}
	return e.WrappedType
}

func (e *GoEnum) IsStruct() bool {
	if v, ok := any(e.WrappedType).(golangStructType); ok {
		return v.IsStruct()
	}
	return false
}

func (e *GoEnum) StructRenderInfo() StructFieldRenderInfo {
	if v, ok := any(e.WrappedType).(structFieldRenderer); ok {
		return v.StructRenderInfo()
	}
	return StructFieldRenderInfo{}
}
