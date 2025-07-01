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
	// IsPublisher is true if the generation of publisher code is enabled
	IsPublisher bool
	// IsSubscriber is true if the generation of subscriber code is enabled
	IsSubscriber bool

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

	// ProtoOperations is a list of prebuilt ProtoOperation objects for each supported protocol
	ProtoOperations []*ProtoOperation
}

// Channel returns the Channel that this operation is bound with.
func (o *Operation) Channel() *Channel {
	return o.ChannelPromise.T()
}

// Bindings returns the [Bindings] object or nil if no bindings are set.
func (o *Operation) Bindings() *Bindings {
	return o.BindingsPromise.T()
}

// Messages returns a list of messages defined for this operation. Returns empty list if no messages are set in
// the operation, to get the bound messages use [BoundMessages] method.
func (o *Operation) Messages() []common.Artifact {
	return lo.Map(o.MessagesPromises, func(prm *lang.Promise[*Message], _ int) common.Artifact { return prm.T() })
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

// BoundMessages returns a list of Message that are bound to this operation. Returns all messages in the channel
// if no messages are set in the operation.
func (o *Operation) BoundMessages() []common.Artifact {
	if o.UseAllChannelMessages {
		return o.Channel().Messages()
	}

	return o.Messages()
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
	// Type is a protocol-specific operation's Go struct
	Type     *lang.GoStruct
	Protocol string
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
