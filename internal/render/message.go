package render

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

// Message represents a message object.
type Message struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the message as it was defined in the AsyncAPI document.
	OriginalName string
	// ContentType is the message's content type if set.
	ContentType string

	// Dummy is true when message is ignored (x-ignore: true)
	Dummy bool
	// IsSelectable is true if message should get to selections
	IsSelectable bool
	// IsPublisher is true if the generation of publisher code is enabled
	IsPublisher bool
	// IsSubscriber is true if the generation of subscriber code is enabled
	IsSubscriber bool

	// OutType is a Go struct for the outgoing message. Usually has name like ``MessageOut''
	OutType *lang.GoStruct
	// InType is a Go struct for the incoming message. Usually has name like ``MessageIn''
	InType *lang.GoStruct

	// HeadersTypePromise is a Go struct for message headers. Nil if headers type is not set in document.
	HeadersTypePromise *lang.GolangTypePromise
	// HeadersTypeDefault is a type that is used for headers in message code when headers are not set in the document.
	// Typically, it's ``map[string]string''.
	HeadersTypeDefault common.GolangType
	// PayloadTypePromise is the type of the message payload. Nil if payload type is not set in document.
	PayloadTypePromise *lang.GolangTypePromise
	// PayloadTypeDefault is a type that is used for payload in message code when payload type is not set in the document.
	// Typically, it's ``any''.
	PayloadTypeDefault common.GolangType

	// AllActiveChannelsPromise contains all active channels in the document. Used to find the channels that this message
	// is bound to on the rendering stage.
	AllActiveChannelsPromise *lang.ListPromise[common.Artifact]
	// AllActiveOperationsPromise contains all active operations in the document. Used to find the operations that this
	// message is bound to on the rendering stage.
	AllActiveOperationsPromise *lang.ListPromise[common.Artifact]

	// BindingsType is a Go struct for message bindings. Nil if message bindings are not set.
	BindingsType *lang.GoStruct
	// BindingsPromise is a promise to message bindings contents. Nil if message bindings are not set.
	BindingsPromise *lang.Promise[*Bindings]

	// CorrelationIDPromise is a CorrelationID object defined for the message. Nil if correlationID is not defined.
	CorrelationIDPromise *lang.Promise[*CorrelationID]

	// AsyncAPIPromise is an AsyncAPI root object.
	AsyncAPIPromise *lang.Promise[*AsyncAPI]

	// ProtoMessages is a list of prebuilt ProtoMessage objects for each supported protocol
	ProtoMessages []*ProtoMessage
}

// HeadersType returns a Go type or lang.Ref of headers defined for message in the document.
// If headers is not set, returns the HeadersTypeDefault.
func (m *Message) HeadersType() common.GolangType {
	if m.HeadersTypePromise != nil {
		return m.HeadersTypePromise.T()
	}
	return m.HeadersTypeDefault
}

// PayloadType returns a Go type or lang.Ref of payload defined for message in the document.
// If payload is not set, returns the PayloadTypeDefault.
func (m *Message) PayloadType() common.GolangType {
	if m.PayloadTypePromise != nil {
		return m.PayloadTypePromise.T()
	}
	return m.PayloadTypeDefault
}

// Bindings returns the Bindings object or nil if no bindings are set.
func (m *Message) Bindings() *Bindings {
	if m.BindingsPromise != nil {
		return m.BindingsPromise.T()
	}
	return nil
}

// CorrelationID returns the CorrelationID object or nil if no correlationID is set.
func (m *Message) CorrelationID() *CorrelationID {
	if m.CorrelationIDPromise != nil {
		return m.CorrelationIDPromise.T()
	}
	return nil
}

// AsyncAPI returns the AsyncAPI object.
func (m *Message) AsyncAPI() *AsyncAPI {
	return m.AsyncAPIPromise.T()
}

// EffectiveContentType returns the message's content type for the message if set or the content type of the
// document if set or [DefaultContentType].
func (m *Message) EffectiveContentType() string {
	if m.Dummy {
		return ""
	}
	res, _ := lo.Coalesce(m.ContentType, m.AsyncAPIPromise.T().EffectiveDefaultContentType())
	return res
}

// SelectProtoObject returns a selectable ProtoMessage object for the given protocol or nil if not found or
// the message is not selectable.
func (m *Message) SelectProtoObject(protocol string) *ProtoMessage {
	objects := lo.Filter(m.ProtoMessages, func(p *ProtoMessage, _ int) bool {
		return p.Selectable() && p.Protocol == protocol
	})
	return lo.FirstOr(objects, nil)
}

