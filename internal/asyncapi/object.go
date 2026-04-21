package asyncapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/types"

	yaml "gopkg.in/yaml.v3"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"

	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

// Object describes the [JSON Schema Specification Draft 07]
//
// [JSON Schema Specification Draft 07]: https://json-schema.org/specification-links#draft-7
type Object struct {
	Type                 *types.Union2[string, []string]            `json:"type,omitzero" yaml:"type"`
	AdditionalItems      *types.Union2[Object, bool]                `json:"additionalItems,omitzero" yaml:"additionalItems"`
	AdditionalProperties *types.Union2[Object, bool]                `json:"additionalProperties,omitzero" yaml:"additionalProperties"`
	AllOf                []Object                                   `json:"allOf,omitzero" yaml:"allOf" cgen:"selectable"`
	AnyOf                []Object                                   `json:"anyOf,omitzero" yaml:"anyOf" cgen:"selectable"`
	Const                *types.Union2[json.RawMessage, yaml.Node]  `json:"const,omitzero" yaml:"const"`
	Contains             *Object                                    `json:"contains,omitzero" yaml:"contains"`
	Default              *types.Union2[json.RawMessage, yaml.Node]  `json:"default,omitzero" yaml:"default"`
	Definitions          types.OrderedMap[string, Object]           `json:"definitions,omitzero" yaml:"definitions"`
	Deprecated           *bool                                      `json:"deprecated,omitzero" yaml:"deprecated"`
	Description          string                                     `json:"description,omitzero" yaml:"description"`
	Discriminator        string                                     `json:"discriminator,omitzero" yaml:"discriminator"`
	Else                 *Object                                    `json:"else,omitzero" yaml:"else"`
	Enum                 []any                                      `json:"enum,omitzero" yaml:"enum"`
	Examples             []types.Union2[json.RawMessage, yaml.Node] `json:"examples,omitzero" yaml:"examples"`
	ExclusiveMaximum     *types.Union2[bool, json.Number]           `json:"exclusiveMaximum,omitzero" yaml:"exclusiveMaximum"`
	ExclusiveMinimum     *types.Union2[bool, json.Number]           `json:"exclusiveMinimum,omitzero" yaml:"exclusiveMinimum"`
	ExternalDocs         *ExternalDocumentation                     `json:"externalDocs,omitzero" yaml:"externalDocs"`
	Format               string                                     `json:"format,omitzero" yaml:"format"`
	If                   *Object                                    `json:"if,omitzero" yaml:"if"`
	Items                *types.Union2[Object, []Object]            `json:"items,omitzero" yaml:"items"`
	MaxItems             *int                                       `json:"maxItems,omitzero" yaml:"maxItems"`
	MaxLength            *int                                       `json:"maxLength,omitzero" yaml:"maxLength"`
	MaxProperties        *int                                       `json:"maxProperties,omitzero" yaml:"maxProperties"`
	Maximum              *json.Number                               `json:"maximum,omitzero" yaml:"maximum"`
	MinItems             *int                                       `json:"minItems,omitzero" yaml:"minItems"`
	MinLength            *int                                       `json:"minLength,omitzero" yaml:"minLength"`
	MinProperties        *int                                       `json:"minProperties,omitzero" yaml:"minProperties"`
	Minimum              *json.Number                               `json:"minimum,omitzero" yaml:"minimum"`
	MultipleOf           *json.Number                               `json:"multipleOf,omitzero" yaml:"multipleOf"`
	Not                  *Object                                    `json:"not,omitzero" yaml:"not"`
	OneOf                []Object                                   `json:"oneOf,omitzero" yaml:"oneOf" cgen:"selectable"`
	Pattern              string                                     `json:"pattern,omitzero" yaml:"pattern"`
	PatternProperties    types.OrderedMap[string, Object]           `json:"patternProperties,omitzero" yaml:"patternProperties"` // Mapping regex->schema
	Properties           types.OrderedMap[string, Object]           `json:"properties,omitzero" yaml:"properties"`
	PropertyNames        *Object                                    `json:"propertyNames,omitzero" yaml:"propertyNames"`
	ReadOnly             *bool                                      `json:"readOnly,omitzero" yaml:"readOnly"`
	Required             []string                                   `json:"required,omitzero" yaml:"required"`
	Then                 *Object                                    `json:"then,omitzero" yaml:"then"`
	Title                string                                     `json:"title,omitzero" yaml:"title"`
	UniqueItems          *bool                                      `json:"uniqueItems,omitzero" yaml:"uniqueItems"`

	XNullable     *bool                                                     `json:"x-nullable,omitzero" yaml:"x-nullable"`
	XGoType       *types.Union2[string, xGoType]                            `json:"x-go-type,omitzero" yaml:"x-go-type"`
	XGoName       string                                                    `json:"x-go-name,omitzero" yaml:"x-go-name"`
	XGoTags       *types.Union2[[]string, types.OrderedMap[string, string]] `json:"x-go-tags,omitzero" yaml:"x-go-tags"`
	XGoTagsValues []string                                                  `json:"x-go-tags-values,omitzero" yaml:"x-go-tags-values"`
	XIgnore       bool                                                      `json:"x-ignore,omitzero" yaml:"x-ignore"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

func (o Object) Compile(ctx *compile.Context) error {
	obj, err := o.build(ctx, ctx.Stack.Top().Flags, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (o Object) build(ctx *compile.Context, flags map[common.SchemaTag]string, objectKey string) (common.Artifact, error) {
	_, isSelectable := flags[common.SchemaTagSelectable]
	ignore := o.XIgnore
	if ignore {
		ctx.Logger.Debug("Object denoted to be ignored")
		return &lang.GoSimple{TypeName: "any", IsInterface: true, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}, nil
	}
	if o.Ref != "" {
		refName := objectKey
		// Ignore the objectKey in definitions other than `components.schemas`, generate a unique name instead
		if !isSelectable {
			refName = ctx.GenerateObjName("", "")
		}

		return registerRef(ctx, o.Ref, refName, &isSelectable), nil
	}

	if o.Type == nil {
		ctx.Logger.Warn("Empty object type is deprecated, guessing it automatically. Hint: probably you wrote `type: null` instead of `type: \"null\"`?")
		o.Type = o.guessObjectType(ctx)
	}

	if len(o.OneOf)+len(o.AnyOf)+len(o.AllOf) > 0 {
		ctx.Logger.Trace("Object", "type", "union")
		return o.buildUnionStruct(ctx, flags) // TODO: process other items that can be set along with oneof/anyof/allof
	}

	typeName, nullable, err := o.getTypeName(ctx)
	if err != nil {
		return nil, err
	}

	// One type: { "type": "something" }
	golangType, err := o.buildGolangType(ctx, flags, typeName)
	if err != nil {
		return nil, err
	}

	nullable = nullable || lo.FromPtr(o.XNullable)
	if nullable {
		ctx.Logger.Trace("Object is nullable, making it pointer")
		golangType = &lang.GoPointer{Type: golangType}
	}

	return golangType, nil
}

