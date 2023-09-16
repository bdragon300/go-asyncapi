package scan

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
)

type compiler interface {
	Compile(ctx *common.CompileContext) error
}

type orderedMap interface {
	OrderedMap()
}

func WalkSchema(ctx *common.CompileContext, object reflect.Value) error {
	gather := func(_ctx *common.CompileContext, _obj reflect.Value) error {
		// BFS tree traversal
		if v, ok := _obj.Interface().(compiler); ok {
			if (_obj.Kind() == reflect.Pointer || _obj.Kind() == reflect.Interface) && _obj.IsNil() {
				return nil
			}
			if e := v.Compile(_ctx); e != nil {
				return e
			}
		}
		return WalkSchema(_ctx, _obj)
	}

	if _, ok := object.Interface().(orderedMap); ok {
		mKeys := object.MethodByName("Keys")
		keys := mKeys.Call(nil)[0]
		mMustGet := object.MethodByName("MustGet")
		for j := 0; j < keys.Len(); j++ {
			key := keys.Index(j)
			val := mMustGet.Call([]reflect.Value{key})[0]
			pushStack(ctx, key.String(), ctx.Stack.Top().Flags)
			if err := gather(ctx, val); err != nil {
				return err
			}
			ctx.Stack.Pop()
		}
	}
	// TODO: add Unions

	if object.Kind() == reflect.Pointer {
		object = reflect.Indirect(object)
	}

	switch object.Kind() {
	case reflect.Struct:
		objectTyp := object.Type()
		for i := 0; i < object.NumField(); i++ {
			fld := objectTyp.Field(i)
			fldVal := object.Field(i)
			if !fld.IsExported() || fld.Anonymous {
				continue
			}

			pushStack(ctx, getFieldJSONName(fld), parseTags(fld))
			if err := gather(ctx, fldVal); err != nil {
				return err
			}
			ctx.Stack.Pop()
		}
	case reflect.Map:
		panic("Use OrderedMap instead to keep schema definitions order!")
	case reflect.Array, reflect.Slice:
		for j := 0; j < object.Len(); j++ {
			pushStack(ctx, strconv.Itoa(j), ctx.Stack.Top().Flags)
			if err := gather(ctx, object.Index(j)); err != nil {
				return err
			}
			ctx.Stack.Pop()
		}
	}

	return nil
}

func parseTags(field reflect.StructField) (tags map[common.SchemaTag]string) {
	tagVal, ok := field.Tag.Lookup(common.TagName)
	if !ok {
		return nil
	}

	tags = make(map[common.SchemaTag]string)
	for _, part := range strings.Split(tagVal, ",") {
		if part == "" {
			continue
		}
		part = strings.Trim(part, " ")
		k, v, _ := strings.Cut(part, "=")
		tags[common.SchemaTag(strings.Trim(k, " "))] = strings.Trim(v, " '")
	}
	return
}

func pushStack(ctx *common.CompileContext, pathItem string, flags map[common.SchemaTag]string) {
	if flags == nil {
		flags = make(map[common.SchemaTag]string)
	}
	pkgKind := common.RuntimePackageKind
	if len(ctx.Stack.Items()) > 0 {
		pkgKind = ctx.Stack.Top().PackageKind
	}
	if v, ok := flags[common.SchemaTagPackageDown]; ok {
		pkgKind = common.PackageKind(v)
	}
	item := common.ContextStackItem{
		Path:        pathItem,
		Flags:       flags,
		PackageKind: pkgKind,
		ObjName:     "",
	}
	ctx.Stack.Push(item)
}

func getFieldJSONName(f reflect.StructField) string {
	if tagVal, ok := f.Tag.Lookup("json"); ok {
		parts := strings.Split(tagVal, ",")
		if len(parts) == 0 || parts[0] == "-" {
			return ""
		}
		return parts[0]
	}
	return f.Name
}