// BoundChannels returns a list of Channel objects that this message is bound to.
func (m *Message) BoundChannels() []*Channel {
	r := lo.FilterMap(m.AllActiveChannelsPromise.T(), func(c common.Artifact, _ int) (*Channel, bool) {
		ch := common.DerefArtifact(c).(*Channel)
		return ch, lo.ContainsBy(ch.BoundMessages(), func(item *Message) bool {
			return common.CheckSameArtifacts(item, m)
		})
	})
	// ListPromise is filled up by linker, which doesn't guarantee the order. So, sort items by name
	slices.SortFunc(r, func(a, b *Channel) int { return cmp.Compare(a.Name(), b.Name()) })
	return r
}

// BoundOperations returns a list of Operation that this message is bound to.
func (m *Message) BoundOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact(o).(*Operation)
		return op, lo.ContainsBy(op.BoundMessages(), func(item *Message) bool {
			return common.CheckSameArtifacts(item, m)
		})
	})
	// ListPromise is filled up by linker in any order. So, sort items by name to make results stable
	slices.SortFunc(r, func(a, b *Operation) int { return cmp.Compare(a.Name(), b.Name()) })
	return r
}

// BoundAllPubOperations returns a list of Operation that this message is bound to,
// including those where the message is bound via OperationReply, and where the Operation or OperationReply is a publisher.
func (m *Message) BoundAllPubOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact(o).(*Operation)
		return op, op.IsPublisher && lo.Contains(op.BoundMessages(), m) || op.IsReplyPublisher && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// BoundAllSubOperations returns a list of Operation that this message is bound to,
// including those where the message is bound via OperationReply, and where the Operation or OperationReply is a subscriber.
func (m *Message) BoundAllSubOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact(o).(*Operation)
		return op, op.IsSubscriber && lo.Contains(op.BoundMessages(), m) || op.IsReplySubscriber && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// BoundPubReplyOperations returns a list of Operation that this message is bound to via OperationReply only,
// where the OperationReply is for publishing (i.e. Operation is for subscribing).
func (m *Message) BoundPubReplyOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact(o).(*Operation)
		return op, op.IsSubscriber && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// BoundSubReplyOperations returns a list of Operation that this message is bound to via OperationReply only,
// where the OperationReply is for subscribing (i.e. Operation is for publishing).
func (m *Message) BoundSubReplyOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact(o).(*Operation)
		return op, op.IsPublisher && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// BindingsProtocols returns a list of protocols that have bindings defined for this message.
func (m *Message) BindingsProtocols() (res []string) {
	if m.BindingsType == nil {
		return nil
	}
	if m.BindingsPromise != nil {
		res = append(res, m.BindingsPromise.T().Values.Keys()...)
		res = append(res, m.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

// ProtoBindingsValue returns the struct initialization [lang.GoValue] of BindingsType for the given protocol.
// The returned value contains all constant bindings values defined in document for the protocol.
// If no bindings are set for the protocol, returns an empty [lang.GoValue].
func (m *Message) ProtoBindingsValue(protoName string) common.Artifact {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ServerBindings", Import: protoName, IsRuntimeImport: true},
		EmptyCurlyBrackets: true,
	}
	if m.BindingsPromise != nil {
		if b, ok := m.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

func (m *Message) Name() string {
	return m.OriginalName
}

func (m *Message) Kind() common.ArtifactKind {
	return common.ArtifactKindMessage
}

func (m *Message) Selectable() bool {
	return !m.Dummy && m.IsSelectable // Select only the messages defined in the `channels` section`
}

func (m *Message) Visible() bool {
	return !m.Dummy
}

func (m *Message) String() string {
	return "Message(" + m.OriginalName + ")"
}

func (m *Message) Pinnable() bool {
	return true
}

type ProtoMessage struct {
	*Message
	Protocol string
}

func (p *ProtoMessage) Selectable() bool {
	return !p.Dummy && p.isBound()
}

func (p *ProtoMessage) String() string {
	return fmt.Sprintf("ProtoMessage[%s](%s)", p.Protocol, p.OriginalName)
}

// isBound returns true if the message is bound to the protocol
func (p *ProtoMessage) isBound() bool {
	res := lo.SomeBy(p.BoundChannels(), func(c *Channel) bool {
		return !lo.IsNil(c.SelectProtoObject(p.Protocol))
	}) || lo.SomeBy(p.BoundOperations(), func(o *Operation) bool {
		return !lo.IsNil(o.SelectProtoObject(p.Protocol))
	})

	return res
}
