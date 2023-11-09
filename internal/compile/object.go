package compile

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"

	yaml "gopkg.in/yaml.v3"

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
		return err
	}
	ctx.PutToCurrentPkg(obj)
	return nil
}

func buildGolangType(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (common.GolangType, error) {
	if schema.Ref != "" {
		ctx.LogDebug("Ref", "$ref", schema.Ref)
		res := assemble.NewRefLinkAsGolangType(schema.Ref, common.LinkOriginUser)
		ctx.Linker.Add(res)
		return res, nil
	}

	if len(schema.OneOf)+len(schema.AnyOf)+len(schema.AllOf) > 0 {
		ctx.LogDebug("Object is union struct")
		return buildUnionStruct(ctx, schema) // TODO: process other items that can be set along with oneof/anyof/allof
	}

	fixMissingTypeValue(ctx, &schema)

	schemaType := schema.Type
	typ := schemaType.V0
	// TODO: x-nullable
	if schemaType.Selector == 1 { // Multiple types, e.g. { "type": [ "object", "array", "null" ] }
		t, err := simplifyMultiType(schemaType.V1)
		if err != nil {
			return nil, common.CompileError{Err: err, Path: ctx.PathRef()}
		}
		typ = t
		ctx.LogDebug(fmt.Sprintf("Multitype object type inferred as %q", typ))
	}

	_, noInline := flags[common.SchemaTagNoInline]
	// TODO: "type": { "enum": [ "residential", "business" ] }
	// One type: { "type": "object" }
	langTyp := ""
	switch typ {
	case "object":
		ctx.LogDebug("Object is struct")
		ctx.IncrementLogCallLvl()
		res, err := buildLangStruct(ctx, schema, flags)
		ctx.DecrementLogCallLvl()
		return res, err
	case "array":
		ctx.LogDebug("Object is array")
		ctx.IncrementLogCallLvl()
		res, err := buildLangArray(ctx, schema, flags)
		ctx.DecrementLogCallLvl()
		return res, err
	case "null", "":
		ctx.LogDebug("Object is nullable any")
		res := &assemble.NullableType{
			Type:   &assemble.Simple{Name: "any", IsIface: true},
			Render: noInline,
		}
		return res, nil
	case "boolean":
		ctx.LogDebug("Object is bool")
		langTyp = "bool"
	case "integer":
		// TODO: "format:"
		ctx.LogDebug("Object is int")
		langTyp = "int"
	case "number":
		// TODO: "format:"
		ctx.LogDebug("Object is float64")
		langTyp = "float64"
	case "string":
		ctx.LogDebug("Object is string")
		langTyp = "string"
	default:
		return nil, common.CompileError{Err: fmt.Errorf("unknown jsonschema type %q", typ), Path: ctx.PathRef()}
	}

	return &assemble.TypeAlias{
		BaseType: assemble.BaseType{
			Name:        ctx.GenerateObjName(schema.Title, ""),
			Description: schema.Description,
			Render:      noInline,
			PackageName: ctx.TopPackageName(),
		},
		AliasedType: &assemble.Simple{Name: langTyp},
	}, nil
}

// fixMissingTypeValue is backwards compatible, guessing the users intention when they didn't specify a type.
func fixMissingTypeValue(ctx *common.CompileContext, s *Object) {
	if s.Type == nil {
		if s.Ref == "" && s.Properties.Len() > 0 {
			ctx.LogDebug("Object type is empty, determined `object` because of `properties` presence")
			s.Type = utils.ToUnion2[string, []string]("object")
			return
		}
		// TODO: fix type when AllOf, AnyOf, OneOf
		if s.Items != nil {
			ctx.LogDebug("Object type is empty, determined `array` because of `items` presence")
			s.Type = utils.ToUnion2[string, []string]("array")
			return
		}

		ctx.LogDebug("Object type is empty, guessing it `object` by default")
		s.Type = utils.ToUnion2[string, []string]("object")
	}
}

func simplifyMultiType(schemaType []string) (string, error) {
	nullable := lo.Contains(schemaType, "null")
	typs := lo.Reject(schemaType, func(item string, _ int) bool { return item == "null" }) // Throw out null (if any)
	switch {
	case len(typs) > 1: // More than one type along with null -> 'any'
		return "", nil
	case len(typs) > 0: // One type along with null -> pointer to this type
		return typs[0], nil
	case nullable: // Null only -> 'any', that can be only nil
		return "null", nil
	default:
		return "", errors.New("empty object type")
	}
}

func buildLangStruct(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (*assemble.Struct, error) {
	_, noInline := flags[common.SchemaTagNoInline]
	res := assemble.Struct{
		BaseType: assemble.BaseType{
			Name:        ctx.GenerateObjName(schema.Title, ""),
			Description: schema.Description,
			Render:      noInline,
			PackageName: ctx.TopPackageName(),
		},
	}
	// TODO: cache the object name in case any sub-schemas recursively reference it

	var msgLinks *assemble.LinkList[*assemble.Message]
	// Collect all messages to retrieve struct field tags
	if ctx.TopPackageName() == "models" {
		msgLinks = assemble.NewListCbLink[*assemble.Message](func(item common.Assembler, _ []string) bool {
			_, ok := item.(*assemble.Message)
			return ok
		})
		ctx.Linker.AddMany(msgLinks)
	}

	// regular properties
	for _, entry := range schema.Properties.Entries() {
		ctx.LogDebug("Object property", "name", entry.Key)
		ref := path.Join(ctx.PathRef(), "properties", entry.Key)
		langObj := assemble.NewRefLinkAsGolangType(ref, common.LinkOriginInternal)
		ctx.Linker.Add(langObj)
		f := assemble.StructField{
			Name:         utils.ToGolangName(entry.Key, true),
			MarshalName:  entry.Key,
			Type:         langObj,
			ForcePointer: lo.Contains(schema.Required, entry.Key),
			Description:  entry.Value.Description,
			TagsSource:   msgLinks,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	if schema.AdditionalProperties != nil {
		switch schema.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			ctx.LogDebug("Object additional properties as an object")
			ref := path.Join(ctx.PathRef(), "additionalProperties")
			langObj := assemble.NewRefLinkAsGolangType(ref, common.LinkOriginInternal)
			f := assemble.StructField{
				Name: "AdditionalProperties",
				Type: &assemble.Map{
					BaseType: assemble.BaseType{
						Name:        ctx.GenerateObjName(schema.Title, "AdditionalProperties"),
						Description: schema.AdditionalProperties.V0.Description,
						Render:      false,
						PackageName: ctx.TopPackageName(),
					},
					KeyType:   &assemble.Simple{Name: "string"},
					ValueType: langObj,
				},
				Description: schema.AdditionalProperties.V0.Description,
			}
			res.Fields = append(res.Fields, f)
		case 1:
			ctx.LogDebug("Object additional properties as boolean flag")
			if schema.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := assemble.TypeAlias{
					BaseType: assemble.BaseType{
						Name:        ctx.GenerateObjName(schema.Title, "AdditionalPropertiesValue"),
						Description: "",
						Render:      false,
						PackageName: ctx.TopPackageName(),
					},
					AliasedType: &assemble.Simple{Name: "any", IsIface: true},
				}
				f := assemble.StructField{
					Name: "AdditionalProperties",
					Type: &assemble.Map{
						BaseType: assemble.BaseType{
							Name:        ctx.GenerateObjName(schema.Title, "AdditionalProperties"),
							Description: "",
							Render:      false,
							PackageName: ctx.TopPackageName(),
						},
						KeyType:   &assemble.Simple{Name: "string"},
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

func buildLangArray(ctx *common.CompileContext, schema Object, flags map[common.SchemaTag]string) (*assemble.Array, error) {
	_, noInline := flags[common.SchemaTagNoInline]
	res := assemble.Array{
		BaseType: assemble.BaseType{
			Name:        ctx.GenerateObjName(schema.Title, ""),
			Description: schema.Description,
			Render:      noInline,
			PackageName: ctx.TopPackageName(),
		},
		ItemsType: nil,
	}

	switch {
	case schema.Items != nil && schema.Items.Selector == 0: // Only one "type:" of items
		ctx.LogDebug("Object items (single type)")
		ref := path.Join(ctx.PathRef(), "items")
		res.ItemsType = assemble.NewRefLinkAsGolangType(ref, common.LinkOriginInternal)
	case schema.Items == nil || schema.Items.Selector == 1: // No items or Several types for each item sequentially
		ctx.LogDebug("Object items (zero or several types)")
		valTyp := assemble.TypeAlias{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(schema.Title, "ItemsItemValue"),
				Description: "",
				Render:      false,
				PackageName: ctx.TopPackageName(),
			},
			AliasedType: &assemble.Simple{Name: "any", IsIface: true},
		}
		res.ItemsType = &assemble.Map{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(schema.Title, "ItemsItem"),
				Description: "",
				Render:      false,
				PackageName: ctx.TopPackageName(),
			},
			KeyType:   &assemble.Simple{Name: "string"},
			ValueType: &valTyp,
		}
	}

	return &res, nil
}

func buildUnionStruct(ctx *common.CompileContext, schema Object) (*assemble.UnionStruct, error) {
	res := assemble.UnionStruct{
		Struct: assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(schema.Title, ""),
				Description: schema.Description,
				Render:      true, // Always render unions as separate types
				PackageName: ctx.TopPackageName(),
			},
		},
	}

	// Collect all messages to retrieve struct field tags
	msgLinks := assemble.NewListCbLink[*assemble.Message](func(item common.Assembler, _ []string) bool {
		_, ok := item.(*assemble.Message)
		return ok
	})
	ctx.Linker.AddMany(msgLinks)

	res.Fields = lo.Times(len(schema.OneOf), func(index int) assemble.StructField {
		ref := path.Join(ctx.PathRef(), "oneOf", strconv.Itoa(index))
		langTyp := assemble.NewRefLinkAsGolangType(ref, common.LinkOriginInternal)
		ctx.Linker.Add(langTyp)
		return assemble.StructField{Type: langTyp, ForcePointer: true}
	})
	res.Fields = append(res.Fields, lo.Times(len(schema.AnyOf), func(index int) assemble.StructField {
		ref := path.Join(ctx.PathRef(), "anyOf", strconv.Itoa(index))
		langTyp := assemble.NewRefLinkAsGolangType(ref, common.LinkOriginInternal)
		ctx.Linker.Add(langTyp)
		return assemble.StructField{Type: langTyp, ForcePointer: true}
	})...)
	res.Fields = append(res.Fields, lo.Times(len(schema.AllOf), func(index int) assemble.StructField {
		ref := path.Join(ctx.PathRef(), "allOf", strconv.Itoa(index))
		langTyp := assemble.NewRefLinkAsGolangType(ref, common.LinkOriginInternal)
		ctx.Linker.Add(langTyp)
		return assemble.StructField{Type: langTyp}
	})...)

	return &res, nil
}
