package asyncapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"

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
	Type                 *types.Union2[string, []string]            `json:"type" yaml:"type"`
	AdditionalItems      *types.Union2[Object, bool]                `json:"additionalItems" yaml:"additionalItems"`
	AdditionalProperties *types.Union2[Object, bool]                `json:"additionalProperties" yaml:"additionalProperties"`
	AllOf                []Object                                   `json:"allOf" yaml:"allOf" cgen:"definition"`
	AnyOf                []Object                                   `json:"anyOf" yaml:"anyOf" cgen:"definition"`
	Const                *types.Union2[json.RawMessage, yaml.Node]  `json:"const" yaml:"const"`
	Contains             *Object                                    `json:"contains" yaml:"contains"`
	Default              *types.Union2[json.RawMessage, yaml.Node]  `json:"default" yaml:"default"`
	Definitions          types.OrderedMap[string, Object]           `json:"definitions" yaml:"definitions"`
	Deprecated           *bool                                      `json:"deprecated" yaml:"deprecated"`
	Description          string                                     `json:"description" yaml:"description"`
	Discriminator        string                                     `json:"discriminator" yaml:"discriminator"`
	Else                 *Object                                    `json:"else" yaml:"else"`
	Enum                 []types.Union2[json.RawMessage, yaml.Node] `json:"enum" yaml:"enum"`
	Examples             []types.Union2[json.RawMessage, yaml.Node] `json:"examples" yaml:"examples"`
	ExclusiveMaximum     *types.Union2[bool, json.Number]           `json:"exclusiveMaximum" yaml:"exclusiveMaximum"`
	ExclusiveMinimum     *types.Union2[bool, json.Number]           `json:"exclusiveMinimum" yaml:"exclusiveMinimum"`
	ExternalDocs         *ExternalDocumentation                     `json:"externalDocs" yaml:"externalDocs"`
	Format               string                                     `json:"format" yaml:"format"`
	If                   *Object                                    `json:"if" yaml:"if"`
	Items                *types.Union2[Object, []Object]            `json:"items" yaml:"items"`
	MaxItems             *int                                       `json:"maxItems" yaml:"maxItems"`
	MaxLength            *int                                       `json:"maxLength" yaml:"maxLength"`
	MaxProperties        *int                                       `json:"maxProperties" yaml:"maxProperties"`
	Maximum              *json.Number                               `json:"maximum" yaml:"maximum"`
	MinItems             *int                                       `json:"minItems" yaml:"minItems"`
	MinLength            *int                                       `json:"minLength" yaml:"minLength"`
	MinProperties        *int                                       `json:"minProperties" yaml:"minProperties"`
	Minimum              *json.Number                               `json:"minimum" yaml:"minimum"`
	MultipleOf           *json.Number                               `json:"multipleOf" yaml:"multipleOf"`
	Not                  *Object                                    `json:"not" yaml:"not"`
	OneOf                []Object                                   `json:"oneOf" yaml:"oneOf" cgen:"definition"`
	Pattern              string                                     `json:"pattern" yaml:"pattern"`
	PatternProperties    types.OrderedMap[string, Object]           `json:"patternProperties" yaml:"patternProperties"` // Mapping regex->schema
	Properties           types.OrderedMap[string, Object]           `json:"properties" yaml:"properties"`
	PropertyNames        *Object                                    `json:"propertyNames" yaml:"propertyNames"`
	ReadOnly             *bool                                      `json:"readOnly" yaml:"readOnly"`
	Required             []string                                   `json:"required" yaml:"required"`
	Then                 *Object                                    `json:"then" yaml:"then"`
	Title                string                                     `json:"title" yaml:"title"`
	UniqueItems          *bool                                      `json:"uniqueItems" yaml:"uniqueItems"`

	XNullable     *bool                                                     `json:"x-nullable" yaml:"x-nullable"`
	XGoType       *types.Union2[string, xGoType]                            `json:"x-go-type" yaml:"x-go-type"`
	XGoName       string                                                    `json:"x-go-name" yaml:"x-go-name"`
	XGoTags       *types.Union2[[]string, types.OrderedMap[string, string]] `json:"x-go-tags" yaml:"x-go-tags"`
	XGoTagsValues []string                                                  `json:"x-go-tags-values" yaml:"x-go-tags-values"`
	XIgnore       bool                                                      `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
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
		ctx.Logger.Trace("Ref", "$ref", o.Ref)

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

	// TODO: "type": { "enum": [ "residential", "business" ] }
	// One type: { "type": "object" }
	golangType, err := o.buildGolangType(ctx, flags, typeName)
	if err != nil {
		return nil, err
	}

	nullable = nullable || lo.FromPtr(o.XNullable)
	if nullable {
		ctx.Logger.Trace("Object is nullable, make it pointer")
		golangType = &lang.GoPointer{Type: golangType}
	}

	return golangType, nil
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

func (o Object) buildGolangType(ctx *compile.Context, flags map[common.SchemaTag]string, typeName string) (golangType common.GolangType, err error) {
	var aliasedType *lang.GoSimple

	if o.XGoType != nil {
		replaceType := o.XGoType.Selector == 0 && o.XGoType.V0 != "" || o.XGoType.Selector == 1 && o.XGoType.V1.Type != ""
		if replaceType {
			f := o.buildXGoType(ctx)

			ctx.Logger.Trace("Object with replaced type using x-go-type", "type", f.String())
			return f, nil
		}
	}

	if typeName == "object" {
		ctx.Logger.Trace("Object", "type", "struct")
		ctx.Logger.NextCallLevel()
		golangType, err = o.buildLangStruct(ctx, flags)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
		return
	}

	switch typeName {
	case "array":
		ctx.Logger.Trace("Object", "type", "array")
		ctx.Logger.NextCallLevel()
		golangType, err = o.buildLangArray(ctx, flags)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
	case "null", "":
		ctx.Logger.Trace("Object", "type", "any")
		golangType = &lang.GoSimple{TypeName: "any", IsInterface: true, OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	case "boolean":
		ctx.Logger.Trace("Object", "type", "bool")
		aliasedType = &lang.GoSimple{TypeName: "bool", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	case "integer":
		ctx.Logger.Trace("Object", "type", "int")
		aliasedType = &lang.GoSimple{TypeName: "int", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	case "number":
		ctx.Logger.Trace("Object", "type", "float64")
		aliasedType = &lang.GoSimple{TypeName: "float64", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	case "string":
		ctx.Logger.Trace("Object", "type", "string")
		aliasedType = &lang.GoSimple{TypeName: "string", OriginalType: typeName, OriginalFormat: o.Format, StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx)}
	default:
		return nil, types.CompileError{Err: fmt.Errorf("unknown jsonschema type %q", typeName), Path: ctx.CurrentRefPointer()}
	}

	if aliasedType != nil {
		_, isComponent := flags[common.SchemaTagComponent]
		_, hasDefinition := flags[common.SchemaTagDefinition]
		golangType = &lang.GoTypeDefinition{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(o.Title, ""),
				Description:   o.Description,
				HasDefinition: hasDefinition,
				ArtifactKind:  lo.Ternary(isComponent, common.ArtifactKindSchema, common.ArtifactKindOther),
			},
			RedefinedType: aliasedType,
		}
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

func (o Object) buildLangStruct(ctx *compile.Context, flags map[common.SchemaTag]string) (*lang.GoStruct, error) {
	_, hasDefinition := flags[common.SchemaTagDefinition]
	_, isComponent := flags[common.SchemaTagComponent]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  ctx.GenerateObjName(objName, ""),
			Description:   o.Description,
			HasDefinition: hasDefinition,
			ArtifactKind:  lo.Ternary(isComponent, common.ArtifactKindSchema, common.ArtifactKindOther),
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
		contentTypesFunc = func() []string {
			tagNames := lo.Uniq(lo.Map(messagesPrm.T(), func(item *render.Message, _ int) string {
				return guessTagByContentType(item.EffectiveContentType())
			}))
			slices.Sort(tagNames)
			return tagNames
		}
	}

	// regular properties
	for _, entry := range o.Properties.Entries() {
		ctx.Logger.Trace("Object property", "name", entry.Key)
		ref := ctx.CurrentRefPointer("properties", entry.Key)
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)

		var langObj common.GolangType = prm
		if lo.Contains(o.Required, entry.Key) {
			langObj = &lang.GoPointer{Type: langObj}
		}

		propName, _ := lo.Coalesce(entry.Value.XGoName, entry.Key)
		f := lang.GoStructField{
			OriginalName:     utils.ToGolangName(propName, true),
			MarshalName:      entry.Key,
			Description:      entry.Value.Description,
			Type:             langObj,
			ContentTypesFunc: contentTypesFunc,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	// TODO: unmarshal extra fields somehow somewhere to AdditionalProperties field
	if o.AdditionalProperties != nil {
		propName, _ := lo.Coalesce(o.AdditionalProperties.V0.XGoName, o.Title)
		switch o.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			// TODO: handle $ref in AdditionalProperties items
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
					RedefinedType: &lang.GoSimple{TypeName: "any", IsInterface: true},
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
	_, hasDefinition := flags[common.SchemaTagDefinition]
	_, isComponent := flags[common.SchemaTagComponent]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.GoArray{
		BaseType: lang.BaseType{
			OriginalName:  ctx.GenerateObjName(objName, ""),
			Description:   o.Description,
			HasDefinition: hasDefinition,
			ArtifactKind:  lo.Ternary(isComponent, common.ArtifactKindSchema, common.ArtifactKindOther),
		},
		ItemsType:             nil,
		StructFieldRenderInfo: o.getStructFieldRenderInfo(ctx),
	}

	switch {
	case o.Items != nil && o.Items.Selector == 0: // Only one "type:" of items
		ctx.Logger.Trace("Object items", "typesCount", "single")
		ref := ctx.CurrentRefPointer("items")
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)
		res.ItemsType = prm
	case o.Items == nil || o.Items.Selector == 1: // No items or Several types for each item sequentially
		ctx.Logger.Trace("Object items", "typesCount", "zero or several")
		valTyp := lang.GoTypeDefinition{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(objName, "ItemsItemValue"),
				Description:   "",
				HasDefinition: false,
			},
			RedefinedType: &lang.GoSimple{TypeName: "any", IsInterface: true},
		}
		res.ItemsType = &lang.GoMap{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(objName, "ItemsItem"),
				Description:   "",
				HasDefinition: false,
			},
			KeyType:   &lang.GoSimple{TypeName: "string"},
			ValueType: &valTyp,
		}
	}

	return &res, nil
}

func (o Object) buildUnionStruct(ctx *compile.Context, flags map[common.SchemaTag]string) (*lang.UnionStruct, error) {
	_, hasDefinition := flags[common.SchemaTagDefinition]
	_, isComponent := flags[common.SchemaTagComponent]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.UnionStruct{
		GoStruct: lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(objName, ""),
				Description:   o.Description,
				HasDefinition: hasDefinition,
				ArtifactKind:  lo.Ternary(isComponent, common.ArtifactKindSchema, common.ArtifactKindOther),
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
			ctx.Logger.Trace("Extra tags", "tags", lo.FromEntries(res.Tags.Entries()))
		}
	}
	if res.TagValues = o.XGoTagsValues; len(res.TagValues) > 0 {
		ctx.Logger.Trace("Extra tags values", "values", res.TagValues)
	}

	return res
}
