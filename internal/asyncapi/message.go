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
	if m.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", m.Ref)
		res := render.NewRendererPromise(m.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	obj := render.Message{
		Name: messageKey,
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
	obj.ContentType, _ = lo.Coalesce(m.ContentType, ctx.ObjectsStore.DefaultContentType())
	ctx.Logger.Trace(fmt.Sprintf("Message content type is %q", obj.ContentType))
	allServersLnk := render.NewListCbPromise[*render.Server](func(item common.Renderer, path []string) bool {
		_, ok := item.(*render.Server)
		return ok
	})
	ctx.PutListPromise(allServersLnk)
	obj.AllServers = allServersLnk

	// Link to Headers struct if any
	if m.Headers != nil {
		ctx.Logger.Trace("Message headers")
		ref := ctx.PathRef() + "/headers"
		obj.HeadersTypeLink = render.NewPromise[*render.Struct](ref, common.PromiseOriginInternal)
		ctx.PutPromise(obj.HeadersTypeLink)
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
		for _, e := range m.Bindings.Entries() {
			ctx.Logger.Trace("Message bindings", "proto", e.Key)
			f, ok := ProtoMessageBindingsBuilder[e.Key]
			if !ok {
				ctx.Logger.Warn(fmt.Sprintf("Skip unsupported bindings protocol %q", e.Key))
				continue
			}
			ctx.Logger.NextCallLevel()
			protoMethod, err := f(ctx, &m, obj.BindingsStruct, messageKey)
			ctx.Logger.PrevCallLevel()
			if err != nil {
				return nil, err
			}
			obj.BindingsStructProtoMethods.Set(e.Key, protoMethod)
		}
		ctx.Logger.PrevCallLevel()
	}
	return &obj, nil
}

func (m Message) setStructFields(ctx *common.CompileContext, langMessage *render.Message) {
	fields := []render.StructField{
		{Name: "Payload", Type: langMessage.PayloadType},
	}
	if langMessage.HeadersTypeLink != nil {
		ctx.Logger.Trace("Message headers has a concrete type")
		lnk := render.NewGolangTypePromise(langMessage.HeadersTypeLink.Ref(), common.PromiseOriginInternal)
		ctx.PutPromise(lnk)
		fields = append(fields, render.StructField{Name: "Headers", Type: lnk})
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
		lnk := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(lnk)
		return lnk
	}

	ctx.Logger.Trace("Message payload has `any` type")
	return &render.Simple{Name: "any", IsIface: true}
}

type CorrelationID struct {
	Description string `json:"description" yaml:"description"`
	Location    string `json:"location" yaml:"location"`

	Ref string `json:"$ref" yaml:"$ref"`
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
