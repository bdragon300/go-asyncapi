package lang

import (
	"fmt"
	"reflect"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

// GoValue represents a Go value of any other type that can be rendered in Go code.
//
// This value can be a constant, a struct, array or map initialization expression. This is suitable when some data
// from the AsyncAPI document should get to the code as initialization value of some type. For example, the AsyncAPI bindings.
type GoValue struct {
	BaseJSONPointed
	// Type is the type that is rendered before the value, e.g. ``int(123)'' or ``map[string]string{"123", "456"}''.
	// If nil, the value will be rendered as a bare untyped value, like ``123'' or ``{"123", "456"}''.
	Type common.GolangType
	// EmptyCurlyBrackets affects to the rendering if Empty returns true. If true, then empty value will
	// be rendered as ``{}``, otherwise as ``nil``.
	EmptyCurlyBrackets bool // If GoValue is empty: `{}` if true, or `nil` otherwise
	// LiteralValue is a value that should be rendered as a literal value, like ``123`` or ``"hello"``.
	LiteralValue any
	// ArrayValues is a list of values that should be rendered as an array/slice initialization in curly brackets.
	ArrayValues []any
	// StructValues is a list of key-value pairs that should be rendered as a struct initialization in curly brackets.
	StructValues types.OrderedMap[string, any]
	// MapValues is a list of key-value pairs that should be rendered as a map initialization in curly brackets.
	MapValues types.OrderedMap[string, any]
}

// Empty returns true if the GoValue represents nil or empty or zero value.
func (gv *GoValue) Empty() bool {
	return gv.LiteralValue == nil && gv.StructValues.Len() == 0 && gv.MapValues.Len() == 0 && gv.ArrayValues == nil
}

func (gv *GoValue) Name() string {
	return ""
}

func (gv *GoValue) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (gv *GoValue) Selectable() bool {
	return false
}

func (gv *GoValue) Visible() bool {
	return true
}

func (gv *GoValue) String() string {
	switch {
	case gv.LiteralValue != nil:
		return fmt.Sprintf("GoValue:%v", gv.LiteralValue)
	case gv.StructValues.Len() > 0:
		return fmt.Sprintf("GoValue:{%v...}", lo.Slice(gv.StructValues.Entries(), 0, 2))
	case gv.MapValues.Len() > 0:
		return fmt.Sprintf("GoValue:{%v...}", lo.Slice(gv.MapValues.Entries(), 0, 2))
	case gv.ArrayValues != nil:
		return fmt.Sprintf("GoValue:{%v...}", lo.Slice(gv.ArrayValues, 0, 2))
	}
	return "GoValue:nil"
}

func (gv *GoValue) CanBeAddressed() bool {
	return gv.Type != nil && gv.Type.CanBeAddressed()
}

func (gv *GoValue) CanBeDereferenced() bool {
	return gv.Type != nil && gv.Type.CanBeDereferenced()
}

func (gv *GoValue) GoTemplate() string {
	return "code/lang/govalue"
}

// ConstructGoValue converts the input object into its GoValue representation, that can be rendered in templates as
// type initialization expression. The result contains the same values and fields as input, recursively, except the
// resulted type, which is set to outputType. Fields in input (and its nested structs) with empty value or that are in
// excludeStructFields are skipped.
//
// In other words, this function "marshals" the value into a GoValue object with the outputType type info attached.
func ConstructGoValue(input any, excludeStructFields []string, outputType common.GolangType) *GoValue {
	type stringAnyMap interface {
		Entries() []lo.Entry[string, any]
	}

	res := GoValue{Type: outputType}
	if input == nil {
		return &res
	}

	switch v := input.(type) {
	case stringAnyMap:
		for _, e := range v.Entries() {
			res.StructValues.Set(e.Key, ConstructGoValue(e.Value, excludeStructFields, nil))
		}
		return &res
	case *GoValue:
		return v
	case GoValue:
		return &v
	}

	rtyp := reflect.TypeOf(input)
	rval := reflect.ValueOf(input)

	switch rtyp.Kind() {
	case reflect.Slice, reflect.Array:
		elemType := rtyp.Elem()
		var elemSize int
		if rtyp.Kind() == reflect.Array {
			elemSize = rtyp.Len()
		}
		if res.Type == nil {
			res.Type = &GoArray{
				BaseType: BaseType{OriginalName: rtyp.Name(), Import: rtyp.PkgPath()},
				ItemsType: &GoSimple{
					TypeName:    elemType.Name(),
					IsInterface: elemType.Kind() == reflect.Interface,
					Import:      elemType.PkgPath(),
				},
				Size: elemSize,
			}
		}
		res.EmptyCurlyBrackets = true

		for i := 0; i < rval.Len(); i++ {
			res.ArrayValues = append(res.ArrayValues, ConstructGoValue(rval.Index(i).Interface(), excludeStructFields, nil))
		}
		return &res
	case reflect.Map:
		keyType := rtyp.Key()
		elemType := rtyp.Elem()
		if res.Type == nil {
			res.Type = &GoMap{
				BaseType: BaseType{OriginalName: rtyp.Name(), Import: rtyp.PkgPath()},
				KeyType: &GoSimple{
					TypeName:    keyType.Name(),
					IsInterface: keyType.Kind() == reflect.Interface,
					Import:      keyType.PkgPath(),
				},
				ValueType: &GoSimple{
					TypeName:    elemType.Name(),
					IsInterface: elemType.Kind() == reflect.Interface,
					Import:      elemType.PkgPath(),
				},
			}
		}
		res.EmptyCurlyBrackets = true

		for _, k := range rval.MapKeys() {
			res.MapValues.Set(k.String(), ConstructGoValue(rval.MapIndex(k).Interface(), excludeStructFields, nil))
		}
		return &res
	case reflect.Struct:
		if res.Type == nil {
			res.Type = &GoStruct{
				BaseType: BaseType{OriginalName: rtyp.Name(), Import: rtyp.PkgPath()},
			}
		}
		res.EmptyCurlyBrackets = true

		for i := 0; i < rval.NumField(); i++ {
			ftyp := rtyp.Field(i)
			fval := rval.Field(i)
			if fval.IsZero() || lo.Contains(excludeStructFields, ftyp.Name) {
				continue // Skip empty values in struct initializations, or if it excluded
			}
			res.StructValues.Set(ftyp.Name, ConstructGoValue(fval.Interface(), excludeStructFields, nil))
		}
		return &res
	case reflect.Pointer, reflect.Interface:
		pval := reflect.Indirect(rval)
		goval := ConstructGoValue(pval.Interface(), excludeStructFields, nil)
		return &GoValue{LiteralValue: goval, Type: &GoPointer{Type: outputType}}
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return &GoValue{LiteralValue: rval.Interface(), Type: outputType}
	}

	panic(fmt.Errorf("cannot construct Value from a input of type %T", input))
}