// guessObjectType is backwards compatible, guessing the user intention when they didn't specify a type.
func (o Object) guessObjectType(ctx *compile.Context) *types.Union2[string, []string] {
	switch {
	case o.Ref == "" && o.Properties.Len() > 0:
		ctx.Logger.Trace("Determined `type: object` because of `properties` presence")
		return types.ToUnion2[string, []string]("object")
	case o.Items != nil: // TODO: fix type when AllOf, AnyOf, OneOf
		ctx.Logger.Trace("Determined `type: array` because of `items` presence")
		return types.ToUnion2[string, []string]("array")
	default:
		ctx.Logger.Trace("Determined `type: object` as a default object type")
		return types.ToUnion2[string, []string]("object")
	}
}

// getTypeName returns the jsonschema type name of the object. It also returns whether the object is nullable.
func (o Object) getTypeName(ctx *compile.Context) (typeName string, nullable bool, err error) {
	schemaType := o.Type
	typeName = schemaType.V0

	if schemaType.Selector == 1 { // Multiple types, e.g. { "type": [ "object", "array", "null" ] }
		nullable = lo.Contains(schemaType.V1, "null")
		typs := lo.Reject(schemaType.V1, func(item string, _ int) bool { return item == "null" }) // Throw out null (if any)

		switch {
		case len(typs) > 1: // More than one type along with null -> 'any'
			typeName = ""
		case len(typs) == 1: // One type along with null -> pointer to this type
			typeName = typs[0]
		case nullable: // Null only -> 'any', that can be only nil
			typeName = "null"
		default:
			err = types.CompileError{Err: errors.New("empty object type"), Path: ctx.CurrentRefPointer()}
			return
		}
		ctx.Logger.Trace(fmt.Sprintf("Multitype object type inferred as %q", typeName))
	}
	return
}

