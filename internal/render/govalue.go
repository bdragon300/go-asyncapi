package render

import (
	"fmt"
	"reflect"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
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
	// TODO: logs
	var valueStmt *j.Statement
	switch {
	case gv.Literal != nil:
		valueStmt = j.Lit(gv.Literal)
		if v, ok := gv.Literal.(common.Renderer); ok {
			valueStmt = j.Add(utils.ToCode(v.RenderUsage(ctx))...)
		}
	case gv.DictVals.Len() > 0:
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
		valueStmt = lo.Ternary(gv.NilCurlyBrakets, j.Values(), j.Nil())
	}

	if gv.Type == nil {
		return []*j.Statement{valueStmt}
	}

	stmt := &j.Statement{}
	if v, ok := gv.Type.(golangPointerWrapperType); ok && v.IsPointer() {
		if gv.Empty() {
			if gv.NilCurlyBrakets {
				// &{} -> ToPtr({})
				return []*j.Statement{j.Qual(ctx.RuntimePackage(""), "ToPtr").Call(j.Values())}
			}
			// &nil -> nil
			return []*j.Statement{j.Nil()}
		}
		if gv.Literal != nil {
			if t, hasType := v.WrappedGolangType(); hasType {
				// &int(123) -> ToPtr(int(123))
				return []*j.Statement{j.Qual(ctx.RuntimePackage(""), "ToPtr").Call(
					j.Add(utils.ToCode(t.RenderUsage(ctx))...).Call(valueStmt),
				)}
			}
			// &123 -> ToPtr(123)
			return []*j.Statement{j.Qual(ctx.RuntimePackage(""), "ToPtr").Call(valueStmt)}
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

func (gv GoValue) String() string {
	switch {
	case gv.Literal != nil:
		return fmt.Sprintf("%v", gv.Literal)
	case gv.StructVals.Len() > 0:
		return fmt.Sprintf("{%v...}", lo.Slice(gv.StructVals.Entries(), 0, 2))
	case gv.DictVals.Len() > 0:
		return fmt.Sprintf("{%v...}", lo.Slice(gv.DictVals.Entries(), 0, 2))
	case gv.ArrayVals != nil:
		return fmt.Sprintf("{%v...}", lo.Slice(gv.ArrayVals, 0, 2))
	}
	return "nil"
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
			res.Type = &Array{
				BaseType: BaseType{Name: rtyp.Name(), PackageName: rtyp.PkgPath()},
				ItemsType: &Simple{
					Name:    elemType.Name(),
					IsIface: elemType.Kind() == reflect.Interface,
					Package: elemType.PkgPath(),
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
			res.Type = &Map{
				BaseType: BaseType{Name: rtyp.Name(), PackageName: rtyp.PkgPath()},
				KeyType: &Simple{
					Name:    keyType.Name(),
					IsIface: keyType.Kind() == reflect.Interface,
					Package: keyType.PkgPath(),
				},
				ValueType: &Simple{
					Name:    elemType.Name(),
					IsIface: elemType.Kind() == reflect.Interface,
					Package: elemType.PkgPath(),
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
			res.Type = &Struct{
				BaseType: BaseType{Name: rtyp.Name(), PackageName: rtyp.PkgPath()},
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
		return &GoValue{Literal: val, Type: &Pointer{Type: overrideType}}
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return &GoValue{Literal: rval.Interface(), Type: overrideType}
	}

	panic(fmt.Sprintf("Cannot construct Value from a value of type %T", value))
}
