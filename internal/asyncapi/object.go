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

func (o Object) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := o.build(ctx, ctx.Stack.Top().Flags)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (o Object) build(ctx *common.CompileContext, flags map[common.SchemaTag]string) (common.GolangType, error) {
	if o.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", o.Ref)
		res := render.NewGolangTypePromise(o.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	if len(o.OneOf)+len(o.AnyOf)+len(o.AllOf) > 0 {
		ctx.Logger.Trace("Object is union struct")
		return o.buildUnionStruct(ctx) // TODO: process other items that can be set along with oneof/anyof/allof
	}

	if o.Type == nil {
		o = o.fixMissingObjectType(ctx)
	}

	typeName, nullable, err := o.getTypeName(ctx)
	if err != nil {
		return nil, err
	}

	nullable = nullable || lo.FromPtr(o.XNullable)

	// TODO: "type": { "enum": [ "residential", "business" ] }
	// One type: { "type": "object" }
	golangType, err := o.buildGolangType(ctx, flags, typeName)
	if err != nil {
		return nil, err
	}

	if nullable {
		ctx.Logger.Trace("Object is nullable, make it pointer")
		_, directRender := flags[common.SchemaTagDirectRender]
		golangType = &render.Pointer{Type: golangType, DirectRender: directRender}
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
			err = types.CompileError{Err: errors.New("empty object type"), Path: ctx.PathRef()}
			return
		}
		ctx.Logger.Trace(fmt.Sprintf("Multitype object type inferred as %q", typeName))
	}
	return
}

func (o Object) buildGolangType(ctx *common.CompileContext, flags map[common.SchemaTag]string, typeName string) (golangType common.GolangType, err error) {
	var aliasedType *render.Simple

	switch typeName {
	case "object":
		ctx.Logger.Trace("Object is struct")
		ctx.Logger.NextCallLevel()
		golangType, err = o.buildLangStruct(ctx, flags)
		ctx.Logger.PrevCallLevel()
		if err != nil {
			return nil, err
		}
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
		golangType = &render.Simple{Name: "any", IsIface: true}
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
		_, directRender := flags[common.SchemaTagDirectRender]
		golangType = &render.TypeAlias{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(o.Title, ""),
				Description:  o.Description,
				DirectRender: directRender,
				PackageName:  ctx.TopPackageName(),
			},
			AliasedType: aliasedType,
		}
	}
	return golangType, nil
}

// fixMissingObjectType is backwards compatible, guessing the users intention when they didn't specify a type.
func (o Object) fixMissingObjectType(ctx *common.CompileContext) Object {
	switch {
	case o.Ref == "" && o.Properties.Len() > 0:
		ctx.Logger.Trace("Object type is empty, determined `object` because of `properties` presence")
		o.Type = types.ToUnion2[string, []string]("object")
	case o.Items != nil: // TODO: fix type when AllOf, AnyOf, OneOf
		ctx.Logger.Trace("Object type is empty, determined `array` because of `items` presence")
		o.Type = types.ToUnion2[string, []string]("array")
	default:
		ctx.Logger.Trace("Object type is empty, guessing it `object` by default")
		o.Type = types.ToUnion2[string, []string]("object")
	}
	return o
}

