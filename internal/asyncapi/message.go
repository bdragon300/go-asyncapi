package asyncapi

import (
	"encoding/json"
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	yaml "gopkg.in/yaml.v3"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

type Message struct {
	Headers       *Object                `json:"headers" yaml:"headers" cgen:"data_model"`
	Payload       *Object                `json:"payload" yaml:"payload" cgen:"data_model"` // TODO: other formats
	CorrelationID *CorrelationID         `json:"correlationId" yaml:"correlationId"`
	ContentType   string                 `json:"contentType" yaml:"contentType"`
	Name          string                 `json:"name" yaml:"name"` // TODO: use it for naming
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

func (m Message) Compile(ctx *compile.Context) error {
	obj, err := m.build(ctx, ctx.Stack.Top().Key, ctx.Stack.Top().Flags)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (m Message) build(ctx *compile.Context, messageKey string, flags map[common.SchemaTag]string) (common.Artifact, error) {
	if m.XIgnore {
		ctx.Logger.Debug("Message denoted to be ignored")
		return &render.Message{Dummy: true}, nil
	}

	_, isSelectable := flags[common.SchemaTagSelectable]
	if m.Ref != "" {
		key := messageKey
		path := ctx.Stack.Items()
		// Do not consider the name which a message $ref is registered with inside the channel, keeping the original name in code.
		if path[0].Key != "components" || path[1].Key != "messages" {
			key = ""
		}
		return registerRef(ctx, m.Ref, key, lo.Ternary(isSelectable, lo.ToPtr(true), nil)), nil
	}

	msgName, _ := lo.Coalesce(m.XGoName, messageKey)
	res := render.Message{
		OriginalName: msgName,
		ContentType:  m.ContentType,
		IsSelectable: isSelectable,
		IsPublisher:  ctx.CompileOpts.GeneratePublishers,
		IsSubscriber: ctx.CompileOpts.GenerateSubscribers,
		// map[string]any
		HeadersTypeDefault: &lang.GoMap{
			KeyType:   &lang.GoSimple{TypeName: "string"},
			ValueType: &lang.GoSimple{TypeName: "any", IsInterface: true},
		},
		// any
		PayloadTypeDefault: &lang.GoSimple{TypeName: "any", IsInterface: true},
	}
	ctx.Logger.Trace(fmt.Sprintf("Message content type is %q", res.ContentType))

	// Gather all channels and operations to find out further (after linking) which ones are bound with this message
	prmCh := lang.NewListCbPromise[common.Artifact](func(item common.Artifact) bool {
		path := item.Pointer().Pointer
		if len(path) < 2 || len(path) >= 2 && path[0] != "channels" {
			return false
		}
		return item.Kind() == common.ArtifactKindChannel && item.Visible()
	}, nil)
	res.AllActiveChannelsPromise = prmCh
	ctx.PutListPromise(prmCh)

	prmOp := lang.NewListCbPromise[common.Artifact](func(item common.Artifact) bool {
		path := item.Pointer().Pointer
		if len(path) < 2 || len(path) >= 2 && path[0] != "operations" {
			return false
		}
		return item.Kind() == common.ArtifactKindOperation && item.Visible()
	}, nil)
	res.AllActiveOperationsPromise = prmOp
	ctx.PutListPromise(prmOp)

	prmAsyncAPI := lang.NewCbPromise[*render.AsyncAPI](func(item common.Artifact) bool {
		_, ok := item.(*render.AsyncAPI)
		return ok
	}, nil)
	res.AsyncAPIPromise = prmAsyncAPI
	ctx.PutPromise(prmAsyncAPI)

	if m.Headers != nil {
		ctx.Logger.Trace("Message headers")
		ref := ctx.CurrentPositionRef("headers")
		res.HeadersTypePromise = lang.NewGolangTypePromise(ref, nil)
		res.HeadersTypePromise.AssignErrorNote = "Probably the headers schema has type other than of 'object'?"
		ctx.PutPromise(res.HeadersTypePromise)
	}
	if m.Payload != nil {
		ctx.Logger.Trace("Message payload")
		ref := ctx.CurrentPositionRef("payload")
		res.PayloadTypePromise = lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(res.PayloadTypePromise)
	}
	res.InType, res.OutType = m.buildInOutStructs(ctx, res, msgName)

	// Bindings
	if m.Bindings != nil {
		ctx.Logger.Trace("Message bindings")
		res.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(msgName, "Bindings"),
				HasDefinition: true,
			},
		}

		ref := ctx.CurrentPositionRef("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref, nil)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Link to CorrelationID if any
	if m.CorrelationID != nil {
		ctx.Logger.Trace("Message correlationId")
		ref := ctx.CurrentPositionRef("correlationId")
		res.CorrelationIDPromise = lang.NewPromise[*render.CorrelationID](ref, nil)
		ctx.PutPromise(res.CorrelationIDPromise)
	}

	// Build protocol-specific messages for all supported protocols
	// Here we don't know yet which channels this message is bound with, so we don't have the protocols list to compile.
	ctx.Logger.Trace("Prebuild the messages for every supported protocol")
	var protoMessages []*render.ProtoMessage
	for _, proto := range ctx.SupportedProtocols() {
		ctx.Logger.Trace("Message", "proto", proto)
		protoMessages = append(protoMessages, &render.ProtoMessage{
			Message:  &res,
			Protocol: proto,
		})
	}
	res.ProtoMessages = protoMessages

	return &res, nil
}

func (m Message) buildInOutStructs(ctx *compile.Context, message render.Message, msgName string) (in, out *lang.GoStruct) {
	headerType := message.HeadersTypeDefault
	if message.HeadersTypePromise != nil {
		headerType = message.HeadersTypePromise
	}
	payloadType := message.PayloadTypeDefault
	if message.PayloadTypePromise != nil {
		payloadType = message.PayloadTypePromise
	}
	out = &lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  ctx.GenerateObjName(msgName, "Out"),
			Description:   utils.JoinNonemptyStrings("\n", m.Summary+" (Outbound Message)", m.Description),
			HasDefinition: true,
		},
		Fields: []lang.GoStructField{
			{OriginalName: utils.ToGolangName(string(lang.RuntimeExpressionStructFieldKindPayload), true), Type: payloadType},
			{OriginalName: utils.ToGolangName(string(lang.RuntimeExpressionStructFieldKindHeaders), true), Type: headerType},
		},
	}
	in = &lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  ctx.GenerateObjName(msgName, "In"),
			Description:   utils.JoinNonemptyStrings("\n", m.Summary+" (Inbound Message)", m.Description),
			HasDefinition: true,
		},
		Fields: []lang.GoStructField{
			{OriginalName: utils.ToGolangName(string(lang.RuntimeExpressionStructFieldKindPayload), false), Type: payloadType},
			{OriginalName: utils.ToGolangName(string(lang.RuntimeExpressionStructFieldKindHeaders), false), Type: headerType},
		},
	}

	return
}

type Tag struct {
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description" yaml:"description"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`

	Ref string `json:"$ref" yaml:"$ref"`
}

type MessageExample struct {
	Headers types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"headers" yaml:"headers"`
	Payload *types.Union2[json.RawMessage, yaml.Node]                          `json:"payload" yaml:"payload"`
	Name    string                                                             `json:"name" yaml:"name"`
	Summary string                                                             `json:"summary" yaml:"summary"`
}

type MessageTrait struct {
	Headers       *Object                `json:"headers" yaml:"headers"`
	CorrelationID *CorrelationID         `json:"correlationId" yaml:"correlationId"`
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
