package asyncapi

import (
	"encoding/json"
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	yaml "gopkg.in/yaml.v3"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

type protoMessageBindingsBuilderFunc func(ctx *common.CompileContext, message *Message, bindingsStruct *render.Struct, name string) (common.Renderer, error)

var ProtoMessageBindingsBuilder = map[string]protoMessageBindingsBuilderFunc{}

type Message struct {
	MessageID     string                                                             `json:"messageId" yaml:"messageId"`
	Headers       *Object                                                            `json:"headers" yaml:"headers"`
	Payload       *Object                                                            `json:"payload" yaml:"payload"` // TODO: other formats
	CorrelationID *CorrelationID                                                     `json:"correlationId" yaml:"correlationId"`
	SchemaFormat  string                                                             `json:"schemaFormat" yaml:"schemaFormat"`
	ContentType   string                                                             `json:"contentType" yaml:"contentType"`
	Name          string                                                             `json:"name" yaml:"name"`
	Title         string                                                             `json:"title" yaml:"title"`
	Summary       string                                                             `json:"summary" yaml:"summary"`
	Description   string                                                             `json:"description" yaml:"description"`
	Tags          []Tag                                                              `json:"tags" yaml:"tags"`
	ExternalDocs  *ExternalDocumentation                                             `json:"externalDocs" yaml:"externalDocs"`
	Bindings      types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`
	Examples      []MessageExample                                                   `json:"examples" yaml:"examples"`
	Traits        []MessageTrait                                                     `json:"traits" yaml:"traits"`

	XGoType *types.Union2[string, xGoType] `json:"x-go-type" yaml:"x-go-type"`
	XGoName string                         `json:"x-go-name" yaml:"x-go-name"`
	XIgnore bool                           `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (m Message) Compile(ctx *common.CompileContext) error {
	ctx.SetTopObjName(ctx.Stack.Top().Path)
	obj, err := m.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (m Message) build(ctx *common.CompileContext, messageKey string) (common.Renderer, error) {
	if m.XIgnore {
		ctx.Logger.Debug("Message denoted to be ignored")
		return &render.Simple{Name: "any", IsIface: true}, nil
	}
	if m.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", m.Ref)
		res := render.NewRendererPromise(m.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	if m.XGoType != nil {
		t := buildXGoType(m.XGoType)
		ctx.Logger.Trace("Message is a custom type", "type", t.String())
		return t, nil
	}

	objName, _ := lo.Coalesce(m.XGoName, messageKey)
	obj := render.Message{
		Name: objName,
		OutStruct: &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(m.Name, "Out"),
				Description:  utils.JoinNonemptyStrings("\n", m.Summary+" (Outbound Message)", m.Description),
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
		},
		InStruct: &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(m.Name, "In"),
				Description:  utils.JoinNonemptyStrings("\n", m.Summary+" (Inbound Message)", m.Description),
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
		},
		PayloadType:         m.getPayloadType(ctx),
		PayloadHasSchema:    m.Payload != nil && m.Payload.Ref == "",
		HeadersFallbackType: &render.Map{KeyType: &render.Simple{Name: "string"}, ValueType: &render.Simple{Name: "any", IsIface: true}},
	}
	obj.ContentType, _ = lo.Coalesce(m.ContentType, ctx.Storage.DefaultContentType())
	ctx.Logger.Trace(fmt.Sprintf("Message content type is %q", obj.ContentType))
	allServersPrm := render.NewListCbPromise[*render.Server](func(item common.Renderer, path []string) bool {
		_, ok := item.(*render.Server)
		return ok
	})
	ctx.PutListPromise(allServersPrm)
	obj.AllServers = allServersPrm

	// Link to Headers struct if any
	if m.Headers != nil {
		ctx.Logger.Trace("Message headers")
		ref := ctx.PathRef() + "/headers"
		obj.HeadersTypePromise = render.NewPromise[*render.Struct](ref, common.PromiseOriginInternal)
		ctx.PutPromise(obj.HeadersTypePromise)
	}
	m.setStructFields(ctx, &obj)

	// Bindings
	if m.Bindings.Len() > 0 {
		ctx.Logger.Trace("Message bindings")
		ctx.Logger.NextCallLevel()
		obj.BindingsStruct = &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(m.Name, "Bindings"),
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
			Fields: nil,
		}
		for _, k := range m.Bindings.Keys() {
			ctx.Logger.Trace("Message bindings", "proto", k)
			f, ok := ProtoMessageBindingsBuilder[k]
			if !ok {
				ctx.Logger.Warn(fmt.Sprintf("Skip unsupported bindings protocol %q", k))
				continue
			}
			ctx.Logger.NextCallLevel()
			protoMethod, err := f(ctx, &m, obj.BindingsStruct, messageKey)
			ctx.Logger.PrevCallLevel()
			if err != nil {
				return nil, err
			}
			obj.BindingsStructProtoMethods.Set(k, protoMethod)
		}
		ctx.Logger.PrevCallLevel()
	}

	// Link to CorrelationID if any
	if m.CorrelationID != nil {
		ctx.Logger.Trace("Message correlationId")
		ref := ctx.PathRef() + "/correlationId"
		obj.CorrelationIDPromise = render.NewPromise[*render.CorrelationID](ref, common.PromiseOriginInternal)
		ctx.PutPromise(obj.CorrelationIDPromise)
	}
	return &obj, nil
}

func (m Message) setStructFields(ctx *common.CompileContext, langMessage *render.Message) {
	fields := []render.StructField{
		{Name: "Payload", Type: langMessage.PayloadType},
	}
	if langMessage.HeadersTypePromise != nil {
		ctx.Logger.Trace("Message headers has a concrete type")
		prm := render.NewGolangTypePromise(langMessage.HeadersTypePromise.Ref(), common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		fields = append(fields, render.StructField{Name: "Headers", Type: prm})
	} else {
		ctx.Logger.Trace("Message headers has `any` type")
		fields = append(fields, render.StructField{Name: "Headers", Type: langMessage.HeadersFallbackType})
	}

	langMessage.OutStruct.Fields = fields
	langMessage.InStruct.Fields = fields
}

func (m Message) getPayloadType(ctx *common.CompileContext) common.GolangType {
	if m.Payload != nil {
		ctx.Logger.Trace("Message payload has a concrete type")
		ref := ctx.PathRef() + "/payload"
		prm := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return prm
	}

	ctx.Logger.Trace("Message payload has `any` type")
	return &render.Simple{Name: "any", IsIface: true}
}

type Tag struct {
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description" yaml:"description"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
}

type MessageExample struct {
	Headers types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"headers" yaml:"headers"`
	Payload *types.Union2[json.RawMessage, yaml.Node]                          `json:"payload" yaml:"payload"`
	Name    string                                                             `json:"name" yaml:"name"`
	Summary string                                                             `json:"summary" yaml:"summary"`
}

type MessageTrait struct {
	MessageID     string                                                             `json:"messageId" yaml:"messageId"`
	Headers       *Object                                                            `json:"headers" yaml:"headers"`
	CorrelationID *CorrelationID                                                     `json:"correlationId" yaml:"correlationId"`
	SchemaFormat  string                                                             `json:"schemaFormat" yaml:"schemaFormat"`
	ContentType   string                                                             `json:"contentType" yaml:"contentType"`
	Name          string                                                             `json:"name" yaml:"name"`
	Title         string                                                             `json:"title" yaml:"title"`
	Summary       string                                                             `json:"summary" yaml:"summary"`
	Description   string                                                             `json:"description" yaml:"description"`
	Tags          []Tag                                                              `json:"tags" yaml:"tags"`
	ExternalDocs  *ExternalDocumentation                                             `json:"externalDocs" yaml:"externalDocs"`
	Bindings      types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`
	Examples      []MessageExample                                                   `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
