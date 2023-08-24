package scan

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
)

const tagName = "cgen"

type SchemaTag string

const (
	SchemaTagNoInline    SchemaTag = "noinline"
	SchemaTagPackageDown SchemaTag = "packageDown"
)

type builder interface {
	Build(ctx *Context) error
}

type orderedMap interface {
	OrderedMap()
}

func WalkSchema(ctx *Context, object reflect.Value) error {
	objectTyp := object.Type()

	gather := func(_ctx *Context, _obj reflect.Value) error {
		if err := WalkSchema(_ctx, _obj); err != nil {
			return err
		}
		if v, ok := _obj.Interface().(builder); ok {
			if (_obj.Kind() == reflect.Pointer || _obj.Kind() == reflect.Interface) && _obj.IsNil() {
				return nil
			}
			if e := v.Build(_ctx); e != nil {
				return e
			}
		}
		return nil
	}

	if _, ok := object.Interface().(orderedMap); ok {
		mKeys := object.MethodByName("Keys")
		keys := mKeys.Call(nil)[0]
		mMustGet := object.MethodByName("MustGet")
		for j := 0; j < keys.Len(); j++ {
			key := keys.Index(j)
			val := mMustGet.Call([]reflect.Value{key})[0]
			pushStack(ctx, key.String(), ctx.Top().Flags)
			if err := gather(ctx, val); err != nil {
				return err
			}
			ctx.Pop()
		}
	}
	// TODO: add Unions

	switch object.Kind() {
	case reflect.Struct:
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
			ctx.Pop()
		}
	case reflect.Map:
		panic("Use OrderedMap instead to keep schema definitions order!")
	case reflect.Array, reflect.Slice:
		for j := 0; j < object.Len(); j++ {
			pushStack(ctx, strconv.Itoa(j), ctx.Top().Flags)
			if err := gather(ctx, object.Index(j)); err != nil {
				return err
			}
			ctx.Pop()
		}
	}

	return nil
}

func parseTags(field reflect.StructField) (tags map[SchemaTag]string) {
	tagVal, ok := field.Tag.Lookup(tagName)
	if !ok {
		return nil
	}

	tags = make(map[SchemaTag]string)
	for _, part := range strings.Split(tagVal, ",") {
		if part == "" {
			continue
		}
		part = strings.Trim(part, " ")
		k, v, _ := strings.Cut(part, "=")
		tags[SchemaTag(strings.Trim(k, " "))] = strings.Trim(v, " '")
	}
	return
}

func pushStack(ctx *Context, pathItem string, flags map[SchemaTag]string) {
	if flags == nil {
		flags = make(map[SchemaTag]string)
	}
	pkgKind := common.DefaultPackageKind
	if len(ctx.Stack) > 0 {
		pkgKind = ctx.Top().PackageKind
	}
	if v, ok := flags[SchemaTagPackageDown]; ok {
		pkgKind = common.PackageKind(v)
	}
	item := ContextStackItem{
		Path:        pathItem,
		Flags:       flags,
		PackageKind: pkgKind,
	}
	ctx.Push(item)
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

