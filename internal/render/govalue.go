package render

import (
	"fmt"
	"reflect"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type GoValue struct {
	Type            common.GolangType // if not nil then it will be rendered before value
	Literal         any
	ArrayVals       []any
	StructVals      types.OrderedMap[string, any]
	DictVals        types.OrderedMap[string, any]
	NilCurlyBrakets bool // If GoValue is empty: `{}` if true, or `nil` otherwise
}

type golangPointerWrapperType interface {
	golangTypeWrapperType
	golangPointerType
}

func (gv GoValue) DirectRendering() bool {
	return false
}

func (gv GoValue) RenderDefinition(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (gv GoValue) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	ctx.LogRender("GoValue", "", "", "usage", gv.DirectRendering(), "type", gv.Type)
	defer ctx.LogReturn()

	var valueStmt *j.Statement
	switch {
	case gv.Literal != nil:
		ctx.Logger.Trace("Literal", "value", gv.Literal)
		valueStmt = j.Lit(gv.Literal)
		if v, ok := gv.Literal.(common.Renderer); ok {
			valueStmt = j.Add(utils.ToCode(v.RenderUsage(ctx))...)
		}
	case gv.DictVals.Len() > 0:
		ctx.Logger.Trace("DictVals", "value", gv.DictVals)
		valueStmt = j.Values(j.DictFunc(func(d j.Dict) {
			for _, e := range gv.DictVals.Entries() {
				l := []j.Code{j.Lit(e.Value)}
				if v, ok := e.Value.(common.Renderer); ok {
					l = utils.ToCode(v.RenderUsage(ctx))
				}
				d[j.Lit(e.Key)] = j.Add(l...)
			}
		}))
	case gv.StructVals.Len() > 0:
		ctx.Logger.Trace("StructVals", "value", gv.StructVals)
		valueStmt = j.Values(j.DictFunc(func(d j.Dict) {
			for _, e := range gv.StructVals.Entries() {
				l := []j.Code{j.Lit(e.Value)}
				if v, ok := e.Value.(common.Renderer); ok {
					l = utils.ToCode(v.RenderUsage(ctx))
				}
				d[j.Id(e.Key)] = j.Add(l...)
			}
		}))
	case gv.ArrayVals != nil:
		ctx.Logger.Trace("ArrayVals", "value", gv.ArrayVals)
		valueStmt = j.ValuesFunc(func(g *j.Group) {
			for _, v := range gv.ArrayVals {
				l := []j.Code{j.Lit(v)}
				if v, ok := v.(common.Renderer); ok {
					l = utils.ToCode(v.RenderUsage(ctx))
				}
				g.Add(l...)
			}
		})
	default:
		ctx.Logger.Trace("Empty", "value", lo.Ternary(gv.NilCurlyBrakets, "{}", "nil"))
		valueStmt = lo.Ternary(gv.NilCurlyBrakets, j.Values(), j.Nil())
	}

	if gv.Type == nil {
		return []*j.Statement{valueStmt}
	}

	stmt := &j.Statement{}
	if v, ok := gv.Type.(golangPointerWrapperType); ok && v.IsPointer() {
		ctx.Logger.Trace("pointer")
		if gv.Empty() {
			if gv.NilCurlyBrakets {
				// &{} -> ToPtr({})
				return []*j.Statement{j.Qual(ctx.RuntimeModule(""), "ToPtr").Call(j.Values())}
			}
			// &nil -> nil
			return []*j.Statement{j.Nil()}
		}
		if gv.Literal != nil {
			if t, hasType := v.WrappedGolangType(); hasType {
				// &int(123) -> ToPtr(int(123))
				return []*j.Statement{j.Qual(ctx.RuntimeModule(""), "ToPtr").Call(
					j.Add(utils.ToCode(t.RenderUsage(ctx))...).Call(valueStmt),
				)}
			}
			// &123 -> ToPtr(123)
			return []*j.Statement{j.Qual(ctx.RuntimeModule(""), "ToPtr").Call(valueStmt)}
		}

		// &AnyType{}
		// &map[string]int{}
		// &[]int{}
		stmt = stmt.Op("&")
	}
	stmt = stmt.Add(utils.ToCode(gv.Type.RenderUsage(ctx))...)
	if gv.Literal != nil {
		// int(123)
		return []*j.Statement{stmt.Call(j.Add(valueStmt))}
	}
	return []*j.Statement{stmt.Add(valueStmt)}
}

func (gv GoValue) Empty() bool {
	return gv.Literal == nil && gv.StructVals.Len() == 0 && gv.DictVals.Len() == 0 && gv.ArrayVals == nil
}

func (gv GoValue) ID() string {
	return "GoValue"
}

func (gv GoValue) String() string {
	switch {
	case gv.Literal != nil:
		return fmt.Sprintf("GoValue %v", gv.Literal)
	case gv.StructVals.Len() > 0:
		return fmt.Sprintf("GoValue {%v...}", lo.Slice(gv.StructVals.Entries(), 0, 2))
	case gv.DictVals.Len() > 0:
		return fmt.Sprintf("GoValue {%v...}", lo.Slice(gv.DictVals.Entries(), 0, 2))
	case gv.ArrayVals != nil:
		return fmt.Sprintf("GoValue {%v...}", lo.Slice(gv.ArrayVals, 0, 2))
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
			res.StructVals.Set(e.Key, ConstructGoValue(e.Value, excludeFields, nil))
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
					Name:    elemType.Name(),
					IsIface: elemType.Kind() == reflect.Interface,
					Import:  elemType.PkgPath(),
				},
				Size: elemSize,
			}
		}
		res.NilCurlyBrakets = true

		for i := 0; i < rval.Len(); i++ {
			res.ArrayVals = append(res.ArrayVals, ConstructGoValue(rval.Index(i).Interface(), excludeFields, nil))
		}
		return &res
	case reflect.Map:
		keyType := rtyp.Key()
		elemType := rtyp.Elem()
		if res.Type == nil {
			res.Type = &GoMap{
				BaseType: BaseType{Name: rtyp.Name(), Import: rtyp.PkgPath()},
				KeyType: &GoSimple{
					Name:    keyType.Name(),
					IsIface: keyType.Kind() == reflect.Interface,
					Import:  keyType.PkgPath(),
				},
				ValueType: &GoSimple{
					Name:    elemType.Name(),
					IsIface: elemType.Kind() == reflect.Interface,
					Import:  elemType.PkgPath(),
				},
			}
		}
		res.NilCurlyBrakets = true

		for _, k := range rval.MapKeys() {
			res.DictVals.Set(k.String(), ConstructGoValue(rval.MapIndex(k).Interface(), excludeFields, nil))
		}
		return &res
	case reflect.Struct:
		if res.Type == nil {
			res.Type = &GoStruct{
				BaseType: BaseType{Name: rtyp.Name(), Import: rtyp.PkgPath()},
			}
		}
		res.NilCurlyBrakets = true

		for i := 0; i < rval.NumField(); i++ {
			ftyp := rtyp.Field(i)
			fval := rval.Field(i)
			if fval.IsZero() || lo.Contains(excludeFields, ftyp.Name) {
				continue // Skip empty values in struct initializations, or if it excluded
			}
			res.StructVals.Set(ftyp.Name, ConstructGoValue(fval.Interface(), excludeFields, nil))
		}
		return &res
	case reflect.Pointer, reflect.Interface:
		pval := reflect.Indirect(rval)
		val := ConstructGoValue(pval.Interface(), excludeFields, nil)
		return &GoValue{Literal: val, Type: &GoPointer{Type: overrideType}}
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return &GoValue{Literal: rval.Interface(), Type: overrideType}
	}

	panic(fmt.Errorf("cannot construct Value from a value of type %T", value))
}
