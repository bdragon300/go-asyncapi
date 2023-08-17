package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/lang/types"
	"github.com/bdragon300/asyncapi-codegen/internal/scancontext"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

const pathSep = "_"

type Object struct {
	Ref  string                          `json:"$ref" yaml:"$ref"`
	Type *utils.Union2[string, []string] `json:"type" yaml:"type"`

	AdditionalItems      *utils.Union2[Object, bool]                `json:"additionalItems" yaml:"additionalItems"`
	AdditionalProperties *utils.Union2[Object, bool]                `json:"additionalProperties" yaml:"additionalProperties"`
	AllOf                []Object                                   `json:"allOf" yaml:"allOf"`
	AnyOf                []Object                                   `json:"anyOf" yaml:"anyOf"`
	Const                *utils.Union2[json.RawMessage, yaml.Node]  `json:"const" yaml:"const"`
	Contains             *Object                                    `json:"contains" yaml:"contains"`
	Default              *utils.Union2[json.RawMessage, yaml.Node]  `json:"default" yaml:"default"`
	Definitions          map[string]Object                          `json:"definitions" yaml:"definitions"`
	Deprecated           *bool                                      `json:"deprecated" yaml:"deprecated"`
	Description          string                                     `json:"description" yaml:"description"`
	Discriminator        string                                     `json:"discriminator" yaml:"discriminator"`
	Else                 *Object                                    `json:"else" yaml:"else"`
	Enum                 []utils.Union2[json.RawMessage, yaml.Node] `json:"enum" yaml:"enum"`
	Examples             []utils.Union2[json.RawMessage, yaml.Node] `json:"examples" yaml:"examples"`
	ExclusiveMaximum     *utils.Union2[bool, json.Number]           `json:"exclusiveMaximum" yaml:"exclusiveMaximum"`
	ExclusiveMinimum     *utils.Union2[bool, json.Number]           `json:"exclusiveMinimum" yaml:"exclusiveMinimum"`
	ExternalDocs         *ExternalDocsItem                          `json:"externalDocs" yaml:"externalDocs"`
	Format               string                                     `json:"format" yaml:"format"`
	If                   *Object                                    `json:"if" yaml:"if"`
	Items                *utils.Union2[Object, []Object]            `json:"items" yaml:"items"`
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
	OneOf                []Object                                   `json:"oneOf" yaml:"oneOf"`
	Pattern              string                                     `json:"pattern" yaml:"pattern"`
	PatternProperties    map[string]Object                          `json:"patternProperties" yaml:"patternProperties"`
	Properties           map[string]Object                          `json:"properties" yaml:"properties"`
	PropertyNames        *Object                                    `json:"propertyNames" yaml:"propertyNames"`
	ReadOnly             *bool                                      `json:"readOnly" yaml:"readOnly"`
	Required             []string                                   `json:"required" yaml:"required"`
	Then                 *Object                                    `json:"then" yaml:"then"`
	Title                string                                     `json:"title" yaml:"title"`
	UniqueItems          *bool                                      `json:"uniqueItems" yaml:"uniqueItems"`
}

func (o Object) Build(ctx *scancontext.Context) error {
	langObj, err := buildLangType(ctx, o, ctx.Top().Flags)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.Buckets[scancontext.BucketLangType].Put(ctx, langObj.(scancontext.LangRenderer))
	return nil
}

func buildLangType(ctx *scancontext.Context, schema Object, flags map[string]string) (types.LangType, error) {
	if schema.Ref != "" {
		res := &types.TypeBindWrapper{
			BaseType: types.BaseType{Inline: true},
			RefQuery: scancontext.NewRefQuery[types.LangType](ctx, schema.Ref),
		}
		ctx.RefMgr.Add(res.RefQuery, scancontext.BucketLangType)
		return res, nil
	}

	fixMissingTypeValue(&schema)

	schemaType := schema.Type
	typ := schemaType.V0
	nullable := false
	if schemaType.Selector == 1 { // Multiple types, e.g. { "type": [ "object", "array", "null" ] }
		typ, nullable = inspectMultiType(schemaType.V1)
	}

	_, noInline := flags["noinline"]
	// TODO: "type": { "enum": [ "residential", "business" ] }
	// One type: { "type": "object" }
	langTyp := ""
	switch typ {
	case "object":
		res, err := buildLangStruct(ctx, schema, flags)
		if err != nil {
			return nil, err
		}
		res.Nullable = nullable
		return res, nil
	case "array":
		res, err := buildLangArray(ctx, schema, flags)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "null", "":
		return &types.Any{
			BaseType: types.BaseType{
				DefaultName: getTypeName(ctx, schema.Title, ""),
				Description: schema.Description,
				Inline:      !noInline,
			},
		}, nil
	case "boolean":
		langTyp = "bool"
	case "integer":
		// TODO: "format:"
		langTyp = "int"
	case "number":
		// TODO: "format:"
		langTyp = "float64"
	case "string":
		langTyp = "string"
	default:
		return nil, fmt.Errorf("unknown jsonschema type %q", typ)
	}

	return &types.PrimitiveType{
		BaseType: types.BaseType{
			DefaultName: getTypeName(ctx, schema.Title, ""),
			Description: schema.Description,
			Inline:      !noInline,
		},
		LangType: langTyp,
		Nullable: nullable,
	}, nil
}

