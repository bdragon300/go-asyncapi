package asyncapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"slices"
	"strconv"

	"github.com/bdragon300/go-asyncapi/internal/types"

	yaml "gopkg.in/yaml.v3"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"

	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

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

func (o Object) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj, err := o.build(ctx, ctx.Stack.Top().Flags, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (o Object) build(ctx *common.CompileContext, flags map[common.SchemaTag]string, objectKey string) (common.GolangType, error) {
	_, isComponent := flags[common.SchemaTagComponent]
	ignore := o.XIgnore || isComponent //&& !ctx.CompileOpts.ModelOpts.IsAllowedName(objectKey))
	if ignore {
		ctx.Logger.Debug("Object denoted to be ignored")
		return &lang.GoSimple{Name: "any", IsInterface: true}, nil
	}
	if o.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", o.Ref)
		res := lang.NewGolangTypePromise(o.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	if o.Type == nil {
		o.Type = o.getDefaultObjectType(ctx)
	}

	if len(o.OneOf)+len(o.AnyOf)+len(o.AllOf) > 0 {
		ctx.Logger.Trace("Object is union struct")
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

func (o Object) getTypeName(ctx *common.CompileContext) (typeName string, nullable bool, err error) {
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
			err = types.CompileError{Err: errors.New("empty object type"), Path: ctx.PathStackRef()}
			return
		}
		ctx.Logger.Trace(fmt.Sprintf("Multitype object type inferred as %q", typeName))
	}
	return
}

func (o Object) buildGolangType(ctx *common.CompileContext, flags map[common.SchemaTag]string, typeName string) (golangType common.GolangType, err error) {
	var aliasedType *lang.GoSimple

	if typeName == "object" {
		if o.XGoType != nil && !o.XGoType.V1.Embedded {
			f := buildXGoType(o.XGoType)
			ctx.Logger.Trace("Object is custom type", "type", f.String())
			return f, nil
		}

		ctx.Logger.Trace("Object is struct")
		ctx.Logger.NextCallLevel()
		golangType, err = o.buildLangStruct(ctx, flags)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
		return
	}

	if o.XGoType != nil {
		f := buildXGoType(o.XGoType)
		ctx.Logger.Trace("Object is a custom type", "type", f.String())
		return f, nil
	}

	switch typeName {
	case "array":
		ctx.Logger.Trace("Object is array")
		ctx.Logger.NextCallLevel()
		golangType, err = o.buildLangArray(ctx, flags)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
	case "null", "":
		ctx.Logger.Trace("Object is any")
		golangType = &lang.GoSimple{Name: "any", IsInterface: true}
	case "boolean":
		ctx.Logger.Trace("Object is bool")
		aliasedType = &lang.GoSimple{Name: "bool"}
	case "integer":
		// TODO: "format:"
		ctx.Logger.Trace("Object is int")
		aliasedType = &lang.GoSimple{Name: "int"}
	case "number":
		// TODO: "format:"
		ctx.Logger.Trace("Object is float64")
		aliasedType = &lang.GoSimple{Name: "float64"}
	case "string":
		ctx.Logger.Trace("Object is string")
		aliasedType = &lang.GoSimple{Name: "string"}
	default:
		return nil, types.CompileError{Err: fmt.Errorf("unknown jsonschema type %q", typeName), Path: ctx.PathStackRef()}
	}

	if aliasedType != nil {
		_, hasDefinition := flags[common.SchemaTagDefinition]
		golangType = &lang.GoTypeAlias{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(o.Title, ""),
				Description:   o.Description,
				HasDefinition: hasDefinition,
			},
			AliasedType: aliasedType,
		}
	}

	return golangType, nil
}

// getDefaultObjectType is backwards compatible, guessing the user intention when they didn't specify a type.
func (o Object) getDefaultObjectType(ctx *common.CompileContext) *types.Union2[string, []string] {
	switch {
	case o.Ref == "" && o.Properties.Len() > 0:
		ctx.Logger.Trace("Object type is empty, determined `object` because of `properties` presence")
		return types.ToUnion2[string, []string]("object")
	case o.Items != nil: // TODO: fix type when AllOf, AnyOf, OneOf
		ctx.Logger.Trace("Object type is empty, determined `array` because of `items` presence")
		return types.ToUnion2[string, []string]("array")
	default:
		ctx.Logger.Trace("Object type is empty, guessing it `object` by default")
		return types.ToUnion2[string, []string]("object")
	}
}

func (o Object) buildLangStruct(ctx *common.CompileContext, flags map[common.SchemaTag]string) (*lang.GoStruct, error) {
	_, hasDefinition := flags[common.SchemaTagDefinition]
	_, isComponent := flags[common.SchemaTagComponent]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.GoStruct{
		BaseType: lang.BaseType{
			Name:          ctx.GenerateObjName(objName, ""),
			Description:   o.Description,
			HasDefinition: hasDefinition,
		},
		ObjectKind: lo.Ternary(isComponent, common.ObjectKindSchema, common.ObjectKindOther),
	}
	// TODO: cache the object name in case any sub-schemas recursively reference it

	var contentTypesFunc func() []string
	_, isMarshal := flags[common.SchemaTagMarshal]
	if isMarshal {
		messagesPrm := lang.NewListCbPromise[*render.Message](func(item common.Renderable, _ []string) bool {
			_, ok := item.(*render.Message)
			return ok
		})
		ctx.PutListPromise(messagesPrm)
		contentTypesFunc = func() []string {
			tagNames := lo.Uniq(lo.Map(messagesPrm.T(), func(item *render.Message, _ int) string {
				return item.EffectiveContentType()
			}))
			slices.Sort(tagNames)
			return tagNames
		}
	}

	// Embed external type into the current one, if x-go-type->embedded == true
	if o.XGoType != nil && o.XGoType.V1.Embedded {
		f := buildXGoType(o.XGoType)
		ctx.Logger.Trace("Object struct embedded custom type", "type", f.String())
		res.Fields = append(res.Fields, lang.GoStructField{Type: f})
	}

	// regular properties
	for _, entry := range o.Properties.Entries() {
		ctx.Logger.Trace("Object property", "name", entry.Key)
		ref := ctx.PathStackRef("properties", entry.Key)
		prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)

		var langObj common.GolangType = prm
		if lo.Contains(o.Required, entry.Key) {
			langObj = &lang.GoPointer{Type: langObj}
		}

		propName, _ := lo.Coalesce(entry.Value.XGoName, entry.Key)
		xTags, xTagNames, xTagVals := entry.Value.xGoTagsInfo(ctx)
		f := lang.GoStructField{
			Name:             utils.ToGolangName(propName, true),
			MarshalName:      entry.Key,
			Description:      entry.Value.Description,
			Type:             langObj,
			ContentTypesFunc: contentTypesFunc,
			ExtraTags:        xTags,
			ExtraTagNames:    xTagNames,
			ExtraTagValues:   xTagVals,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	// TODO: unmarshal extra fields somehow somewhere to AdditionalProperties field
	if o.AdditionalProperties != nil {
		propName, _ := lo.Coalesce(o.AdditionalProperties.V0.XGoName, o.Title)
		switch o.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			ctx.Logger.Trace("Object additional properties as an object")
			ref := ctx.PathStackRef("additionalProperties")
			prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
			ctx.PutPromise(prm)
			xTags, xTagNames, xTagVals := o.AdditionalProperties.V0.xGoTagsInfo(ctx)
			f := lang.GoStructField{
				Name:        "AdditionalProperties",
				Description: o.AdditionalProperties.V0.Description,
				Type: &lang.GoMap{
					BaseType: lang.BaseType{
						Name:          ctx.GenerateObjName(propName, "AdditionalProperties"),
						Description:   o.AdditionalProperties.V0.Description,
						HasDefinition: false,
					},
					KeyType:   &lang.GoSimple{Name: "string"},
					ValueType: prm,
				},
				ExtraTags:      xTags,
				ExtraTagNames:  xTagNames,
				ExtraTagValues: xTagVals,
			}
			res.Fields = append(res.Fields, f)
		case 1:
			ctx.Logger.Trace("Object additional properties as boolean flag")
			if o.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := lang.GoTypeAlias{
					BaseType: lang.BaseType{
						Name:          ctx.GenerateObjName(propName, "AdditionalPropertiesValue"),
						Description:   "",
						HasDefinition: false,
					},
					AliasedType: &lang.GoSimple{Name: "any", IsInterface: true},
				}
				f := lang.GoStructField{
					Name: "AdditionalProperties",
					Type: &lang.GoMap{
						BaseType: lang.BaseType{
							Name:          ctx.GenerateObjName(propName, "AdditionalProperties"),
							Description:   "",
							HasDefinition: false,
						},
						KeyType:   &lang.GoSimple{Name: "string"},
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

func (o Object) buildLangArray(ctx *common.CompileContext, flags map[common.SchemaTag]string) (*lang.GoArray, error) {
	_, hasDefinition := flags[common.SchemaTagDefinition]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.GoArray{
		BaseType: lang.BaseType{
			Name:          ctx.GenerateObjName(objName, ""),
			Description:   o.Description,
			HasDefinition: hasDefinition,
		},
		ItemsType: nil,
	}

	switch {
	case o.Items != nil && o.Items.Selector == 0: // Only one "type:" of items
		ctx.Logger.Trace("Object items (single type)")
		ref := ctx.PathStackRef("items")
		prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		res.ItemsType = prm
	case o.Items == nil || o.Items.Selector == 1: // No items or Several types for each item sequentially
		ctx.Logger.Trace("Object items (zero or several types)")
		valTyp := lang.GoTypeAlias{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(objName, "ItemsItemValue"),
				Description:   "",
				HasDefinition: false,
			},
			AliasedType: &lang.GoSimple{Name: "any", IsInterface: true},
		}
		res.ItemsType = &lang.GoMap{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(objName, "ItemsItem"),
				Description:   "",
				HasDefinition: false,
			},
			KeyType:   &lang.GoSimple{Name: "string"},
			ValueType: &valTyp,
		}
	}

	return &res, nil
}

func (o Object) buildUnionStruct(ctx *common.CompileContext, flags map[common.SchemaTag]string) (*lang.UnionStruct, error) {
	_, hasDefinition := flags[common.SchemaTagDefinition]
	_, isComponent := flags[common.SchemaTagComponent]
	objName, _ := lo.Coalesce(o.XGoName, o.Title)
	res := lang.UnionStruct{
		GoStruct: lang.GoStruct{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(objName, ""),
				Description:   o.Description,
				HasDefinition: hasDefinition,
			},
			ObjectKind: lo.Ternary(isComponent, common.ObjectKindSchema, common.ObjectKindOther),
		},
	}

	// Collect all messages to retrieve struct field tags
	messagesPrm := lang.NewListCbPromise[*render.Message](func(item common.Renderable, _ []string) bool {
		_, ok := item.(*render.Message)
		return ok
	})
	ctx.PutListPromise(messagesPrm)

	res.Fields = lo.Times(len(o.OneOf), func(index int) lang.GoStructField {
		ref := ctx.PathStackRef("oneOf", strconv.Itoa(index))
		prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return lang.GoStructField{Type: &lang.GoPointer{Type: prm}}
	})
	res.Fields = append(res.Fields, lo.Times(len(o.AnyOf), func(index int) lang.GoStructField {
		ref := ctx.PathStackRef("anyOf", strconv.Itoa(index))
		prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return lang.GoStructField{Type: &lang.GoPointer{Type: prm}}
	})...)
	res.Fields = append(res.Fields, lo.Times(len(o.AllOf), func(index int) lang.GoStructField {
		ref := ctx.PathStackRef("allOf", strconv.Itoa(index))
		prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return lang.GoStructField{Type: prm}
	})...)

	return &res, nil
}

func (o Object) xGoTagsInfo(ctx *common.CompileContext) (xTags types.OrderedMap[string, string], xTagNames []string, xTagValues []string) {
	if o.XGoTags != nil {
		switch o.XGoTags.Selector {
		case 0:
			xTagNames = o.XGoTags.V0
			ctx.Logger.Trace("Extra tags", "names", xTagNames)
		case 1:
			xTags = o.XGoTags.V1
			ctx.Logger.Trace("Extra tags", "tags", lo.FromEntries(xTags.Entries()))
		}
	}
	if xTagValues = o.XGoTagsValues; len(xTagValues) > 0 {
		ctx.Logger.Trace("Extra tags values", "values", xTagValues)
	}
	return
}
