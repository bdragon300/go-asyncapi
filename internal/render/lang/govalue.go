package lang

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
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
	return common.ObjectKindLang
}

func (gv GoValue) Selectable() bool {
	return false
}

func (gv GoValue) RenderContext() common.RenderContext {
	return context.Context
}

func (gv GoValue) U() string {
	//ctx.LogStartRender("GoValue", "", "", "usage", gv.IsDefinition(), "type", gv.Type)
	//defer ctx.LogFinishRender()
	//
	//var valueStmt *j.Statement
	//switch {
	//case gv.LiteralValue != nil:
	//	ctx.Logger.Trace("LiteralValue", "value", gv.LiteralValue)
	//	valueStmt = j.Lit(gv.LiteralValue)
	//	if v, ok := gv.LiteralValue.(common.Renderer); ok {
	//		valueStmt = j.Add(utils.ToCode(v.RenderUsage())...)
	//	}
	//case gv.MapValues.Len() > 0:
	//	ctx.Logger.Trace("MapValues", "value", gv.MapValues)
	//	valueStmt = j.Values(j.DictFunc(func(d j.Dict) {
	//		for _, e := range gv.MapValues.Entries() {
	//			l := []j.Code{j.Lit(e.Value)}
	//			if v, ok := e.Value.(common.Renderer); ok {
	//				l = utils.ToCode(v.RenderUsage())
	//			}
	//			d[j.Lit(e.Key)] = j.Add(l...)
	//		}
	//	}))
	//case gv.StructValues.Len() > 0:
	//	ctx.Logger.Trace("StructValues", "value", gv.StructValues)
	//	valueStmt = j.Values(j.DictFunc(func(d j.Dict) {
	//		for _, e := range gv.StructValues.Entries() {
	//			l := []j.Code{j.Lit(e.Value)}
	//			if v, ok := e.Value.(common.Renderer); ok {
	//				l = utils.ToCode(v.RenderUsage())
	//			}
	//			d[j.Id(e.Key)] = j.Add(l...)
	//		}
	//	}))
	//case gv.ArrayValues != nil:
	//	ctx.Logger.Trace("ArrayValues", "value", gv.ArrayValues)
	//	valueStmt = j.ValuesFunc(func(g *j.Group) {
	//		for _, v := range gv.ArrayValues {
	//			l := []j.Code{j.Lit(v)}
	//			if v, ok := v.(common.Renderer); ok {
	//				l = utils.ToCode(v.RenderUsage())
	//			}
	//			g.Add(l...)
	//		}
	//	})
	//default:
	//	ctx.Logger.Trace("Empty", "value", lo.Ternary(gv.EmptyCurlyBrackets, "{}", "nil"))
	//	valueStmt = lo.Ternary(gv.EmptyCurlyBrackets, j.Values(), j.Nil())
	//}
	//
	//if gv.Type == nil {
	//	return []*j.Statement{valueStmt}
	//}
	//
	//stmt := &j.Statement{}
	//if v, ok := gv.Type.(GolangPointerWrapperType); ok && v.IsPointer() {
	//	ctx.Logger.Trace("pointer")
	//	if gv.Empty() {
	//		if gv.EmptyCurlyBrackets {
	//			// &{} -> ToPtr({})
	//			return []*j.Statement{j.Qual(context.Context.RuntimeModule(""), "ToPtr").Call(j.Values())}
	//		}
	//		// &nil -> nil
	//		return []*j.Statement{j.Nil()}
	//	}
	//	if gv.LiteralValue != nil {
	//		if t, hasType := v.WrappedGolangType(); hasType {
	//			// &int(123) -> ToPtr(int(123))
	//			return []*j.Statement{j.Qual(context.Context.RuntimeModule(""), "ToPtr").Call(
	//				j.Add(utils.ToCode(t.U())...).Call(valueStmt),
	//			)}
	//		}
	//		// &123 -> ToPtr(123)
	//		return []*j.Statement{j.Qual(context.Context.RuntimeModule(""), "ToPtr").Call(valueStmt)}
	//	}
	//
	//	// &AnyType{}
	//	// &map[string]int{}
	//	// &[]int{}
	//	stmt = stmt.Op("&")
	//}
	//stmt = stmt.Add(utils.ToCode(gv.Type.U())...)
	//if gv.LiteralValue != nil {
	//	// int(123)
	//	return []*j.Statement{stmt.Call(j.Add(valueStmt))}
	//}
	//return []*j.Statement{stmt.Add(valueStmt)}
	panic("not implemented")
}

func (gv GoValue) AsRenderer(v any) common.Renderer {
	if r, ok := v.(common.Renderer); ok {
		return r
	}
	return nil
}

func (gv GoValue) AsPointerWrapperType(v any) GolangPointerWrapperType {
	if r, ok := v.(GolangPointerWrapperType); ok {
		return r
	}
	return nil
}

func (gv GoValue) AsGolangType(v GolangPointerWrapperType) common.GolangType {
	if r, ok := v.WrappedGolangType(); ok {
		return r
	}
	return nil
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

type stringAnyMap interface {
	Entries() []lo.Entry[string, any]
}

func ConstructGoValue(value any, excludeFields []string, overrideType common.GolangType) *GoValue {
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