// fixMissingTypeValue is backwards compatible, guessing the users intention when they didn't specify a type.
func fixMissingTypeValue(s *Object) {
	if s.Type == nil {
		if s.Ref == "" && len(s.Properties) > 0 {
			s.Type = utils.ToUnion2[string, []string]("object")
			return
		}
		// TODO: fix type when AllOf, AnyOf, OneOf
		if s.Items != nil {
			s.Type = utils.ToUnion2[string, []string]("array")
			return
		}
		panic("Unable to determine object type")
	}
}

func inspectMultiType(schemaType []string) (typ string, nullable bool) {
	nullable = lo.Contains(schemaType, "null")
	typs := lo.Reject(schemaType, func(item string, _ int) bool { return item == "null" }) // Throw out null (if any)
	switch {
	case len(typs) > 2: // More than one type along with null -> 'any'
		return "", nullable
	case len(typs) == 0: // Null only -> 'any', that can be only nil
		return "null", nullable
	default: // One type along with null -> pointer to this type
		return typs[0], nullable
	}
}

func buildLangStruct(ctx *scancontext.Context, schema Object, flags map[string]string) (*types.Struct, error) {
	_, noInline := flags["noinline"]
	res := types.Struct{
		BaseType: types.BaseType{
			DefaultName: getTypeName(ctx, schema.Title, ""),
			Description: schema.Description,
			Inline:      !noInline,
		},
		Fields: make(map[string]types.Field),
	}
	// TODO: cache the object name in case any sub-schemas recursively reference it

	// regular properties
	for propKey, prop := range schema.Properties {
		langObj, err := buildLangType(ctx, prop, map[string]string{})
		if err != nil {
			return nil, err
		}
		f := types.Field{
			Name:          getFieldName(propKey),
			Type:          langObj,
			RequiredValue: lo.Contains(schema.Required, propKey),
			Tags:          nil, // TODO
			Description:   prop.Description,
		}
		res.Fields[f.Name] = f
	}

	// additionalProperties with typed sub-schema
	if schema.AdditionalProperties != nil {
		switch schema.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			langObj, err := buildLangType(ctx, schema.AdditionalProperties.V0, map[string]string{})
			if err != nil {
				return nil, err
			}
			f := types.Field{
				Name: "AdditionalProperties",
				Type: &types.Map{
					BaseType: types.BaseType{
						DefaultName: getTypeName(ctx, schema.Title, "AdditionalProperties"),
						Description: schema.AdditionalProperties.V0.Description,
						Inline:      true,
					},
					KeyType:   "string",
					ValueType: langObj,
				},
				RequiredValue: false,
				Tags:          nil, // TODO
				Description:   schema.AdditionalProperties.V0.Description,
			}
			res.Fields[f.Name] = f
		case 1:
			if schema.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := types.Any{
					BaseType: types.BaseType{
						DefaultName: getTypeName(ctx, schema.Title, "AdditionalPropertiesValue"),
						Description: "",
						Inline:      true,
					},
				}
				f := types.Field{
					Name: "AdditionalProperties",
					Type: &types.Map{
						BaseType: types.BaseType{
							DefaultName: getTypeName(ctx, schema.Title, "AdditionalProperties"),
							Description: "",
							Inline:      true,
						},
						KeyType:   "string",
						ValueType: &valTyp,
					},
					RequiredValue: false,
					Tags:          nil, // TODO
				}
				res.Fields[f.Name] = f
			}
		}
	}

	return &res, nil
}

func buildLangArray(ctx *scancontext.Context, schema Object, flags map[string]string) (*types.Array, error) {
	_, noInline := flags["noinline"]
	res := types.Array{
		BaseType: types.BaseType{
			DefaultName: getTypeName(ctx, schema.Title, ""),
			Description: schema.Description,
			Inline:      !noInline,
		},
		ItemsType: nil,
	}

	switch {
	case schema.Items != nil && schema.Items.Selector == 0: // Only one "type:" of items
		langObj, err := buildLangType(ctx, schema.Items.V0, flags)
		if err != nil {
			return nil, err
		}
		res.ItemsType = langObj
	case schema.Items == nil || schema.Items.Selector == 1: // No items or Several types for each item sequentially
		valTyp := types.Any{
			BaseType: types.BaseType{
				DefaultName: getTypeName(ctx, schema.Title, "ItemsItemValue"),
				Description: "",
				Inline:      true,
			},
		}
		res.ItemsType = &types.Map{
			BaseType: types.BaseType{
				DefaultName: getTypeName(ctx, schema.Title, "ItemsItem"),
				Description: "",
				Inline:      true,
			},
			KeyType:   "string",
			ValueType: &valTyp,
		}
	}

	return &res, nil
}

func getFieldName(srcName string) string {
	return utils.NormalizeGolangName(srcName)
}

func getTypeName(ctx *scancontext.Context, title, suffix string) string {
	n := title
	if n == "" {
		n = strings.Join(ctx.PathStack(), pathSep)
	}
	if suffix != "" {
		n += pathSep + suffix
	}
	return utils.NormalizeGolangName(n)
}
