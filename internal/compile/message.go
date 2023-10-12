package compile

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

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
	ctx.SetTopObjName(ctx.Stack.Top().Path) // TODO: use title
	obj, err := m.build(ctx)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.PutToCurrentPkg(obj)
	return nil
}

func (m Message) build(ctx *common.CompileContext) (common.Assembler, error) {
	if m.ContentType != "" && m.ContentType != "application/json" {
		return nil, fmt.Errorf("now is supported only application/json") // TODO: support other content types
	}
	if m.Ref != "" {
		res := assemble.NewRefLinkAsAssembler(m.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	obj := assemble.Message{
		OutStruct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", "Out"),
				Description: utils.JoinNonemptyStrings("\n", m.Summary+" (Outbound Message)", m.Description),
				Render:      true,
				Package:     ctx.TopPackageName(),
			},
		},
		InStruct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", "In"),
				Description: utils.JoinNonemptyStrings("\n", m.Summary+" (Inbound Message)", m.Description),
				Render:      true,
				Package:     ctx.TopPackageName(),
			},
		},
		PayloadType:         m.getPayloadType(ctx),
		PayloadHasSchema:    m.Payload != nil,
		HeadersFallbackType: &assemble.Map{KeyType: &assemble.Simple{Type: "string"}, ValueType: &assemble.Simple{Type: "any", IsIface: true}},
	}
	allServersLnk := assemble.NewListCbLink[*assemble.Server](func(item common.Assembler, path []string) bool {
		_, ok := item.(*assemble.Server)
		return ok
	})
	ctx.Linker.AddMany(allServersLnk)
	obj.AllServers = allServersLnk

	// Link to Headers struct if any
	if m.Headers != nil {
		ref := ctx.PathRef() + "/headers"
		obj.HeadersTypeLink = assemble.NewRefLink[*assemble.Struct](ref)
		ctx.Linker.Add(obj.HeadersTypeLink)
	}
	m.setStructFields(ctx, &obj)
	return &obj, nil
}

func (m Message) setStructFields(ctx *common.CompileContext, langMessage *assemble.Message) {
	fields := []assemble.StructField{
		{
			Name:        "ID",
			Description: "ID is unique string used to identify the message. Case-sensitive.",
			Type:        &assemble.Simple{Type: "string"},
		},
		{Name: "Payload", Type: langMessage.PayloadType},
	}
	if langMessage.HeadersTypeLink != nil {
		lnk := assemble.NewRefLinkAsGolangType(langMessage.HeadersTypeLink.Ref())
		ctx.Linker.Add(lnk)
		fields = append(fields, assemble.StructField{Name: "Headers", Type: lnk})
	} else {
		fields = append(fields, assemble.StructField{Name: "Headers", Type: langMessage.HeadersFallbackType})
	}

	langMessage.OutStruct.Fields = fields
	langMessage.InStruct.Fields = fields
}

func (m Message) getPayloadType(ctx *common.CompileContext) common.GolangType {
	if m.Payload != nil {
		ref := ctx.PathRef() + "/payload"
		lnk := assemble.NewRefLinkAsGolangType(ref)
		ctx.Linker.Add(lnk)
		return lnk
	}
	return &assemble.Simple{Type: "any", IsIface: true}
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
