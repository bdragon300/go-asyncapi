package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"

	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

const pathSep = "_"

type Object struct {
	Type                 *utils.Union2[string, []string]            `json:"type" yaml:"type"`
	AdditionalItems      *utils.Union2[Object, bool]                `json:"additionalItems" yaml:"additionalItems"`
	AdditionalProperties *utils.Union2[Object, bool]                `json:"additionalProperties" yaml:"additionalProperties"`
	AllOf                []Object                                   `json:"allOf" yaml:"allOf"`
	AnyOf                []Object                                   `json:"anyOf" yaml:"anyOf"`
	Const                *utils.Union2[json.RawMessage, yaml.Node]  `json:"const" yaml:"const"`
	Contains             *Object                                    `json:"contains" yaml:"contains"`
	Default              *utils.Union2[json.RawMessage, yaml.Node]  `json:"default" yaml:"default"`
	Definitions          utils.OrderedMap[string, Object]           `json:"definitions" yaml:"definitions"`
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
	PatternProperties    utils.OrderedMap[string, Object]           `json:"patternProperties" yaml:"patternProperties"`
	Properties           utils.OrderedMap[string, Object]           `json:"properties" yaml:"properties"`
	PropertyNames        *Object                                    `json:"propertyNames" yaml:"propertyNames"`
	ReadOnly             *bool                                      `json:"readOnly" yaml:"readOnly"`
	Required             []string                                   `json:"required" yaml:"required"`
	Then                 *Object                                    `json:"then" yaml:"then"`
	Title                string                                     `json:"title" yaml:"title"`
	UniqueItems          *bool                                      `json:"uniqueItems" yaml:"uniqueItems"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (o Object) Build(ctx *scan.Context) error {
	langObj, err := buildLangType(ctx, o, ctx.Top().Flags)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.CurrentPackage().Put(ctx, langObj)
	return nil
}

func buildLangType(ctx *scan.Context, schema Object, flags map[scan.SchemaTag]string) (lang.LangType, error) {
	if schema.Ref != "" {
		res := &lang.DeferTypeRenderer{
			Package:  common.ModelsPackageKind,
			RefQuery: scan.NewRefQuery[lang.LangType](ctx, schema.Ref),
		}
		ctx.RefMgr.Add(res.RefQuery, common.ModelsPackageKind)
		return res, nil
	}

	fixMissingTypeValue(&schema)

	schemaType := schema.Type
	typ := schemaType.V0
	nullable := false             // TODO: x-nullable
	if schemaType.Selector == 1 { // Multiple types, e.g. { "type": [ "object", "array", "null" ] }
		typ, nullable = inspectMultiType(schemaType.V1)
	}

	_, noInline := flags[scan.SchemaTagNoInline]
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
		return &lang.TypeAlias{
			BaseType: lang.BaseType{
				Name:        getTypeName(ctx, schema.Title, ""),
				Description: schema.Description,
				Render:      noInline,
			},
			AliasedType: &lang.Simple{TypeName: "any"},
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

	return &lang.TypeAlias{
		BaseType: lang.BaseType{
			Name:        getTypeName(ctx, schema.Title, ""),
			Description: schema.Description,
			Render:      noInline,
		},
		AliasedType: &lang.Simple{TypeName: langTyp},
		Nullable:    nullable,
	}, nil
}

// fixMissingTypeValue is backwards compatible, guessing the users intention when they didn't specify a type.
func fixMissingTypeValue(s *Object) {
	if s.Type == nil {
		if s.Ref == "" && s.Properties.Len() > 0 {
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

func buildLangStruct(ctx *scan.Context, schema Object, flags map[scan.SchemaTag]string) (*lang.Struct, error) {
	_, noInline := flags[scan.SchemaTagNoInline]
	res := lang.Struct{
		BaseType: lang.BaseType{
			Name:        getTypeName(ctx, schema.Title, ""),
			Description: schema.Description,
			Render:      noInline,
		},
		Fields: make([]lang.StructField, 0),
	}
	// TODO: cache the object name in case any sub-schemas recursively reference it

	// regular properties
	for _, key := range schema.Properties.Keys() {
		langObj, err := buildLangType(ctx, schema.Properties.MustGet(key), map[scan.SchemaTag]string{})
		if err != nil {
			return nil, err
		}
		f := lang.StructField{
			Name:          getFieldName(key),
			Type:          langObj,
			RequiredValue: lo.Contains(schema.Required, key),
			Tags:          nil, // TODO
			Description:   schema.Properties.MustGet(key).Description,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	if schema.AdditionalProperties != nil {
		switch schema.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			langObj, err := buildLangType(ctx, schema.AdditionalProperties.V0, map[scan.SchemaTag]string{})
			if err != nil {
				return nil, err
			}
			f := lang.StructField{
				Name: "AdditionalProperties",
				Type: &lang.Map{
					BaseType: lang.BaseType{
						Name:        getTypeName(ctx, schema.Title, "AdditionalProperties"),
						Description: schema.AdditionalProperties.V0.Description,
						Render:      false,
					},
					KeyType:   &lang.Simple{TypeName: "string"},
					ValueType: langObj,
				},
				RequiredValue: false,
				Tags:          nil, // TODO
				Description:   schema.AdditionalProperties.V0.Description,
			}
			res.Fields = append(res.Fields, f)
		case 1:
			if schema.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := lang.TypeAlias{
					BaseType: lang.BaseType{
						Name:        getTypeName(ctx, schema.Title, "AdditionalPropertiesValue"),
						Description: "",
						Render:      false,
					},
					AliasedType: &lang.Simple{TypeName: "any"},
				}
				f := lang.StructField{
					Name: "AdditionalProperties",
					Type: &lang.Map{
						BaseType: lang.BaseType{
							Name:        getTypeName(ctx, schema.Title, "AdditionalProperties"),
							Description: "",
							Render:      false,
						},
						KeyType:   &lang.Simple{TypeName: "string"},
						ValueType: &valTyp,
					},
					RequiredValue: false,
					Tags:          nil, // TODO
				}
				res.Fields = append(res.Fields, f)
			}
		}
	}

	return &res, nil
}

func buildLangArray(ctx *scan.Context, schema Object, flags map[scan.SchemaTag]string) (*lang.Array, error) {
	_, noInline := flags[scan.SchemaTagNoInline]
	res := lang.Array{
		BaseType: lang.BaseType{
			Name:        getTypeName(ctx, schema.Title, ""),
			Description: schema.Description,
			Render:      noInline,
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
		valTyp := lang.TypeAlias{
			BaseType: lang.BaseType{
				Name:        getTypeName(ctx, schema.Title, "ItemsItemValue"),
				Description: "",
				Render:      false,
			},
			AliasedType: &lang.Simple{TypeName: "any"},
		}
		res.ItemsType = &lang.Map{
			BaseType: lang.BaseType{
				Name:        getTypeName(ctx, schema.Title, "ItemsItem"),
				Description: "",
				Render:      false,
			},
			KeyType:   &lang.Simple{TypeName: "string"},
			ValueType: &valTyp,
		}
	}

	return &res, nil
}

func getFieldName(srcName string) string {
	return utils.NormalizeGolangName(srcName)
}

func getTypeName(ctx *scan.Context, title, suffix string) string {
	n := title
	if n == "" {
		n = strings.Join(ctx.PathStack(), pathSep)
	}
	if suffix != "" {
		n += pathSep + suffix
	}
	return utils.NormalizeGolangName(n)
}
