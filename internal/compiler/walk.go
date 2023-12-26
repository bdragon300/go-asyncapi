package compiler

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
)

type compiledObject interface {
	Compile(ctx *common.CompileContext) error
}

type orderedMap interface {
	OrderedMap()
}

func WalkAndCompile(ctx *common.CompileContext, object reflect.Value) error {
	if _, ok := object.Interface().(orderedMap); ok {
		mEntries := object.MethodByName("Entries")
		entries := mEntries.Call(nil)[0]
		for j := 0; j < entries.Len(); j++ {
			entry := entries.Index(j)
			pushStack(ctx, entry.FieldByName("Key").String(), ctx.Stack.Top().Flags)
			if err := traverse(ctx, entry.FieldByName("Value")); err != nil {
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
			if err := traverse(ctx, fldVal); err != nil {
				return err
			}
			ctx.Stack.Pop()
		}
	case reflect.Map:
		panic("Use OrderedMap instead to keep schema definitions order!")
	case reflect.Array, reflect.Slice:
		for j := 0; j < object.Len(); j++ {
			pushStack(ctx, strconv.Itoa(j), ctx.Stack.Top().Flags)
			if err := traverse(ctx, object.Index(j)); err != nil {
				return err
			}
			ctx.Stack.Pop()
		}
	}

	return nil
}

func traverse(ctx *common.CompileContext, object reflect.Value) error {
	// BFS tree traversal
	if v, ok := object.Interface().(compiledObject); ok {
		if (object.Kind() == reflect.Pointer || object.Kind() == reflect.Interface) && object.IsNil() {
			return nil
		}
		ctx.Logger.Debug(reflect.Indirect(object).Type().Name())
		ctx.Logger.NextCallLevel()
		if e := v.Compile(ctx); e != nil {
			ctx.Logger.Fatal("Compilation error", e)
			return e
		}
		ctx.Logger.PrevCallLevel()
	}
	return WalkAndCompile(ctx, object)
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
	var pkgName string
	if len(ctx.Stack.Items()) > 0 {
		pkgName = ctx.TopPackageName()
	}
	if v, ok := flags[common.SchemaTagPackageDown]; ok {
		pkgName = v
	}
	item := common.ContextStackItem{
		Path:        pathItem,
		Flags:       flags,
		PackageName: pkgName,
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