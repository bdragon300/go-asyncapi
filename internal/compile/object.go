package compile

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"

	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/samber/lo"
)

type Object struct {
	Type                 *utils.Union2[string, []string]            `json:"type" yaml:"type"`
	AdditionalItems      *utils.Union2[Object, bool]                `json:"additionalItems" yaml:"additionalItems"`
	AdditionalProperties *utils.Union2[Object, bool]                `json:"additionalProperties" yaml:"additionalProperties"`
	AllOf                []Object                                   `json:"allOf" yaml:"allOf" cgen:"noinline"`
	AnyOf                []Object                                   `json:"anyOf" yaml:"anyOf" cgen:"noinline"`
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
	ExternalDocs         *ExternalDocumentation                     `json:"externalDocs" yaml:"externalDocs"`
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
	OneOf                []Object                                   `json:"oneOf" yaml:"oneOf" cgen:"noinline"`
	Pattern              string                                     `json:"pattern" yaml:"pattern"`
	PatternProperties    utils.OrderedMap[string, Object]           `json:"patternProperties" yaml:"patternProperties"` // Mapping regex->schema
	Properties           utils.OrderedMap[string, Object]           `json:"properties" yaml:"properties"`
	PropertyNames        *Object                                    `json:"propertyNames" yaml:"propertyNames"`
	ReadOnly             *bool                                      `json:"readOnly" yaml:"readOnly"`
	Required             []string                                   `json:"required" yaml:"required"`
	Then                 *Object                                    `json:"then" yaml:"then"`
	Title                string                                     `json:"title" yaml:"title"`
	UniqueItems          *bool                                      `json:"uniqueItems" yaml:"uniqueItems"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (m Object) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := buildGolangType(ctx, m, ctx.Stack.Top().Flags)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.PutToCurrentPkg(obj)
	return nil
}

func buildGolangType(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (common.GolangType, error) {
	if schema.Ref != "" {
		res := assemble.NewRefLinkAsGolangType(schema.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	if len(schema.OneOf)+len(schema.AnyOf)+len(schema.AllOf) > 0 {
		return buildUnionStruct(ctx, schema) // TODO: process other items that can be set along with oneof/anyof/allof
	}

	fixMissingTypeValue(&schema)

	schemaType := schema.Type
	typ := schemaType.V0
	// TODO: x-nullable
	if schemaType.Selector == 1 { // Multiple types, e.g. { "type": [ "object", "array", "null" ] }
		typ = simplifyMultiType(schemaType.V1)
	}

	_, noInline := flags[common.SchemaTagNoInline]
	// TODO: "type": { "enum": [ "residential", "business" ] }
	// One type: { "type": "object" }
	langTyp := ""
	switch typ {
	case "object":
		res, err := buildLangStruct(ctx, schema, flags)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "array":
		res, err := buildLangArray(ctx, schema, flags)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "null", "":
		return &assemble.NullableType{
			Type:   &assemble.Simple{Type: "any", IsIface: true},
			Render: noInline,
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

	return &assemble.TypeAlias{
		BaseType: assemble.BaseType{
			Name:        ctx.GenerateObjName("", ""),
			Description: schema.Description,
			Render:      noInline,
			Package:     ctx.TopPackageName(),
		},
		AliasedType: &assemble.Simple{Type: langTyp},
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

func simplifyMultiType(schemaType []string) string {
	nullable := lo.Contains(schemaType, "null")
	typs := lo.Reject(schemaType, func(item string, _ int) bool { return item == "null" }) // Throw out null (if any)
	switch {
	case len(typs) > 1: // More than one type along with null -> 'any'
		return ""
	case len(typs) > 0: // One type along with null -> pointer to this type
		return typs[0]
	case nullable: // Null only -> 'any', that can be only nil
		return "null"
	default:
		panic("Empty schema type")
	}
}

func buildLangStruct(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (*assemble.Struct, error) {
	_, noInline := flags[common.SchemaTagNoInline]
	res := assemble.Struct{
		BaseType: assemble.BaseType{
			Name:        ctx.GenerateObjName("", ""),
			Description: schema.Description,
			Render:      noInline,
			Package:     ctx.TopPackageName(),
		},
	}
	// TODO: cache the object name in case any sub-schemas recursively reference it

	// regular properties
	for _, entry := range schema.Properties.Entries() {
		ref := path.Join(ctx.PathRef(), "properties", entry.Key)
		langObj := assemble.NewRefLinkAsGolangType(ref)
		ctx.Linker.Add(langObj)
		f := assemble.StructField{
			Name:         utils.ToGolangName(entry.Key, true),
			Type:         langObj,
			ForcePointer: lo.Contains(schema.Required, entry.Key),
			Tags:         nil, // TODO
			Description:  entry.Value.Description,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	if schema.AdditionalProperties != nil {
		switch schema.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			ref := path.Join(ctx.PathRef(), "additionalProperties")
			langObj := assemble.NewRefLinkAsGolangType(ref)
			f := assemble.StructField{
				Name: "AdditionalProperties",
				Type: &assemble.Map{
					BaseType: assemble.BaseType{
						Name:        ctx.GenerateObjName("", "AdditionalProperties"),
						Description: schema.AdditionalProperties.V0.Description,
						Render:      false,
						Package:     ctx.TopPackageName(),
					},
					KeyType:   &assemble.Simple{Type: "string"},
					ValueType: langObj,
				},
				Tags:        nil, // TODO
				Description: schema.AdditionalProperties.V0.Description,
			}
			res.Fields = append(res.Fields, f)
		case 1:
			if schema.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := assemble.TypeAlias{
					BaseType: assemble.BaseType{
						Name:        ctx.GenerateObjName("", "AdditionalPropertiesValue"),
						Description: "",
						Render:      false,
						Package:     ctx.TopPackageName(),
					},
					AliasedType: &assemble.Simple{Type: "any", IsIface: true},
				}
				f := assemble.StructField{
					Name: "AdditionalProperties",
					Type: &assemble.Map{
						BaseType: assemble.BaseType{
							Name:        ctx.GenerateObjName("", "AdditionalProperties"),
							Description: "",
							Render:      false,
							Package:     ctx.TopPackageName(),
						},
						KeyType:   &assemble.Simple{Type: "string"},
						ValueType: &valTyp,
					},
					Tags: nil, // TODO
				}
				res.Fields = append(res.Fields, f)
			}
		}
	}

	return &res, nil
}

func buildLangArray(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (*assemble.Array, error) {
	_, noInline := flags[common.SchemaTagNoInline]
	res := assemble.Array{
		BaseType: assemble.BaseType{
			Name:        ctx.GenerateObjName("", ""),
			Description: schema.Description,
			Render:      noInline,
			Package:     ctx.TopPackageName(),
		},
		ItemsType: nil,
	}

	switch {
	case schema.Items != nil && schema.Items.Selector == 0: // Only one "type:" of items
		ref := path.Join(ctx.PathRef(), "items")
		res.ItemsType = assemble.NewRefLinkAsGolangType(ref)
	case schema.Items == nil || schema.Items.Selector == 1: // No items or Several types for each item sequentially
		valTyp := assemble.TypeAlias{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", "ItemsItemValue"),
				Description: "",
				Render:      false,
				Package:     ctx.TopPackageName(),
			},
			AliasedType: &assemble.Simple{Type: "any", IsIface: true},
		}
		res.ItemsType = &assemble.Map{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", "ItemsItem"),
				Description: "",
				Render:      false,
				Package:     ctx.TopPackageName(),
			},
			KeyType:   &assemble.Simple{Type: "string"},
			ValueType: &valTyp,
		}
	}

	return &res, nil
}

func buildUnionStruct(ctx *common.CompileContext, schema Object) (*assemble.UnionStruct, error) {
	res := assemble.UnionStruct{
		Struct: assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", ""),
				Description: schema.Description,
				Render:      true, // Always render unions as separate types
				Package:     ctx.TopPackageName(),
			},
		},
	}

	res.Fields = lo.Times(len(schema.OneOf), func(index int) assemble.StructField {
		ref := path.Join(ctx.PathRef(), "oneOf", strconv.Itoa(index))
		langTyp := assemble.NewRefLinkAsGolangType(ref)
		ctx.Linker.Add(langTyp)
		return assemble.StructField{Type: langTyp, ForcePointer: true, Tags: nil}
	})
	res.Fields = append(res.Fields, lo.Times(len(schema.AnyOf), func(index int) assemble.StructField {
		ref := path.Join(ctx.PathRef(), "anyOf", strconv.Itoa(index))
		langTyp := assemble.NewRefLinkAsGolangType(ref)
		ctx.Linker.Add(langTyp)
		return assemble.StructField{Type: langTyp, ForcePointer: true, Tags: nil}
	})...)
	res.Fields = append(res.Fields, lo.Times(len(schema.AllOf), func(index int) assemble.StructField {
		ref := path.Join(ctx.PathRef(), "allOf", strconv.Itoa(index))
		langTyp := assemble.NewRefLinkAsGolangType(ref)
		ctx.Linker.Add(langTyp)
		return assemble.StructField{Type: langTyp, Tags: nil}
	})...)

	return &res, nil
}

