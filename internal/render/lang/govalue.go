package lang

import (
	"fmt"
	"reflect"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type GoValue struct {
	Type               common.GolangType             // if not nil then it will be rendered before value
	EmptyCurlyBrackets bool                          // If GoValue is empty: `{}` if true, or `nil` otherwise
	LiteralValue       any                           // Render as literal value
	ArrayValues        []any                         // Render as array/slice initialization in curly brackets
	StructValues       types.OrderedMap[string, any] // Render as struct inline definition with following field values in curly brackets
	MapValues          types.OrderedMap[string, any] // Render as map initialization in curly brackets
}

type GolangPointerWrapperType interface {
	GolangTypeWrapperType
	GolangPointerType
}

func (gv GoValue) Kind() common.ObjectKind {
	return common.ObjectKindOther
}

func (gv GoValue) Selectable() bool {
	return false
}

func (gv GoValue) U() string {
	return renderTemplate("lang/govalue/usage", &gv)
}

func (gv GoValue) Empty() bool {
	return gv.LiteralValue == nil && gv.StructValues.Len() == 0 && gv.MapValues.Len() == 0 && gv.ArrayValues == nil
}

func (gv GoValue) String() string {
	switch {
	case gv.LiteralValue != nil:
		return fmt.Sprintf("GoValue %v", gv.LiteralValue)
	case gv.StructValues.Len() > 0:
		return fmt.Sprintf("GoValue {%v...}", lo.Slice(gv.StructValues.Entries(), 0, 2))
	case gv.MapValues.Len() > 0:
		return fmt.Sprintf("GoValue {%v...}", lo.Slice(gv.MapValues.Entries(), 0, 2))
	case gv.ArrayValues != nil:
		return fmt.Sprintf("GoValue {%v...}", lo.Slice(gv.ArrayValues, 0, 2))
	}
	return "GoValue nil"
}


func ConstructGoValue(value any, excludeFields []string, overrideType common.GolangType) *GoValue {
	type stringAnyMap interface {
		Entries() []lo.Entry[string, any]
	}

	res := GoValue{Type: overrideType}
	if value == nil {
		return &res
	}

	switch v := value.(type) {
	case stringAnyMap:
		for _, e := range v.Entries() {
			res.StructValues.Set(e.Key, ConstructGoValue(e.Value, excludeFields, nil))
		}
		return &res
	case *GoValue:
		return v
	case GoValue:
		return &v
	}

	rtyp := reflect.TypeOf(value)
	rval := reflect.ValueOf(value)

	switch rtyp.Kind() {
	case reflect.Slice, reflect.Array:
		elemType := rtyp.Elem()
		var elemSize int
		if rtyp.Kind() == reflect.Array {
			elemSize = rtyp.Len()
		}
		if res.Type == nil {
			res.Type = &GoArray{
				BaseType: BaseType{Name: rtyp.Name(), Import: rtyp.PkgPath()},
				ItemsType: &GoSimple{
					Name:        elemType.Name(),
					IsInterface: elemType.Kind() == reflect.Interface,
					Import:      elemType.PkgPath(),
				},
				Size: elemSize,
			}
		}
		res.EmptyCurlyBrackets = true

		for i := 0; i < rval.Len(); i++ {
			res.ArrayValues = append(res.ArrayValues, ConstructGoValue(rval.Index(i).Interface(), excludeFields, nil))
		}
		return &res
	case reflect.Map:
		keyType := rtyp.Key()
		elemType := rtyp.Elem()
		if res.Type == nil {
			res.Type = &GoMap{
				BaseType: BaseType{Name: rtyp.Name(), Import: rtyp.PkgPath()},
				KeyType: &GoSimple{
					Name:        keyType.Name(),
					IsInterface: keyType.Kind() == reflect.Interface,
					Import:      keyType.PkgPath(),
				},
				ValueType: &GoSimple{
					Name:        elemType.Name(),
					IsInterface: elemType.Kind() == reflect.Interface,
					Import:      elemType.PkgPath(),
				},
			}
		}
		res.EmptyCurlyBrackets = true

		for _, k := range rval.MapKeys() {
			res.MapValues.Set(k.String(), ConstructGoValue(rval.MapIndex(k).Interface(), excludeFields, nil))
		}
		return &res
	case reflect.Struct:
		if res.Type == nil {
			res.Type = &GoStruct{
				BaseType: BaseType{Name: rtyp.Name(), Import: rtyp.PkgPath()},
			}
		}
		res.EmptyCurlyBrackets = true

		for i := 0; i < rval.NumField(); i++ {
			ftyp := rtyp.Field(i)
			fval := rval.Field(i)
			if fval.IsZero() || lo.Contains(excludeFields, ftyp.Name) {
				continue // Skip empty values in struct initializations, or if it excluded
			}
			res.StructValues.Set(ftyp.Name, ConstructGoValue(fval.Interface(), excludeFields, nil))
		}
		return &res
	case reflect.Pointer, reflect.Interface:
		pval := reflect.Indirect(rval)
		val := ConstructGoValue(pval.Interface(), excludeFields, nil)
		return &GoValue{LiteralValue: val, Type: &GoPointer{Type: overrideType}}
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return &GoValue{LiteralValue: rval.Interface(), Type: overrideType}
	}

	panic(fmt.Errorf("cannot construct Value from a value of type %T", value))
}
