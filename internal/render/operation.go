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
	// AllowTwoWayCode is true if we can generate subscriber code for publisher operation, e.g. OperationReply code.
	// Otherwise, this typically means that user forced to generate only one-way code using --only-pub or --only-sub
	// CLI options.
	AllowTwoWayCode bool

	// ChannelPromise is the channel that this operation is bound with.
	ChannelPromise *lang.Promise[*Channel]

	// BindingsType is a Go struct for operation bindings. Nil if no bindings are set.
	BindingsType *lang.GoStruct
	// BindingsPromise is a promise to operation bindings contents. Nil if no bindings are set.
	BindingsPromise *lang.Promise[*Bindings]

	// MessagesPromises is a list of promises to messages that are bound to this operation.
	MessagesPromises []*lang.Promise[*Message]

	// UseAllChannelMessages is true if operation is bound to all messages in the channel (i.e. when messages field is not set).
	UseAllChannelMessages bool

	// OperationReplyPromise is a promise to the operation reply object. Nil if no reply is set.
	OperationReplyPromise *lang.Promise[*OperationReply]

	// ProtoOperations is a list of prebuilt ProtoOperation objects for each supported protocol
	ProtoOperations []*ProtoOperation
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
func (o *Operation) Messages() []common.Artifact {
	return lo.Map(o.MessagesPromises, func(prm *lang.Promise[*Message], _ int) common.Artifact { return prm.T() })
}

// OperationReply returns the [OperationReply] object or nil if no reply is set.
func (o *Operation) OperationReply() *OperationReply {
	if o.OperationReplyPromise != nil {
		return o.OperationReplyPromise.T()
	}
	return nil
}

// SelectProtoObject returns the ProtoOperation object for the given protocol or nil if not found or if
// ProtoOperation is not selectable.
func (o *Operation) SelectProtoObject(protocol string) common.Artifact {
	res := lo.Filter(o.ProtoOperations, func(p *ProtoOperation, _ int) bool {
		return p.Selectable() && p.Protocol == protocol
	})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}

// BoundMessages returns a list of Message that are bound to this operation. If operation does not define messages, returns
// all messages bound to the operation's channel.
func (o *Operation) BoundMessages() []common.Artifact {
	if o.UseAllChannelMessages {
		return o.Channel().BoundMessages()
	}
	return o.Messages()
}

// BoundReplyMessages returns a list of Message that are bound to this Operation's OperationReply.
// If OperationReply does not define messages, returns all messages bound to the operation's channel.
func (o *Operation) BoundReplyMessages() []common.Artifact {
	// According to AsyncAPI spec, get messages from OperationReply attributes "messages" and "channel"
	// respectively. Otherwise, look to operation's channel.
	if o.OperationReply() == nil {
		return nil
	}
	if len(o.OperationReply().boundMessages()) > 0 {
		return o.OperationReply().boundMessages()
	}
	return o.Channel().BoundMessages()
}

// BoundAllMessages returns a list of Message that are bound to this Operation and its OperationReply.
func (o *Operation) BoundAllMessages() []common.Artifact {
	messages := o.BoundMessages()
	// OperationReply may refer to messages different from those the Operation refers.
	// So, join them in a list and deduplicate.
	messages = append(messages, o.BoundReplyMessages()...)

	r := lo.UniqBy(messages, func(m common.Artifact) string { return m.Pointer().String() })
	return r
}

// BindingsProtocols returns a list of protocols that have bindings defined for this operation.
func (o *Operation) BindingsProtocols() (res []string) {
	if o.BindingsType == nil {
		return nil
	}
	if o.BindingsPromise != nil {
		res = append(res, o.BindingsPromise.T().Values.Keys()...)
		res = append(res, o.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

// ProtoBindingsValue returns the struct initialization [lang.GoValue] of BindingsType for the given protocol.
// The returned value contains all constant bindings values defined in document for the protocol.
// If no bindings are set for the protocol, returns an empty [lang.GoValue].
func (o *Operation) ProtoBindingsValue(protoName string) common.Artifact {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "OperationBindings", Import: protoName, IsRuntimeImport: true},
		EmptyCurlyBrackets: true,
	}
	if o.BindingsPromise != nil {
		if b, ok := o.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

func (o *Operation) Name() string {
	return o.OriginalName
}

func (o *Operation) Kind() common.ArtifactKind {
	return common.ArtifactKindOperation // TODO: separate Bindings from Channel, leaving only the Promise, and make its own 4 ArtifactKinds
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

type ProtoOperation struct {
	*Operation
	// ProtoChannelPromise is ProtoChannel with the same Protocol in the bound Channel.
	ProtoChannelPromise *lang.Promise[*ProtoChannel]
	Protocol            string
}

func (p *ProtoOperation) Selectable() bool {
	return !p.Dummy && p.isBound()
}

func (p *ProtoOperation) String() string {
	return fmt.Sprintf("ProtoOperation[%s](%s)", p.Protocol, p.OriginalName)
}

// ProtoChannel returns the ProtoChannel with the same Protocol in the bound Channel.
func (p *ProtoOperation) ProtoChannel() *ProtoChannel {
	return p.ProtoChannelPromise.T()
}

// isBound returns true if operation is bound to at least one server with supported protocol
func (p *ProtoOperation) isBound() bool {
	protos := lo.Map(p.ChannelPromise.T().BoundServers(), func(s common.Artifact, _ int) string {
		srv := common.DerefArtifact(s).(*Server)
		return srv.Protocol
	})
	r := lo.Contains(protos, p.Protocol)
	return r
}