func (o Object) buildGolangType(ctx *compile.Context, flags map[common.SchemaTag]string, typeName string) (finalType common.GolangType, err error) {
	if o.XGoType != nil {
		replaceType := o.XGoType.Selector == 0 && o.XGoType.V0 != "" || o.XGoType.Selector == 1 && o.XGoType.V1.Type != ""
		if replaceType {
			f := o.buildXGoType(ctx)

			ctx.Logger.Trace("Object with replaced type using x-go-type", "type", f.String())
			return f, nil
		}
	}

	var wrappedType common.GolangType
	switch typeName {
	case "null", "":
		ctx.Logger.Trace("Object", "type", "any")
		finalType = &lang.GoSimple{TypeName: "any", IsInterface: true, OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
		if len(o.Enum) > 0 {
			ctx.Logger.Info("Ignoring object's enums because of null/empty type")
		}
	case "object":
		ctx.Logger.Trace("Object", "type", "struct")
		ctx.Logger.NextCallLevel()
		finalType, err = o.buildLangObject(ctx, flags)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
	case "array":
		ctx.Logger.Trace("Object", "type", "array")
		ctx.Logger.NextCallLevel()
		finalType, err = o.buildLangArray(ctx, flags)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
	case "boolean":
		ctx.Logger.Trace("Object", "type", "bool")
		wrappedType = &lang.GoSimple{TypeName: "bool", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	case "integer":
		ctx.Logger.Trace("Object", "type", "int")
		wrappedType = &lang.GoSimple{TypeName: "int", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	case "number":
		ctx.Logger.Trace("Object", "type", "float64")
		wrappedType = &lang.GoSimple{TypeName: "float64", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	case "string":
		ctx.Logger.Trace("Object", "type", "string")
		wrappedType = &lang.GoSimple{TypeName: "string", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	default:
		return nil, types.CompileError{Err: fmt.Errorf("unknown jsonschema type %q", typeName), Path: ctx.CurrentRefPointer()}
	}

	if wrappedType != nil {
		_, isSelectable := flags[common.SchemaTagSelectable]
		finalType = &lang.GoTypeDefinition{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(o.Title, ""),
				Description:   o.Description,
				HasDefinition: isSelectable,
				ArtifactKind:  lo.Ternary(isSelectable, common.ArtifactKindSchema, common.ArtifactKindOther),
			},
			WrappedType: wrappedType,
		}
	}

	if len(o.Enum) > 0 {
		if !finalType.Selectable() {
			ctx.Logger.Info("Ignoring enum for inlined jsonschema object. Hint: move it into a separate definition under components.schemas section and reference it with $ref to make enums work")
		} else {
			ctx.Logger.Trace("Object has enums, wrapping it into GoEnum")
			typ := &lang.GoEnum{WrappedType: finalType}
			ctx.Logger.NextCallLevel()
			typ.PrimitiveEnums, typ.ComplexEnums, err = o.getEnums(ctx, typeName)
			ctx.Logger.PrevCallLevel()
			if err != nil {
				return nil, types.CompileError{Err: err, Path: ctx.CurrentRefPointer()}
			}
			finalType = typ
		}
	}

	return finalType, nil
}

func (o Object) buildLangObject(ctx *compile.Context, flags map[common.SchemaTag]string) (common.GolangType, error) {
	_, isSelectable := flags[common.SchemaTagSelectable]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	if o.Properties.Len() == 0 {
		ctx.Logger.Debug("Object with empty properties, generating a map")
		return &lang.GoMap{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(objName, ""),
				Description:   o.Description,
				HasDefinition: isSelectable,
				ArtifactKind:  lo.Ternary(isSelectable, common.ArtifactKindSchema, common.ArtifactKindOther),
			},
			KeyType:   &lang.GoSimple{TypeName: "string"},
			ValueType: &lang.GoSimple{TypeName: "any", IsInterface: true},
		}, nil
	}

	res := lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  ctx.GenerateObjName(objName, ""),
			Description:   o.Description,
			HasDefinition: isSelectable,
			ArtifactKind:  lo.Ternary(isSelectable, common.ArtifactKindSchema, common.ArtifactKindOther),
		},
		StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx),
	}

	var contentTypesFunc func() []string
	_, isDataModel := flags[common.SchemaTagDataModel]
	if isDataModel {
		ctx.Logger.Trace("Object struct is data model")
		messagesPrm := lang.NewListCbPromise[*render.Message](func(item common.Artifact) bool {
			_, ok := item.(*render.Message)
			return ok
		}, nil)
		ctx.PutListPromise(messagesPrm)
		complexEnumsPrm := lang.NewListCbPromise[*lang.GoEnum](func(item common.Artifact) bool {
			v, ok := item.(*lang.GoEnum)
			return ok && v.ComplexEnums.Len() > 0
		}, nil)
		ctx.PutListPromise(complexEnumsPrm)
		contentTypesFunc = func() []string {
			tagNames := lo.Map(messagesPrm.T(), func(item *render.Message, _ int) string {
				return guessTagByContentType(item.EffectiveContentType())
			})
			if len(complexEnumsPrm.T()) > 0 {
				// Forcibly add "json" field tag to *all* generated models if at least one enum in document has object value.
				// We need json in the model and its inner models because such enums are initialized in the generated code
				// by unmarshalling them from JSON automatically.
				// Another way could be is to track affected models using CompileContext stack, but it would be slightly
				// complicated implementation, and also object values in enums is pretty rare case.
				// But this can be implemented if any issues will arise because of current approach.
				tagNames = append(tagNames, "json")
			}
			tagNames = lo.Uniq(tagNames)
			slices.Sort(tagNames)
			return tagNames
		}
	}

	// regular properties
	for k, v := range o.Properties.Entries() {
		ctx.Logger.Trace("Object property", "name", k)
		ref := ctx.CurrentRefPointer("properties", k)
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)

		var langObj common.GolangType = prm
		if lo.Contains(o.Required, k) {
			langObj = &lang.GoPointer{Type: langObj}
		}

		propName, _ := lo.Coalesce(v.XGoName, k)
		f := lang.GoStructField{
			OriginalName:     utils.ToGolangName(propName, true),
			MarshalName:      k,
			Description:      v.Description,
			Type:             langObj,
			ContentTypesFunc: contentTypesFunc,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	if o.AdditionalProperties != nil {
		propName, _ := lo.Coalesce(o.AdditionalProperties.V0.XGoName, o.Title)
		switch o.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			ctx.Logger.Trace("Object additional properties", "type", "object")
			ref := ctx.CurrentRefPointer("additionalProperties")
			prm := lang.NewGolangTypePromise(ref, nil)
			ctx.PutPromise(prm)
			f := lang.GoStructField{
				OriginalName: "AdditionalProperties",
				Description:  o.AdditionalProperties.V0.Description,
				Type: &lang.GoMap{
					BaseType: lang.BaseType{
						OriginalName:  ctx.GenerateObjName(propName, "AdditionalProperties"),
						Description:   o.AdditionalProperties.V0.Description,
						HasDefinition: false,
					},
					KeyType:               &lang.GoSimple{TypeName: "string"},
					ValueType:             prm,
					StructFieldRenderInfo: o.AdditionalProperties.V0.getStructFieldRenderInfo(ctx),
				},
			}
			res.Fields = append(res.Fields, f)
		case 1:
			ctx.Logger.Trace("Object additional properties", "type", "boolean")
			if o.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := lang.GoTypeDefinition{
					BaseType: lang.BaseType{
						OriginalName:  ctx.GenerateObjName(propName, "AdditionalPropertiesValue"),
						Description:   "",
						HasDefinition: false,
					},
					WrappedType: &lang.GoSimple{TypeName: "any", IsInterface: true},
				}
				f := lang.GoStructField{
					OriginalName: "AdditionalProperties",
					Type: &lang.GoMap{
						BaseType: lang.BaseType{
							OriginalName:  ctx.GenerateObjName(propName, "AdditionalProperties"),
							Description:   "",
							HasDefinition: false,
						},
						KeyType:   &lang.GoSimple{TypeName: "string"},
						ValueType: &valTyp,
					},
					ContentTypesFunc: contentTypesFunc,
				}
				res.Fields = append(res.Fields, f)
			}
		}
	}

	return &res, nil
}

