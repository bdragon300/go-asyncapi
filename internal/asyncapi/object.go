package asyncapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	yaml "gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"

	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/samber/lo"
)

type Object struct {
	Type                 *types.Union2[string, []string]            `json:"type" yaml:"type"`
	AdditionalItems      *types.Union2[Object, bool]                `json:"additionalItems" yaml:"additionalItems"`
	AdditionalProperties *types.Union2[Object, bool]                `json:"additionalProperties" yaml:"additionalProperties"`
	AllOf                []Object                                   `json:"allOf" yaml:"allOf" cgen:"directRender"`
	AnyOf                []Object                                   `json:"anyOf" yaml:"anyOf" cgen:"directRender"`
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
	OneOf                []Object                                   `json:"oneOf" yaml:"oneOf" cgen:"directRender"`
	Pattern              string                                     `json:"pattern" yaml:"pattern"`
	PatternProperties    types.OrderedMap[string, Object]           `json:"patternProperties" yaml:"patternProperties"` // Mapping regex->schema
	Properties           types.OrderedMap[string, Object]           `json:"properties" yaml:"properties"`
	PropertyNames        *Object                                    `json:"propertyNames" yaml:"propertyNames"`
	ReadOnly             *bool                                      `json:"readOnly" yaml:"readOnly"`
	Required             []string                                   `json:"required" yaml:"required"`
	Then                 *Object                                    `json:"then" yaml:"then"`
	Title                string                                     `json:"title" yaml:"title"`
	UniqueItems          *bool                                      `json:"uniqueItems" yaml:"uniqueItems"`

	XNullable         *bool                                     `json:"x-nullable" yaml:"x-nullable"`
	XGoType           *types.Union3[string, []xGoType, xGoType] `json:"x-go-type" yaml:"x-go-type"`
	XGoName           string                                    `json:"x-go-name" yaml:"x-go-name"`
	XGoExtraTags      types.OrderedMap[string, string]          `json:"x-go-extra-tags" yaml:"x-go-extra-tags"`
	XGoTagExtraValues []string                                  `json:"x-go-tag-extra-values" yaml:"x-go-tag-extra-values"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (m Object) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := buildGolangType(ctx, m, ctx.Stack.Top().Flags)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func buildGolangType(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (common.GolangType, error) {
	var err error
	if schema.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", schema.Ref)
		res := render.NewGolangTypePromise(schema.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	if len(schema.OneOf)+len(schema.AnyOf)+len(schema.AllOf) > 0 {
		ctx.Logger.Trace("Object is union struct")
		return buildUnionStruct(ctx, schema) // TODO: process other items that can be set along with oneof/anyof/allof
	}

	fixMissingTypeValue(ctx, &schema)

	schemaType := schema.Type
	typeName := schemaType.V0
	nullable := false
	if schemaType.Selector == 1 { // Multiple types, e.g. { "type": [ "object", "array", "null" ] }
		typeName, nullable, err = simplifyMultiType(schemaType.V1)
		if err != nil {
			return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
		}
		ctx.Logger.Trace(fmt.Sprintf("Multitype object type inferred as %q", typeName))
	}

	nullable = nullable || lo.FromPtr(schema.XNullable)

	_, directRender := flags[common.SchemaTagDirectRender]
	// TODO: "type": { "enum": [ "residential", "business" ] }
	// One type: { "type": "object" }
	var builtType common.GolangType
	var aliasedType *render.Simple

	switch typeName {
	case "object":
		ctx.Logger.Trace("Object is struct")
		ctx.Logger.NextCallLevel()
		if builtType, err = buildLangStruct(ctx, schema, flags); err != nil {
			return nil, err
		}
		ctx.Logger.PrevCallLevel()
	case "array":
		ctx.Logger.Trace("Object is array")
		ctx.Logger.NextCallLevel()
		if builtType, err = buildLangArray(ctx, schema, flags); err != nil {
			return nil, err
		}
		ctx.Logger.PrevCallLevel()
	case "null", "":
		ctx.Logger.Trace("Object is any")
		builtType = &render.Simple{Name: "any", IsIface: true}
	case "boolean":
		ctx.Logger.Trace("Object is bool")
		aliasedType = &render.Simple{Name: "bool"}
	case "integer":
		// TODO: "format:"
		ctx.Logger.Trace("Object is int")
		aliasedType = &render.Simple{Name: "int"}
	case "number":
		// TODO: "format:"
		ctx.Logger.Trace("Object is float64")
		aliasedType = &render.Simple{Name: "float64"}
	case "string":
		ctx.Logger.Trace("Object is string")
		aliasedType = &render.Simple{Name: "string"}
	default:
		return nil, types.CompileError{Err: fmt.Errorf("unknown jsonschema type %q", typeName), Path: ctx.PathRef()}
	}

	if aliasedType != nil {
		builtType = &render.TypeAlias{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(schema.Title, ""),
				Description:  schema.Description,
				DirectRender: directRender,
				PackageName:  ctx.TopPackageName(),
			},
			AliasedType: aliasedType,
		}
	}

	if nullable {
		ctx.Logger.Trace("Object is nullable, make it pointer")
		builtType = &render.Pointer{Type: builtType, DirectRender: directRender}
	}
	return builtType, nil
}

// fixMissingTypeValue is backwards compatible, guessing the users intention when they didn't specify a type.
func fixMissingTypeValue(ctx *common.CompileContext, s *Object) {
	if s.Type == nil {
		if s.Ref == "" && s.Properties.Len() > 0 {
			ctx.Logger.Trace("Object type is empty, determined `object` because of `properties` presence")
			s.Type = types.ToUnion2[string, []string]("object")
			return
		}
		// TODO: fix type when AllOf, AnyOf, OneOf
		if s.Items != nil {
			ctx.Logger.Trace("Object type is empty, determined `array` because of `items` presence")
			s.Type = types.ToUnion2[string, []string]("array")
			return
		}

		ctx.Logger.Trace("Object type is empty, guessing it `object` by default")
		s.Type = types.ToUnion2[string, []string]("object")
	}
}

func simplifyMultiType(schemaType []string) (string, bool, error) {
	nullable := lo.Contains(schemaType, "null")
	typs := lo.Reject(schemaType, func(item string, _ int) bool { return item == "null" }) // Throw out null (if any)
	switch {
	case len(typs) > 1: // More than one type along with null -> 'any'
		return "", nullable, nil
	case len(typs) == 1: // One type along with null -> pointer to this type
		return typs[0], nullable, nil
	case nullable: // Null only -> 'any', that can be only nil
		return "null", nullable, nil
	default:
		return "", nullable, errors.New("empty object type")
	}
}

func buildLangStruct(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (*render.Struct, error) {
	_, noInline := flags[common.SchemaTagDirectRender]
	res := render.Struct{
		BaseType: render.BaseType{
			Name:         ctx.GenerateObjName(schema.Title, ""),
			Description:  schema.Description,
			DirectRender: noInline,
			PackageName:  ctx.TopPackageName(),
		},
	}
	// TODO: cache the object name in case any sub-schemas recursively reference it

	var msgLinks *render.LinkList[*render.Message]
	// Collect all messages to retrieve struct field tags
	if ctx.TopPackageName() == "models" { // TODO: fix hardcode
		msgLinks = render.NewListCbPromise[*render.Message](func(item common.Renderer, _ []string) bool {
			_, ok := item.(*render.Message)
			return ok
		})
		ctx.PutListPromise(msgLinks)
	}

	// regular properties
	for _, entry := range schema.Properties.Entries() {
		ctx.Logger.Trace("Object property", "name", entry.Key)
		ref := path.Join(ctx.PathRef(), "properties", entry.Key)
		prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)

		var langObj common.GolangType = prm
		if lo.Contains(schema.Required, entry.Key) {
			langObj = &render.Pointer{Type: langObj}
		}

		f := render.StructField{
			Name:        utils.ToGolangName(entry.Key, true),
			MarshalName: entry.Key,
			Type:        langObj,
			Description: entry.Value.Description,
			TagsSource:  msgLinks,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	if schema.AdditionalProperties != nil {
		switch schema.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			ctx.Logger.Trace("Object additional properties as an object")
			ref := path.Join(ctx.PathRef(), "additionalProperties")
			langObj := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
			f := render.StructField{
				Name: "AdditionalProperties",
				Type: &render.Map{
					BaseType: render.BaseType{
						Name:         ctx.GenerateObjName(schema.Title, "AdditionalProperties"),
						Description:  schema.AdditionalProperties.V0.Description,
						DirectRender: false,
						PackageName:  ctx.TopPackageName(),
					},
					KeyType:   &render.Simple{Name: "string"},
					ValueType: langObj,
				},
				Description: schema.AdditionalProperties.V0.Description,
			}
			res.Fields = append(res.Fields, f)
		case 1:
			ctx.Logger.Trace("Object additional properties as boolean flag")
			if schema.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := render.TypeAlias{
					BaseType: render.BaseType{
						Name:         ctx.GenerateObjName(schema.Title, "AdditionalPropertiesValue"),
						Description:  "",
						DirectRender: false,
						PackageName:  ctx.TopPackageName(),
					},
					AliasedType: &render.Simple{Name: "any", IsIface: true},
				}
				f := render.StructField{
					Name: "AdditionalProperties",
					Type: &render.Map{
						BaseType: render.BaseType{
							Name:         ctx.GenerateObjName(schema.Title, "AdditionalProperties"),
							Description:  "",
							DirectRender: false,
							PackageName:  ctx.TopPackageName(),
						},
						KeyType:   &render.Simple{Name: "string"},
						ValueType: &valTyp,
					},
					TagsSource: msgLinks,
				}
				res.Fields = append(res.Fields, f)
			}
		}
	}

	return &res, nil
}

func buildLangArray(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (*render.Array, error) {
	_, noInline := flags[common.SchemaTagDirectRender]
	res := render.Array{
		BaseType: render.BaseType{
			Name:         ctx.GenerateObjName(schema.Title, ""),
			Description:  schema.Description,
			DirectRender: noInline,
			PackageName:  ctx.TopPackageName(),
		},
		ItemsType: nil,
	}

	switch {
	case schema.Items != nil && schema.Items.Selector == 0: // Only one "type:" of items
		ctx.Logger.Trace("Object items (single type)")
		ref := path.Join(ctx.PathRef(), "items")
		res.ItemsType = render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
	case schema.Items == nil || schema.Items.Selector == 1: // No items or Several types for each item sequentially
		ctx.Logger.Trace("Object items (zero or several types)")
		valTyp := render.TypeAlias{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(schema.Title, "ItemsItemValue"),
				Description:  "",
				DirectRender: false,
				PackageName:  ctx.TopPackageName(),
			},
			AliasedType: &render.Simple{Name: "any", IsIface: true},
		}
		res.ItemsType = &render.Map{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(schema.Title, "ItemsItem"),
				Description:  "",
				DirectRender: false,
				PackageName:  ctx.TopPackageName(),
			},
			KeyType:   &render.Simple{Name: "string"},
			ValueType: &valTyp,
		}
	}

	return &res, nil
}

func buildUnionStruct(ctx *common.CompileContext, schema Object) (*render.UnionStruct, error) {
	res := render.UnionStruct{
		Struct: render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(schema.Title, ""),
				Description:  schema.Description,
				DirectRender: true, // Always render unions as separate types
				PackageName:  ctx.TopPackageName(),
			},
		},
	}

	// Collect all messages to retrieve struct field tags
	msgLinks := render.NewListCbPromise[*render.Message](func(item common.Renderer, _ []string) bool {
		_, ok := item.(*render.Message)
		return ok
	})
	ctx.PutListPromise(msgLinks)

	res.Fields = lo.Times(len(schema.OneOf), func(index int) render.StructField {
		ref := path.Join(ctx.PathRef(), "oneOf", strconv.Itoa(index))
		langTyp := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(langTyp)
		return render.StructField{Type: &render.Pointer{Type: langTyp}}
	})
	res.Fields = append(res.Fields, lo.Times(len(schema.AnyOf), func(index int) render.StructField {
		ref := path.Join(ctx.PathRef(), "anyOf", strconv.Itoa(index))
		langTyp := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(langTyp)
		return render.StructField{Type: &render.Pointer{Type: langTyp}}
	})...)
	res.Fields = append(res.Fields, lo.Times(len(schema.AllOf), func(index int) render.StructField {
		ref := path.Join(ctx.PathRef(), "allOf", strconv.Itoa(index))
		langTyp := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(langTyp)
		return render.StructField{Type: langTyp}
	})...)

	return &res, nil
}
