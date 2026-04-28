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
	// OriginalContentType is the original message's content type defined in document without considering
	// the global document content type or MessageTrait values. To get the resulting content type, use EffectiveContentType method.
	OriginalContentType string

	// Dummy is true when message is ignored (x-ignore: true)
	Dummy bool
	// IsSelectable is true if message should get to selections
	IsSelectable bool
	// IsPublisher is true if the generation of publisher code is enabled
	IsPublisher bool
	// IsSubscriber is true if the generation of subscriber code is enabled
	IsSubscriber bool

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

	// BindingsPromise is a promise to message bindings contents. Nil if message bindings are not set.
	BindingsPromise *lang.Promise[*Bindings]

	// CorrelationIDPromise is a CorrelationID object defined for the message. Nil if correlationID is not defined.
	CorrelationIDPromise *lang.Promise[*CorrelationID]

	// MessageTraitPromises is a list of promises to MessageTrait objects that are applied to this message. Nil if no message traits are applied.
	MessageTraitPromises []*lang.Promise[*MessageTrait]

	// AsyncAPIPromise is an AsyncAPI root object. Always non-empty.
	AsyncAPIPromise *lang.Promise[*AsyncAPI]
}

// HeadersType returns a Go type of headers defined for message in the document.
// If headers is not set, returns the HeadersTypeDefault.
func (m *Message) HeadersType() common.GolangType {
	if m.HeadersTypePromise != nil {
		return common.DerefArtifact[common.GolangType](m.HeadersTypePromise.T())
	}
	return lo.CoalesceOrEmpty(
		common.LastNonEmptyTargetBy(m.MessageTraitPromises, func(mt *MessageTrait) common.GolangType { return mt.HeadersType() }),
		m.HeadersTypeDefault,
	)
}

// HasHeaders returns true if headers are defined for this message either directly or via message traits.
func (m *Message) HasHeaders() bool {
	return m.HeadersTypePromise != nil || common.LastNonEmptyTargetBy(m.MessageTraitPromises, func(mt *MessageTrait) bool { return mt.HasHeaders() })
}

// PayloadType returns a Go type of payload defined for message in the document.
// If payload is not set, returns the PayloadTypeDefault.
func (m *Message) PayloadType() common.GolangType {
	if m.PayloadTypePromise != nil {
		return common.DerefArtifact[common.GolangType](m.PayloadTypePromise.T())
	}
	return m.PayloadTypeDefault
}

// Bindings returns the Bindings object or nil if no bindings are set.
func (m *Message) Bindings() *Bindings {
	if m.BindingsPromise != nil {
		return m.BindingsPromise.T()
	}
	return common.LastNonEmptyTargetBy(m.MessageTraitPromises, func(mt *MessageTrait) *Bindings { return mt.Bindings() })
}

// CorrelationID returns the CorrelationID object or nil if no correlationID is set.
func (m *Message) CorrelationID() *CorrelationID {
	if m.CorrelationIDPromise != nil {
		return m.CorrelationIDPromise.T()
	}
	return common.LastNonEmptyTargetBy(m.MessageTraitPromises, func(mt *MessageTrait) *CorrelationID { return mt.CorrelationID() })
}

// AsyncAPI returns the AsyncAPI object.
func (m *Message) AsyncAPI() *AsyncAPI {
	return m.AsyncAPIPromise.T()
}

// EffectiveContentType returns the resulting message's content effectively merged from global document's default and
// [MessageTrait] values. If none of them is set, returns [DefaultContentType].
func (m *Message) EffectiveContentType() string {
	if m.Dummy {
		return ""
	}
	res := lo.CoalesceOrEmpty(
		m.OriginalContentType,
		common.LastNonEmptyTargetBy(m.MessageTraitPromises, func(mt *MessageTrait) string { return mt.ContentType }),
		m.AsyncAPIPromise.T().EffectiveDefaultContentType(),
	)
	return res
}

// ProtoMessage returns a selectable ProtoMessage object for the given protocol.
func (m *Message) ProtoMessage(protocol string) *ProtoMessage {
	return &ProtoMessage{Message: m, Protocol: protocol}
}

// BoundChannels returns a list of Channel objects that this message is bound to.
func (m *Message) BoundChannels() []*Channel {
	r := lo.FilterMap(m.AllActiveChannelsPromise.T(), func(c common.Artifact, _ int) (*Channel, bool) {
		ch := common.DerefArtifact[*Channel](c)
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
		op := common.DerefArtifact[*Operation](o)
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
		op := common.DerefArtifact[*Operation](o)
		return op, op.IsPublisher && lo.Contains(op.BoundMessages(), m) || op.IsReplyPublisher && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// BoundAllSubOperations returns a list of Operation that this message is bound to,
// including those where the message is bound via OperationReply, and where the Operation or OperationReply is a subscriber.
func (m *Message) BoundAllSubOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact[*Operation](o)
		return op, op.IsSubscriber && lo.Contains(op.BoundMessages(), m) || op.IsReplySubscriber && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// BoundPubReplyOperations returns a list of Operation that this message is bound to via OperationReply only,
// where the OperationReply is for publishing (i.e. Operation is for subscribing).
func (m *Message) BoundPubReplyOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact[*Operation](o)
		return op, op.IsSubscriber && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// BoundSubReplyOperations returns a list of Operation that this message is bound to via OperationReply only,
// where the OperationReply is for subscribing (i.e. Operation is for publishing).
func (m *Message) BoundSubReplyOperations() []*Operation {
	r := lo.FilterMap(m.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact[*Operation](o)
		return op, op.IsPublisher && lo.Contains(op.BoundReplyMessages(), m)
	})
	return r
}

// ActiveProtocols returns a unique list of protocols that this message is bound to through channels and operations.
func (m *Message) ActiveProtocols() []string {
	protocols := lo.FilterMap(m.BoundChannels(), func(c *Channel, _ int) ([]string, bool) {
		return c.ActiveProtocols(), c.Selectable() && c.Visible()
	})
	protocols = append(protocols, lo.FilterMap(m.BoundOperations(), func(o *Operation, _ int) ([]string, bool) {
		return o.Channel().ActiveProtocols(), o.Selectable() && o.Visible()
	})...)
	return lo.Uniq(lo.Flatten(protocols))
}

// BindingsProtocols returns a list of protocols that have bindings defined for this message.
func (m *Message) BindingsProtocols() []string {
	if m.Bindings() != nil {
		return lo.Uniq(m.Bindings().Protocols())
	}
	return nil
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

func (p *ProtoMessage) Pinnable() bool {
	return false // ProtoMessage is a temporary object, it cannot be referenced from code, so cannot be pinned
}

func (p *ProtoMessage) String() string {
	return fmt.Sprintf("ProtoMessage[%s](%s)", p.Protocol, p.OriginalName)
}