func (o Object) buildLangArray(ctx *compile.Context, flags map[common.SchemaTag]string) (*lang.GoArray, error) {
	_, isSelectable := flags[common.SchemaTagSelectable]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.GoArray{
		BaseType: lang.BaseType{
			OriginalName:  ctx.GenerateObjName(objName, ""),
			Description:   o.Description,
			HasDefinition: isSelectable,
			ArtifactKind:  lo.Ternary(isSelectable, common.ArtifactKindSchema, common.ArtifactKindOther),
		},
		ItemsType:             nil,
		StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx),
	}

	switch {
	case o.Items == nil || o.Items.Selector == 1:
		// Items with unknown type or tuple.
		// It's hard to express the tuple in Go (i.e. a type that is natively marshaled/unmarshalled to array and capable
		// of holding items of different types), so we generate []any.
		// User may specify x-go-type to set a more specific type.
		ctx.Logger.Trace("Object items", "items", lo.Ternary(o.Items == nil, "none", "tuple"))
		res.ItemsType = &lang.GoSimple{TypeName: "any", IsInterface: true}
		hasAdditionalItems := o.AdditionalItems != nil && (o.AdditionalItems.Selector == 0 || o.AdditionalItems.V1)
		if o.Items != nil && !hasAdditionalItems {
			res.Size = len(o.Items.V1)
		}
	case o.Items.Selector == 0: // All items have the same schema
		ctx.Logger.Trace("Object items", "items", "one")
		ref := ctx.CurrentRefPointer("items")
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)
		res.ItemsType = prm
	}

	return &res, nil
}

