package asyncapi

import (
	"encoding/json"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	yaml "gopkg.in/yaml.v3"

	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

type Message struct {
	Headers       *Object                `json:"headers" yaml:"headers" cgen:"marshal"`
	Payload       *Object                `json:"payload" yaml:"payload" cgen:"marshal"` // TODO: other formats
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
	Traits        []MessageTrait         `json:"traits" yaml:"traits"`

	XGoName string `json:"x-go-name" yaml:"x-go-name"`
	XIgnore bool   `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (m Message) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj, err := m.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (m Message) build(ctx *common.CompileContext, messageKey string) (common.Renderable, error) {
	if m.XIgnore {
		ctx.Logger.Debug("Message denoted to be ignored")
		return &render.Message{Dummy: true}, nil
	}

	if m.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", m.Ref)

		// Message is the only type of objects, that has their own root key, the key in components and can be used
		// as ref in other objects at the same time (at channel.[publish|subscribe].message).
		// Therefore, a message object may get to selections more than once, it's needed to handle it in templates.
		refName := messageKey
		pathStack := ctx.Stack.Items()
		// Ignore the messageKey in definitions other than `messages`, since messageKey always be "message" there.
		if messageKey == "message" && len(pathStack) > 3 {
			refName = ""
		}

		// Always draw the promises that are located in the `messages` section
		return registerRef(ctx, m.Ref, refName, nil), nil
	}

	msgName, _ := lo.Coalesce(m.XGoName, messageKey)
	res := render.Message{
		OriginalName: msgName,
		OutType: &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(msgName, "Out"),
				Description:   utils.JoinNonemptyStrings("\n", m.Summary+" (Outbound Message)", m.Description),
				HasDefinition: true,
			},
		},
		InType: &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(msgName, "In"),
				Description:   utils.JoinNonemptyStrings("\n", m.Summary+" (Inbound Message)", m.Description),
				HasDefinition: true,
			},
		},
		PayloadType:         m.getPayloadType(ctx),
		HeadersFallbackType: &lang.GoMap{KeyType: &lang.GoSimple{TypeName: "string"}, ValueType: &lang.GoSimple{TypeName: "any", IsInterface: true}},
		ContentType:         m.ContentType,
		IsSelectable:        true,
		IsPublisher:         ctx.CompileOpts.GeneratePublishers,
		IsSubscriber:        ctx.CompileOpts.GenerateSubscribers,
	}
	ctx.Logger.Trace(fmt.Sprintf("Message content type is %q", res.ContentType))

	// Gather all channels and operations to find out further (after linking) which ones are bound with this message
	prmCh := lang.NewListCbPromise[common.Renderable](func(item common.CompileObject, path []string) bool {
		if len(path) < 2 || len(path) >= 2 && path[0] != "channels" {
			return false
		}
		return item.Kind() == common.ObjectKindChannel && item.Visible()
	}, nil)
	res.AllActiveChannelsPromise = prmCh
	ctx.PutListPromise(prmCh)

	prmOp := lang.NewListCbPromise[common.Renderable](func(item common.CompileObject, path []string) bool {
		if len(path) < 2 || len(path) >= 2 && path[0] != "operations" {
			return false
		}
		return item.Kind() == common.ObjectKindOperation && item.Visible()
	}, nil)
	res.AllActiveOperationsPromise = prmOp
	ctx.PutListPromise(prmOp)

	prmAsyncAPI := lang.NewCbPromise[*render.AsyncAPI](func(item common.CompileObject, path []string) bool {
		_, ok := item.Renderable.(*render.AsyncAPI)
		return ok
	}, nil)
	res.AsyncAPIPromise = prmAsyncAPI
	ctx.PutPromise(prmAsyncAPI)

	// Link to Headers struct if any
	if m.Headers != nil {
		ctx.Logger.Trace("Message headers")
		ref := ctx.PathStackRef("headers")
		res.HeadersTypePromise = lang.NewPromise[*lang.GoStruct](ref, nil)
		res.HeadersTypePromise.AssignErrorNote = "Probably the headers schema has type other than of 'object'?"
		ctx.PutPromise(res.HeadersTypePromise)
	}
	m.setStructFields(ctx, &res)

	// Bindings
	if m.Bindings != nil {
		ctx.Logger.Trace("Message bindings")
		res.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(msgName, "Bindings"),
				HasDefinition: true,
			},
		}

		ref := ctx.PathStackRef("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref,nil)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Link to CorrelationID if any
	if m.CorrelationID != nil {
		ctx.Logger.Trace("Message correlationId")
		ref := ctx.PathStackRef("correlationId")
		res.CorrelationIDPromise = lang.NewPromise[*render.CorrelationID](ref,nil)
		ctx.PutPromise(res.CorrelationIDPromise)
	}

	// Build protocol-specific messages for all supported protocols
	// Here we don't know yet which channels this message is bound with, so we don't have the protocols list to compile.
	ctx.Logger.Trace("Prebuild the messages for every supported protocol")
	var protoMessages []*render.ProtoMessage
	for proto := range ProtocolBuilders {
		ctx.Logger.Trace("Message", "proto", proto)
		protoMessages = append(protoMessages, &render.ProtoMessage{
			Message:  &res,
			Protocol: proto,
		})
	}
	res.ProtoMessages = protoMessages

	return &res, nil
}

func (m Message) setStructFields(ctx *common.CompileContext, langMessage *render.Message) {
	var headerType common.GolangType
	if langMessage.HeadersTypePromise != nil {
		ctx.Logger.Trace("Message headers has a concrete type")
		prm := lang.NewGolangTypePromise(langMessage.HeadersTypePromise.Ref(), nil)
		ctx.PutPromise(prm)
		headerType = prm
	} else {
		ctx.Logger.Trace("Message headers has `any` type")
		headerType = langMessage.HeadersFallbackType
	}

	langMessage.OutType.Fields = []lang.GoStructField{
		{Name: utils.ToGolangName(string(render.CorrelationIDStructFieldKindPayload), true), Type: langMessage.PayloadType},
		{Name: utils.ToGolangName(string(render.CorrelationIDStructFieldKindHeaders), true), Type: headerType},
	}
	langMessage.InType.Fields = []lang.GoStructField{
		{Name: utils.ToGolangName(string(render.CorrelationIDStructFieldKindPayload), false), Type: langMessage.PayloadType},
		{Name: utils.ToGolangName(string(render.CorrelationIDStructFieldKindHeaders), false), Type: headerType},
	}
}

func (m Message) getPayloadType(ctx *common.CompileContext) common.GolangType {
	if m.Payload != nil {
		ctx.Logger.Trace("Message payload has a concrete type")
		ref := ctx.PathStackRef("payload")
		prm := lang.NewGolangTypePromise(ref, nil)
		ctx.PutPromise(prm)
		return prm
	}

	ctx.Logger.Trace("Message payload has `any` type")
	return &lang.GoSimple{TypeName: "any", IsInterface: true}
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
