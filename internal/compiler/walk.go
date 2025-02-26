package compiler

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

type compiledObject interface {
	Compile(ctx *common.CompileContext) error
}

type orderedMap interface {
	OrderedMap()
}

type union interface {
	CurrentValue() any
}

// WalkAndCompile recursively walks through the object and run the compilation logic on each one.
//
// The main aim of this function is to traverse the object and call the Compile method on every object that have it.
// We use the BFS tree traversal algorithm here. Once any Compile call returns an error, the function stops immediately
// and returns this error.
//
// Additionally, the function keeps the compile context up-to-date, maintaining the document path stack and field's
// flags (tags).
func WalkAndCompile(ctx *common.CompileContext, object reflect.Value) error {
	// Special types
	switch v := object.Interface().(type) {
	case orderedMap:
		mEntries := object.MethodByName("Entries")
		entries := mEntries.Call(nil)[0]
		for j := 0; j < entries.Len(); j++ {
			entry := entries.Index(j)
			pushStack(ctx, entry.FieldByName("Key").String(), ctx.Stack.Top().Flags)
			err := traverse(ctx, entry.FieldByName("Value"))
			ctx.Stack.Pop()
			if err != nil {
				return err
			}
		}
		return nil
	case union:
		if (object.Kind() == reflect.Pointer || object.Kind() == reflect.Interface) && object.IsNil() {
			return nil
		}
		if err := traverse(ctx, reflect.ValueOf(v.CurrentValue())); err != nil {
			return err
		}
		return nil
	}

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

// traverse calls the object.Compile method if it has it and runs WalkAndCompile on the object.
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

// parseTags parses the tool's Go struct tag expression into map of tags. Returns nil if tool's tag has not found or no
// Go struct tags defined for the field.
//
// E.g. `json:"name,omitempty" cgen:"foo=bar,baz"` gives `{"foo": "bar", "baz": ""}`.
func parseTags(field reflect.StructField) (tags map[common.SchemaTag]string) {
	tagVal, ok := field.Tag.Lookup(common.SchemaTagName)
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

	// Inherited tags
	if len(ctx.Stack.Items()) > 0 {
		if _, ok := ctx.Stack.Top().Flags[common.SchemaTagDataModel]; ok {
			flags[common.SchemaTagDataModel] = ctx.Stack.Top().Flags[common.SchemaTagDataModel]
		}
	}
	item := common.DocumentTreeItem{
		Key:   pathItem,
		Flags: flags,
	}
	ctx.Stack.Push(item)
}

// getFieldJSONName returns the JSON marshaller field name from the struct field.
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