func (o Object) buildUnionStruct(ctx *compile.Context, flags map[common.SchemaTag]string) (*lang.UnionStruct, error) {
	_, isSelectable := flags[common.SchemaTagSelectable]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.UnionStruct{
		GoStruct: lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(objName, ""),
				Description:   o.Description,
				HasDefinition: isSelectable,
				ArtifactKind:  lo.Ternary(isSelectable, common.ArtifactKindSchema, common.ArtifactKindOther),
			},
			StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx),
		},
	}

	// Collect all messages to retrieve struct field tags
	messagesPrm := lang.NewListCbPromise[*render.Message](func(item common.Artifact) bool {
		_, ok := item.(*render.Message)
		return ok
	}, nil)
	ctx.PutListPromise(messagesPrm)

	res.Fields = lo.Times(len(o.OneOf), func(index int) lang.GoStructField {
		ref := ctx.CurrentRefPointer("oneOf", strconv.Itoa(index))
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)
		return lang.GoStructField{Type: &lang.GoPointer{Type: prm}}
	})
	res.Fields = append(res.Fields, lo.Times(len(o.AnyOf), func(index int) lang.GoStructField {
		ref := ctx.CurrentRefPointer("anyOf", strconv.Itoa(index))
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)
		return lang.GoStructField{Type: &lang.GoPointer{Type: prm}}
	})...)
	res.Fields = append(res.Fields, lo.Times(len(o.AllOf), func(index int) lang.GoStructField {
		ref := ctx.CurrentRefPointer("allOf", strconv.Itoa(index))
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)
		return lang.GoStructField{Type: prm}
	})...)

	return &res, nil
}

