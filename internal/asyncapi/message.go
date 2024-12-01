package asyncapi

import (
	"encoding/json"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/specurl"
	"github.com/bdragon300/go-asyncapi/internal/types"

	yaml "gopkg.in/yaml.v3"

	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

type Message struct {
	MessageID     string                 `json:"messageId" yaml:"messageId"`
	Headers       *Object                `json:"headers" yaml:"headers" cgen:"marshal"`
	Payload       *Object                `json:"payload" yaml:"payload" cgen:"marshal"` // TODO: other formats
	CorrelationID *CorrelationID         `json:"correlationId" yaml:"correlationId"`
	SchemaFormat  string                 `json:"schemaFormat" yaml:"schemaFormat"`
	ContentType   string                 `json:"contentType" yaml:"contentType"`
	Name          string                 `json:"name" yaml:"name"`
	Title         string                 `json:"title" yaml:"title"`
	Summary       string                 `json:"summary" yaml:"summary"`
	Description   string                 `json:"description" yaml:"description"`
	Tags          []Tag                  `json:"tags" yaml:"tags"`
	ExternalDocs  *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
	Bindings      *MessageBindings       `json:"bindings" yaml:"bindings"`
	Examples      []MessageExample       `json:"examples" yaml:"examples"`
	Traits        []MessageTrait         `json:"traits" yaml:"traits"`

	XGoName string `json:"x-go-name" yaml:"x-go-name"`
	XIgnore bool   `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (m Message) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	msgs, err := m.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	for _, msg := range msgs {
		ctx.PutObject(msg)
	}
	return nil
}

func (m Message) build(ctx *common.CompileContext, messageKey string) ([]common.Renderer, error) {
	var res []common.Renderer

	_, isComponent := ctx.Stack.Top().Flags[common.SchemaTagComponent]
	ignore := m.XIgnore || (isComponent && !ctx.CompileOpts.MessageOpts.IsAllowedName(messageKey))
	if ignore {
		ctx.Logger.Debug("Message denoted to be ignored")
		res = append(res, &render.ProtoMessage{Message: &render.Message{Dummy: true}})
		return res, nil
	}
	if m.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", m.Ref)
		prm := lang.NewRendererPromise(m.Ref, common.PromiseOriginUser)
		ctx.PutPromise(prm)
		res = append(res, prm)
		return res, nil
	}

	msgName := messageKey
	// If the message is not a component, but inlined in a channel (i.e. in channels package), the messageKey always
	// will be "message". So, we need to generate a unique name for the message, considering if it's
	// publish/subscribe message, because we don't generate a separate code object for a channel operation,
	// therefore a channel can have two identical messages.
	if ctx.CurrentPackage() == PackageScopeChannels {
		msgName = ctx.GenerateObjName(m.XGoName, "")
	}
	msgName, _ = lo.Coalesce(m.XGoName, msgName)

	baseMessage := render.Message{
		Name: msgName,
		OutType: &lang.GoStruct{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(msgName, "Out"),
				Description:   utils.JoinNonemptyStrings("\n", m.Summary+" (Outbound Message)", m.Description),
				HasDefinition: true,
			},
		},
		InType: &lang.GoStruct{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(msgName, "In"),
				Description:   utils.JoinNonemptyStrings("\n", m.Summary+" (Inbound Message)", m.Description),
				HasDefinition: true,
			},
		},
		PayloadType:         m.getPayloadType(ctx),
		HeadersFallbackType: &lang.GoMap{KeyType: &lang.GoSimple{Name: "string"}, ValueType: &lang.GoSimple{Name: "any", IsInterface: true}},
		ContentType: m.ContentType,
	}
	ctx.Logger.Trace(fmt.Sprintf("Message content type is %q", baseMessage.ContentType))

	// Lookup servers after linking to figure out all protocols the message is used in
	prm := lang.NewListCbPromise[*render.Server](func(item common.Renderer, path []string) bool {
		_, ok := item.(*render.Server)
		return ok
	})
	baseMessage.AllServersPromise = prm

	// Link to Headers struct if any
	if m.Headers != nil {
		ctx.Logger.Trace("Message headers")
		ref := ctx.PathStackRef("headers")
		baseMessage.HeadersTypePromise = lang.NewPromise[*lang.GoStruct](ref, common.PromiseOriginInternal)
		baseMessage.HeadersTypePromise.AssignErrorNote = "Probably the headers schema has type other than of 'object'?"
		ctx.PutPromise(baseMessage.HeadersTypePromise)
	}
	m.setStructFields(ctx, &baseMessage)

	// Bindings
	if m.Bindings != nil {
		ctx.Logger.Trace("Message bindings")
		baseMessage.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(msgName, "Bindings"),
				HasDefinition: true,
			},
			Fields: nil,
		}

		ref := ctx.PathStackRef("bindings")
		baseMessage.BindingsPromise = lang.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(baseMessage.BindingsPromise)
	}

	// Link to CorrelationID if any
	if m.CorrelationID != nil {
		ctx.Logger.Trace("Message correlationId")
		ref := ctx.PathStackRef("correlationId")
		baseMessage.CorrelationIDPromise = lang.NewPromise[*render.CorrelationID](ref, common.PromiseOriginInternal)
		ctx.PutPromise(baseMessage.CorrelationIDPromise)
	}

	// Build protocol-specific messages for all supported protocols
	// Here we don't know yet which channels this message is used by, so we don't have the protocols list to compile.
	ctx.Logger.Trace("Prebuild the messages for every supported protocol")
	for proto := range ProtocolBuilders {
		ctx.Logger.Trace("Message", "proto", proto)
		res = append(res, &render.ProtoMessage{
			Message:   &baseMessage,
			ProtoName: proto,
		})
	}

	return res, nil
}

func (m Message) setStructFields(ctx *common.CompileContext, langMessage *render.Message) {
	fields := []lang.GoStructField{
		{Name: "Payload", Type: langMessage.PayloadType},
	}
	if langMessage.HeadersTypePromise != nil {
		ctx.Logger.Trace("Message headers has a concrete type")
		prm := lang.NewGolangTypePromise(langMessage.HeadersTypePromise.Ref(), common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		fields = append(fields, lang.GoStructField{Name: "Headers", Type: prm})
	} else {
		ctx.Logger.Trace("Message headers has `any` type")
		fields = append(fields, lang.GoStructField{Name: "Headers", Type: langMessage.HeadersFallbackType})
	}

	langMessage.OutType.Fields = fields
	langMessage.InType.Fields = fields
}

func (m Message) getPayloadType(ctx *common.CompileContext) common.GolangType {
	if m.Payload != nil {
		ctx.Logger.Trace("Message payload has a concrete type")
		ref := ctx.PathStackRef("payload")
		prm := lang.NewGolangTypePromise(ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		return prm
	}

	ctx.Logger.Trace("Message payload has `any` type")
	return &lang.GoSimple{Name: "any", IsInterface: true}
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
	MessageID     string                 `json:"messageId" yaml:"messageId"`
	Headers       *Object                `json:"headers" yaml:"headers"`
	CorrelationID *CorrelationID         `json:"correlationId" yaml:"correlationId"`
	SchemaFormat  string                 `json:"schemaFormat" yaml:"schemaFormat"`
	ContentType   string                 `json:"contentType" yaml:"contentType"`
	Name          string                 `json:"name" yaml:"name"`
	Title         string                 `json:"title" yaml:"title"`
	Summary       string                 `json:"summary" yaml:"summary"`
	Description   string                 `json:"description" yaml:"description"`
	Tags          []Tag                  `json:"tags" yaml:"tags"`
	ExternalDocs  *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
	Bindings      *MessageBindings       `json:"bindings" yaml:"bindings"`
	Examples      []MessageExample       `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
