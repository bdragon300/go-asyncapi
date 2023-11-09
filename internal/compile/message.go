package compile

import (
	"encoding/json"
	"fmt"

	yaml "gopkg.in/yaml.v3"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

type protoMessageBindingsBuilderFunc func(ctx *common.CompileContext, message *Message, bindingsStruct *assemble.Struct, name string) (common.Assembler, error)

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
	Bindings      utils.OrderedMap[string, utils.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`
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
	ctx.PutToCurrentPkg(obj)
	return nil
}

func (m Message) build(ctx *common.CompileContext, messageKey string) (common.Assembler, error) {
	if m.Ref != "" {
		ctx.LogDebug("Ref", "$ref", m.Ref)
		res := assemble.NewRefLinkAsAssembler(m.Ref, common.LinkOriginUser)
		ctx.Linker.Add(res)
		return res, nil
	}

	obj := assemble.Message{
		Name: messageKey,
		OutStruct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(m.Name, "Out"),
				Description: utils.JoinNonemptyStrings("\n", m.Summary+" (Outbound Message)", m.Description),
				Render:      true,
				PackageName: ctx.TopPackageName(),
			},
		},
		InStruct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(m.Name, "In"),
				Description: utils.JoinNonemptyStrings("\n", m.Summary+" (Inbound Message)", m.Description),
				Render:      true,
				PackageName: ctx.TopPackageName(),
			},
		},
		PayloadType:         m.getPayloadType(ctx),
		PayloadHasSchema:    m.Payload != nil && m.Payload.Ref == "",
		HeadersFallbackType: &assemble.Map{KeyType: &assemble.Simple{Name: "string"}, ValueType: &assemble.Simple{Name: "any", IsIface: true}},
	}
	obj.ContentType, _ = lo.Coalesce(m.ContentType, ctx.DefaultContentType)
	ctx.LogDebug(fmt.Sprintf("Message content type is %q", obj.ContentType))
	allServersLnk := assemble.NewListCbLink[*assemble.Server](func(item common.Assembler, path []string) bool {
		_, ok := item.(*assemble.Server)
		return ok
	})
	ctx.Linker.AddMany(allServersLnk)
	obj.AllServers = allServersLnk

	// Link to Headers struct if any
	if m.Headers != nil {
		ctx.LogDebug("Message headers")
		ref := ctx.PathRef() + "/headers"
		obj.HeadersTypeLink = assemble.NewRefLink[*assemble.Struct](ref, common.LinkOriginInternal)
		ctx.Linker.Add(obj.HeadersTypeLink)
	}
	m.setStructFields(ctx, &obj)

	// Bindings
	if m.Bindings.Len() > 0 {
		ctx.LogDebug("Message bindings")
		ctx.IncrementLogCallLvl()
		obj.BindingsStruct = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName(m.Name, "Bindings"),
				Render:      true,
				PackageName: ctx.TopPackageName(),
			},
			Fields: nil,
		}
		for _, e := range m.Bindings.Entries() {
			ctx.LogDebug("Message bindings", "proto", e.Key)
			f, ok := ProtoMessageBindingsBuilder[e.Key]
			if !ok {
				ctx.LogWarn(fmt.Sprintf("Skip unsupported bindings protocol %q", e.Key))
				continue
			}
			ctx.IncrementLogCallLvl()
			protoMethod, err := f(ctx, &m, obj.BindingsStruct, messageKey)
			ctx.DecrementLogCallLvl()
			if err != nil {
				return nil, err
			}
			obj.BindingsStructProtoMethods = append(obj.BindingsStructProtoMethods, protoMethod)
		}
		ctx.DecrementLogCallLvl()
	}
	return &obj, nil
}

func (m Message) setStructFields(ctx *common.CompileContext, langMessage *assemble.Message) {
	fields := []assemble.StructField{
		{
			Name:        "ID",
			Description: "ID is unique string used to identify the message. Case-sensitive.",
			Type:        &assemble.Simple{Name: "string"},
		},
		{Name: "Payload", Type: langMessage.PayloadType},
	}
	if langMessage.HeadersTypeLink != nil {
		ctx.LogDebug("Message headers has a concrete type")
		lnk := assemble.NewRefLinkAsGolangType(langMessage.HeadersTypeLink.Ref(), common.LinkOriginInternal)
		ctx.Linker.Add(lnk)
		fields = append(fields, assemble.StructField{Name: "Headers", Type: lnk})
	} else {
		ctx.LogDebug("Message headers has `any` type")
		fields = append(fields, assemble.StructField{Name: "Headers", Type: langMessage.HeadersFallbackType})
	}

	langMessage.OutStruct.Fields = fields
	langMessage.InStruct.Fields = fields
}

func (m Message) getPayloadType(ctx *common.CompileContext) common.GolangType {
	if m.Payload != nil {
		ctx.LogDebug("Message payload has a concrete type")
		ref := ctx.PathRef() + "/payload"
		lnk := assemble.NewRefLinkAsGolangType(ref, common.LinkOriginInternal)
		ctx.Linker.Add(lnk)
		return lnk
	}

	ctx.LogDebug("Message payload has `any` type")
	return &assemble.Simple{Name: "any", IsIface: true}
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
	Headers utils.OrderedMap[string, utils.Union2[json.RawMessage, yaml.Node]] `json:"headers" yaml:"headers"`
	Payload *utils.Union2[json.RawMessage, yaml.Node]                          `json:"payload" yaml:"payload"`
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
	Bindings      utils.OrderedMap[string, utils.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`
	Examples      []MessageExample                                                   `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
