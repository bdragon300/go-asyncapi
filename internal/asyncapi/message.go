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
	obj, err := m.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (m Message) build(ctx *common.CompileContext, messageKey string) (common.Renderable, error) {
	_, isComponent := ctx.Stack.Top().Flags[common.SchemaTagComponent]
	ignore := m.XIgnore //|| (isComponent && !ctx.CompileOpts.MessageOpts.IsAllowedName(messageKey))
	if ignore {
		ctx.Logger.Debug("Message denoted to be ignored")
		return &render.Message{Dummy: true}, nil
	}

	if m.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", m.Ref)

		// Message is the only type of objects, that has their own root key, the key in components and can be used
		// as ref in other objects at the same time (at channel.[publish|subscribe].message).
		// Therefore, a message object may get to selections more than once, it's needed to handle in templates.
		refName := messageKey
		pathStack := ctx.Stack.Items()
		makeSelectable := !isComponent
		// Ignore the messageKey in definitions other than `messages`, since messageKey always be "message" there.
		if messageKey == "message" && len(pathStack) > 3 {
			// Use the OriginalName of message instead.
			refName = ""
			// And force make the message selectable if it was defined in `components.messages` section.
			// TODO: force make selectable only if the channel it refers to is also visible
			makeSelectable = true
		}

		// Always draw the promises that are located in the `messages` section
		prm := lang.NewRef(m.Ref, refName, lo.Ternary(makeSelectable, lo.ToPtr(true), nil))
		ctx.PutPromise(prm)
		return prm, nil
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
		ContentType: m.ContentType,
		IsComponent: isComponent,
	}
	ctx.Logger.Trace(fmt.Sprintf("Message content type is %q", res.ContentType))

	// Lookup servers after linking to figure out all protocols the message is used in
	prm := lang.NewListCbPromise[*render.Server](func(item common.CompileObject, path []string) bool {
		_, ok := item.Renderable.(*render.Server)
		return ok
	})
	res.AllServersPromise = prm
	ctx.PutListPromise(prm)

	prm2 := lang.NewCbPromise[*render.AsyncAPI](func(item common.CompileObject, path []string) bool {
		_, ok := item.Renderable.(*render.AsyncAPI)
		return ok
	}, nil)
	res.AsyncAPIPromise = prm2
	ctx.PutPromise(prm2)

	// Link to Headers struct if any
	if m.Headers != nil {
		ctx.Logger.Trace("Message headers")
		ref := ctx.PathStackRef("headers")
		res.HeadersTypePromise = lang.NewPromise[*lang.GoStruct](ref)
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
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Link to CorrelationID if any
	if m.CorrelationID != nil {
		ctx.Logger.Trace("Message correlationId")
		ref := ctx.PathStackRef("correlationId")
		res.CorrelationIDPromise = lang.NewPromise[*render.CorrelationID](ref)
		ctx.PutPromise(res.CorrelationIDPromise)
	}

	// Build protocol-specific messages for all supported protocols
	// Here we don't know yet which channels this message is used by, so we don't have the protocols list to compile.
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
	fields := []lang.GoStructField{
		{Name: "Payload", Type: langMessage.PayloadType},
	}
	if langMessage.HeadersTypePromise != nil {
		ctx.Logger.Trace("Message headers has a concrete type")
		prm := lang.NewGolangTypePromise(langMessage.HeadersTypePromise.Ref())
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
		prm := lang.NewGolangTypePromise(ref)
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