func (o Object) buildLangStruct(ctx *common.CompileContext, flags map[common.SchemaTag]string) (*render.Struct, error) {
	_, noInline := flags[common.SchemaTagDirectRender]
	res := render.Struct{
		BaseType: render.BaseType{
			Name:         ctx.GenerateObjName(o.Title, ""),
			Description:  o.Description,
			DirectRender: noInline,
			PackageName:  ctx.TopPackageName(),
		},
	}
	// TODO: cache the object name in case any sub-schemas recursively reference it

	var messagesPrm *render.ListPromise[*render.Message]
	// Collect all messages to retrieve struct field tags
	if ctx.TopPackageName() == "models" { // TODO: fix hardcode
		messagesPrm = render.NewListCbPromise[*render.Message](func(item common.Renderer, _ []string) bool {
			_, ok := item.(*render.Message)
			return ok
		})
		ctx.PutListPromise(messagesPrm)
	}

	// regular properties
	for _, entry := range o.Properties.Entries() {
		ctx.Logger.Trace("Object property", "name", entry.Key)
		ref := path.Join(ctx.PathRef(), "properties", entry.Key)
		prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)

		var langObj common.GolangType = prm
		if lo.Contains(o.Required, entry.Key) {
			langObj = &render.Pointer{Type: langObj}
		}

		f := render.StructField{
			Name:        utils.ToGolangName(entry.Key, true),
			MarshalName: entry.Key,
			Type:        langObj,
			Description: entry.Value.Description,
			TagsSource:  messagesPrm,
		}
		res.Fields = append(res.Fields, f)
	}

	// additionalProperties with typed sub-schema
	if o.AdditionalProperties != nil {
		switch o.AdditionalProperties.Selector {
		case 0: // "additionalProperties:" is an object
			ctx.Logger.Trace("Object additional properties as an object")
			ref := path.Join(ctx.PathRef(), "additionalProperties")
			langObj := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
			f := render.StructField{
				Name: "AdditionalProperties",
				Type: &render.Map{
					BaseType: render.BaseType{
						Name:         ctx.GenerateObjName(o.Title, "AdditionalProperties"),
						Description:  o.AdditionalProperties.V0.Description,
						DirectRender: false,
						PackageName:  ctx.TopPackageName(),
					},
					KeyType:   &render.Simple{Name: "string"},
					ValueType: langObj,
				},
				Description: o.AdditionalProperties.V0.Description,
			}
			res.Fields = append(res.Fields, f)
		case 1:
			ctx.Logger.Trace("Object additional properties as boolean flag")
			if o.AdditionalProperties.V1 { // "additionalProperties: true" -- allow any additional properties
				valTyp := render.TypeAlias{
					BaseType: render.BaseType{
						Name:         ctx.GenerateObjName(o.Title, "AdditionalPropertiesValue"),
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
							Name:         ctx.GenerateObjName(o.Title, "AdditionalProperties"),
							Description:  "",
							DirectRender: false,
							PackageName:  ctx.TopPackageName(),
						},
						KeyType:   &render.Simple{Name: "string"},
						ValueType: &valTyp,
					},
					TagsSource: messagesPrm,
				}
				res.Fields = append(res.Fields, f)
			}
		}
	}

	return &res, nil
}

func (o Object) buildLangArray(ctx *common.CompileContext, flags map[common.SchemaTag]string) (*render.Array, error) {
	_, noInline := flags[common.SchemaTagDirectRender]
	res := render.Array{
		BaseType: render.BaseType{
			Name:         ctx.GenerateObjName(o.Title, ""),
			Description:  o.Description,
			DirectRender: noInline,
			PackageName:  ctx.TopPackageName(),
		},
		ItemsType: nil,
	}

	switch {
	case o.Items != nil && o.Items.Selector == 0: // Only one "type:" of items
		ctx.Logger.Trace("Object items (single type)")
		ref := path.Join(ctx.PathRef(), "items")
		res.ItemsType = render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
	case o.Items == nil || o.Items.Selector == 1: // No items or Several types for each item sequentially
		ctx.Logger.Trace("Object items (zero or several types)")
		valTyp := render.TypeAlias{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(o.Title, "ItemsItemValue"),
				Description:  "",
				DirectRender: false,
				PackageName:  ctx.TopPackageName(),
			},
			AliasedType: &render.Simple{Name: "any", IsIface: true},
		}
		res.ItemsType = &render.Map{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(o.Title, "ItemsItem"),
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

func (o Object) buildUnionStruct(ctx *common.CompileContext) (*render.UnionStruct, error) {
	res := render.UnionStruct{
		Struct: render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(o.Title, ""),
				Description:  o.Description,
				DirectRender: true, // Always render unions as separate types
				PackageName:  ctx.TopPackageName(),
			},
		},
	}

	// Collect all messages to retrieve struct field tags
	messagesPrm := render.NewListCbPromise[*render.Message](func(item common.Renderer, _ []string) bool {
		_, ok := item.(*render.Message)
		return ok
	})
	ctx.PutListPromise(messagesPrm)

	res.Fields = lo.Times(len(o.OneOf), func(index int) render.StructField {
		ref := path.Join(ctx.PathRef(), "oneOf", strconv.Itoa(index))
		prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return render.StructField{Type: &render.Pointer{Type: prm}}
	})
	res.Fields = append(res.Fields, lo.Times(len(o.AnyOf), func(index int) render.StructField {
		ref := path.Join(ctx.PathRef(), "anyOf", strconv.Itoa(index))
		prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return render.StructField{Type: &render.Pointer{Type: prm}}
	})...)
	res.Fields = append(res.Fields, lo.Times(len(o.AllOf), func(index int) render.StructField {
		ref := path.Join(ctx.PathRef(), "allOf", strconv.Itoa(index))
		prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return render.StructField{Type: prm}
	})...)

	return &res, nil
}
