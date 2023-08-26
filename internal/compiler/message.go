package compiler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"gopkg.in/yaml.v3"
)

type Message struct {
	MessageID     string                        `json:"messageId" yaml:"messageId"`
	Headers       *Object                       `json:"headers" yaml:"headers"`
	Payload       *Object                       `json:"payload" yaml:"payload"` // TODO: other formats
	CorrelationID *CorrelationID                `json:"correlationId" yaml:"correlationId"`
	SchemaFormat  string                        `json:"schemaFormat" yaml:"schemaFormat"`
	ContentType   string                        `json:"contentType" yaml:"contentType"`
	Name          string                        `json:"name" yaml:"name"`
	Title         string                        `json:"title" yaml:"title"`
	Summary       string                        `json:"summary" yaml:"summary"`
	Description   string                        `json:"description" yaml:"description"`
	Tags          []Tag                         `json:"tags" yaml:"tags"`
	ExternalDocs  *ExternalDocumentation        `json:"externalDocs" yaml:"externalDocs"`
	Bindings      utils.OrderedMap[string, any] `json:"bindings" yaml:"bindings"` // TODO: replace any to common bindings object
	Examples      []MessageExample              `json:"examples" yaml:"examples"`
	Traits        []MessageTrait                `json:"traits" yaml:"traits"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (m Message) Build(ctx *scan.Context) error {
	obj, err := m.buildMessage(ctx, ctx.Top().Path)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.CurrentPackage().Put(ctx, obj)
	return nil
}

func (m Message) buildMessage(ctx *scan.Context, name string) (render.LangRenderer, error) {
	if m.ContentType != "" && m.ContentType != "application/json" {
		return nil, fmt.Errorf("now is supported only application/json") // TODO: support other content types
	}
	if m.Ref != "" {
		res := lang.NewLinkerQueryRendererRef(common.MessagesPackageKind, m.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	typeName := getTypeName(ctx, name, "Message")
	strct := lang.Struct{
		BaseType: lang.BaseType{
			Name:        typeName,
			Description: utils.JoinNonemptyStrings("\n", m.Summary, m.Description),
			Render:      true,
		},
	}

	obj := lang.Message{
		Struct:           &strct,
		PayloadType:      m.getPayloadType(ctx),
		PayloadHasSchema: m.Payload != nil,
		HeadersType:      m.getHeadersType(ctx),
		HeadersHasSchema: m.Headers != nil,
	}
	m.setStructFields(&obj)
	return &obj, nil
}

func (m Message) setStructFields(langMessage *lang.Message) {
	langMessage.Struct.Fields = []lang.StructField{
		{
			Name:        "ID",
			Description: "ID is unique string used to identify the message. Case-sensitive.",
			Type:        &lang.Simple{TypeName: "string"},
		},
		{Name: "Payload", Type: langMessage.PayloadType},
		{Name: "Headers", Type: langMessage.HeadersType},
	}
}

func (m Message) getPayloadType(ctx *scan.Context) lang.LangType {
	if m.Payload != nil {
		path := append(ctx.PathStack(), "payload")
		q := lang.NewLinkerQueryTypePath(ctx.Top().PackageKind, path)
		ctx.Linker.Add(q)
		return q
	}
	return &lang.Simple{TypeName: "any"}
}

func (m Message) getHeadersType(ctx *scan.Context) lang.LangType {
	if m.Headers != nil {
		path := append(ctx.PathStack(), "headers")
		q := lang.NewLinkerQueryTypePath(ctx.Top().PackageKind, path)
		ctx.Linker.Add(q)
		return q
	}
	return &lang.Map{KeyType: &lang.Simple{TypeName: "string"}, ValueType: &lang.Simple{TypeName: "any"}}
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
	MessageID     string                        `json:"messageId" yaml:"messageId"`
	Headers       *Object                       `json:"headers" yaml:"headers"`
	CorrelationID *CorrelationID                `json:"correlationId" yaml:"correlationId"`
	SchemaFormat  string                        `json:"schemaFormat" yaml:"schemaFormat"`
	ContentType   string                        `json:"contentType" yaml:"contentType"`
	Name          string                        `json:"name" yaml:"name"`
	Title         string                        `json:"title" yaml:"title"`
	Summary       string                        `json:"summary" yaml:"summary"`
	Description   string                        `json:"description" yaml:"description"`
	Tags          []Tag                         `json:"tags" yaml:"tags"`
	ExternalDocs  *ExternalDocumentation        `json:"externalDocs" yaml:"externalDocs"`
	Bindings      utils.OrderedMap[string, any] `json:"bindings" yaml:"bindings"` // FIXME: replace any to common bindings object
	Examples      []MessageExample              `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