// buildXGoType builds a GolangType from x-go-type field value
func (o Object) buildXGoType(ctx *compile.Context) (golangType common.GolangType) {
	t := &lang.GoSimple{StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}

	switch o.XGoType.Selector {
	case 0:
		t.TypeName = o.XGoType.V0
	case 1:
		t.TypeName = o.XGoType.V1.Type
		t.Import = o.XGoType.V1.Import.Package
		t.IsInterface = o.XGoType.V1.Hint.Kind == "interface"

		if o.XGoType.V1.Hint.Pointer {
			return &lang.GoPointer{Type: t}
		}
	}

	golangType = t
	return
}

func (o Object) getStructFieldRenderInfo(ctx *compile.Context) lang.StructFieldRenderInfo {
	res := lang.StructFieldRenderInfo{
		IsEmbeddedType: o.XGoType != nil && o.XGoType.Selector == 1 && o.XGoType.V1.Embedded,
	}
	if o.XGoTags != nil {
		switch o.XGoTags.Selector {
		case 0:
			res.TagNames = o.XGoTags.V0
			ctx.Logger.Trace("Extra tags", "names", res.TagNames)
		case 1:
			res.Tags = o.XGoTags.V1
			ctx.Logger.Trace("Extra tags", "tags", maps.Collect(res.Tags.Entries()))
		}
	}
	if res.TagValues = o.XGoTagsValues; len(res.TagValues) > 0 {
		ctx.Logger.Trace("Extra tags values", "values", res.TagValues)
	}

	return res
}

func (o Object) getEnums(ctx *compile.Context, typeName string) (primitiveEnums, complexEnums types.OrderedMap[string, any], err error) {
	suffixes := make(map[string]int)

	getUniqueSuffix := func(suffix string) string {
		suffix = utils.ToGolangNameSuffix(suffix)
		if count, ok := suffixes[suffix]; ok {
			suffixes[suffix] = count + 1
			return fmt.Sprintf("%s_%d", suffix, count+1)
		}
		suffixes[suffix] = 0
		return suffix
	}

	for i, item := range o.Enum {
		rval := reflect.ValueOf(item)
		kind := rval.Kind()
		isInt := lo.Contains([]reflect.Kind{
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		}, kind)
		ctx.Logger.Trace("Enum", "index", i, "value", item, "type", fmt.Sprintf("%T", item), "reflect_kind", kind)

		switch {
		case item == nil:
			ctx.Logger.Debug("Skipping null enum", "index", i)
			continue
		case typeName == "boolean":
			ctx.Logger.Debug("Skipping enum for boolean schema", "index", i, "value", item)
			continue
		case (typeName == "integer" || typeName == "number") && isInt:
			primitiveEnums.Set(getUniqueSuffix(fmt.Sprintf("%d", i)), item)
		case typeName == "number" && (kind == reflect.Float32 || kind == reflect.Float64):
			primitiveEnums.Set(getUniqueSuffix(strings.TrimRight(fmt.Sprintf("%.f", item), "0")), item)
		case typeName == "string" && kind == reflect.String:
			primitiveEnums.Set(getUniqueSuffix(fmt.Sprintf("%s", item)), item)
		case typeName == "integer" && (kind == reflect.Float32 || kind == reflect.Float64):
			// json.Unmarshal unmarshals integers into float64, so this is the most common case for integers
			primitiveEnums.Set(getUniqueSuffix(fmt.Sprintf("%d", item)), item)
		case typeName == "object" && kind == reflect.Map:
			fallthrough
		case typeName == "array" && (kind == reflect.Slice || kind == reflect.Array):
			fallthrough
		case typeName == "":
			complexEnums.Set(fmt.Sprintf("Enum%d", i+1), item)
		default:
			ctx.Logger.Warn("Type mismatch between enum value and schema, skipping it", "enum", fmt.Sprintf("%[1]T(%[1]v)", item), "schema_type", typeName)
		}
	}

	return
}
