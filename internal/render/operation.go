package render

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

// Operation represents an operation object.
type Operation struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the operation as it was defined in the AsyncAPI document.
	OriginalName string
	// Dummy is true when operation is ignored (x-ignore: true)
	Dummy bool

	// IsSelectable is true if operation should get to selections
	IsSelectable bool
	// IsPublisher is true if this is a publishing operation and the generation of publisher code is enabled
	IsPublisher bool
	// IsSubscriber is true if this is a subscribing operation and the generation of subscriber code is enabled
	IsSubscriber bool
	// IsReplyPublisher is true if this operation should generate the publishing code for its OperationReply.
	IsReplyPublisher bool
	// IsReplySubscriber is true if this operation should generate the subscribing code for its OperationReply.
	IsReplySubscriber bool

	// ChannelPromise is the channel that this operation is bound with.
	ChannelPromise *lang.Promise[*Channel]

	// BindingsPromise is a promise to operation bindings contents. Nil if no bindings are set.
	BindingsPromise *lang.Promise[*Bindings]

	// MessagesPromises is a list of promises to messages that are bound to this operation.
	MessagesPromises []*lang.Promise[*Message]

	// UseAllChannelMessages is true if operation is bound to all messages in the channel (i.e. when messages field is not set).
	UseAllChannelMessages bool

	// OperationReplyPromise is a promise to the operation reply object. Nil if no reply is set.
	OperationReplyPromise *lang.Promise[*OperationReply]

	// SecuritySchemePromises is a promises to the security scheme objects defined for this operation.
	SecuritySchemePromises []*lang.Promise[*SecurityScheme]
}

// Channel returns the Channel that this operation is bound with.
func (o *Operation) Channel() *Channel {
	return o.ChannelPromise.T()
}

// Bindings returns the [Bindings] object or nil if no bindings are set.
func (o *Operation) Bindings() *Bindings {
	if o.BindingsPromise != nil {
		return o.BindingsPromise.T()
	}
	return nil
}

// Messages returns a list of messages defined for this operation. Returns empty list if no messages are set in
// the operation, to get the bound messages use [boundMessages] method.
func (o *Operation) Messages() []*Message {
	return lo.Map(o.MessagesPromises, func(prm *lang.Promise[*Message], _ int) *Message { return prm.T() })
}

// OperationReply returns the [OperationReply] object or nil if no reply is set.
func (o *Operation) OperationReply() *OperationReply {
	if o.OperationReplyPromise != nil {
		return o.OperationReplyPromise.T()
	}
	return nil
}

// BoundOperationReplyChannel returns a Channel bound to Operation's OperationReply.
// If OperationReply is not set, returns nil.
func (o *Operation) BoundOperationReplyChannel() *Channel {
	if o.OperationReply() != nil {
		c := o.OperationReply().Channel()
		if c != nil {
			return c
		}
		return o.Channel()
	}
	return nil
}

// ProtoOperation returns the ProtoOperation object for the given protocol.
func (o *Operation) ProtoOperation(protocol string) *ProtoOperation {
	return &ProtoOperation{Operation: o, Protocol: protocol}
}

// BoundMessages returns a list of Message that are bound to this operation. If operation does not define messages, returns
// all messages bound to the operation's channel.
func (o *Operation) BoundMessages() []*Message {
	if o.UseAllChannelMessages {
		return o.Channel().BoundMessages()
	}
	return o.Messages()
}

// BoundReplyMessages returns a list of Message that are bound to this Operation's OperationReply.
// If OperationReply does not specify any messages, returns all messages bound to its channel. If it's empty,
// return messages bound to the operation. If OperationReply is not set, returns nil.
func (o *Operation) BoundReplyMessages() []*Message {
	// According to AsyncAPI spec, get messages from "messages", otherwise from "channel"
	// respectively. Otherwise, look to operation's messages or channel.
	if o.OperationReply() == nil {
		return nil
	}
	if len(o.OperationReply().boundMessages()) > 0 {
		return o.OperationReply().boundMessages()
	}
	return o.BoundMessages()
}

// ActiveProtocols returns a unique list of protocols that are bound to this operation via its channel and
// via its OperationReply's channel (if any).
func (o *Operation) ActiveProtocols() []string {
	res := o.Channel().ActiveProtocols()
	if o.OperationReply() != nil {
		c := o.OperationReply().Channel()
		if c != nil && c.Selectable() && c.Visible() {
			res = append(res, c.ActiveProtocols()...)
		}
	}
	return lo.Uniq(res)
}

// BindingsProtocols returns a list of protocols that have bindings defined for this operation.
func (o *Operation) BindingsProtocols() []string {
	if o.Bindings() != nil {
		return lo.Uniq(o.Bindings().Protocols())
	}
	return nil
}

// SecuritySchemes returns the list of security schemes defined for this operation.
func (o *Operation) SecuritySchemes() []*SecurityScheme {
	return lo.Map(o.SecuritySchemePromises, func(item *lang.Promise[*SecurityScheme], _ int) *SecurityScheme {
		return item.T()
	})
}

func (o *Operation) HasPublishingCode() bool {
	return o.IsPublisher || o.IsReplyPublisher
}

func (o *Operation) HasSubscribingCode() bool {
	return o.IsSubscriber || o.IsReplySubscriber
}

func (o *Operation) Name() string {
	return o.OriginalName
}

func (o *Operation) Kind() common.ArtifactKind {
	return common.ArtifactKindOperation
}

func (o *Operation) Selectable() bool {
	// Proto channels for each supported protocol
	// If bound channel is not selectable, then operation is not selectable as well
	return !o.Dummy && o.ChannelPromise.T().Selectable() && o.IsSelectable && (o.IsPublisher || o.IsSubscriber)
}

func (o *Operation) Visible() bool {
	return !o.Dummy && o.ChannelPromise.T().Visible() && (o.IsPublisher || o.IsSubscriber)
}

func (o *Operation) String() string {
	return "Operation(" + o.OriginalName + ")"
}

func (o *Operation) Pinnable() bool {
	return true
}

type ProtoOperation struct {
	*Operation
	Protocol string
}

func (p *ProtoOperation) Pinnable() bool {
	return false
}

func (p *ProtoOperation) String() string {
	return fmt.Sprintf("ProtoOperation[%s](%s)", p.Protocol, p.OriginalName)
}

// ProtoChannel returns the ProtoChannel with the same Protocol in the bound Channel.
func (p *ProtoOperation) ProtoChannel() *ProtoChannel {
	return p.Channel().ProtoChannel(p.Protocol)
}
