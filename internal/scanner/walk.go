package scanner

import (
	"reflect"
	"strconv"
	"strings"
)

const tagName = "cgen"

type builder interface {
	Build(ctx *Context) error
}

func WalkSchema(ctx *Context, object reflect.Value) error {
	objectTyp := object.Type()

	gather := func(_ctx *Context, _obj reflect.Value) error {
		if v, ok := _obj.Interface().(builder); ok {
			if (_obj.Kind() == reflect.Pointer || _obj.Kind() == reflect.Interface) && _obj.IsNil() {
				return nil
			}
			if e := v.Build(_ctx); e != nil {
				return e
			}
		}
		return WalkSchema(_ctx, _obj)
	}

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
		iter := object.MapRange() // TODO: keep the same order
		for iter.Next() {
			pushStack(ctx, iter.Key().String(), ctx.Top().Flags)
			if err := gather(ctx, iter.Value()); err != nil {
				return err
			}
			ctx.Pop()
		}
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

func parseTags(field reflect.StructField) (tags map[string]string) {
	tagVal, ok := field.Tag.Lookup(tagName)
	if !ok {
		return nil
	}

	tags = make(map[string]string)
	for _, part := range strings.Split(tagVal, ",") {
		if part == "" {
			continue
		}
		part = strings.Trim(part, " ")
		k, v, _ := strings.Cut(part, "=")
		tags[strings.Trim(k, " ")] = strings.Trim(v, " '")
	}
	return
}

func pushStack(ctx *Context, pathItem string, flags map[string]string) {
	if flags == nil {
		flags = make(map[string]string)
	}
	item := ContextStackItem{
		Path:  pathItem,
		Flags: flags,
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
