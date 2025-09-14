package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

type OperationReply struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the operation reply as it was defined in the AsyncAPI document.
	OriginalName string

	// OperationReplyAddressPromise is a promise to the address of the operation reply. Nil if no address is set.
	OperationReplyAddressPromise *lang.Promise[*OperationReplyAddress]

	// ChannelPromise is the channel that this operation reply is bound with. Nil if no channel is set.
	ChannelPromise *lang.Promise[*Channel]

	// MessagesPromises is a list of promises to messages that are bound to this operation reply.
	MessagesPromises []*lang.Promise[*Message]
	// UseAllChannelMessages is true if operation reply is bound to all messages in the channel (i.e. when messages field is not set).
	UseAllChannelMessages bool

	Dummy bool
}

// Channel returns the Channel that this operation reply is bound with. Can be nil if no channel is set.
func (o *OperationReply) Channel() *Channel {
	if o.ChannelPromise != nil {
		return o.ChannelPromise.T()
	}
	return nil
}

// Messages returns a list of messages defined for this operation reply. Returns empty list if no messages are set.
func (o *OperationReply) Messages() []*Message {
	return lo.Map(o.MessagesPromises, func(prm *lang.Promise[*Message], _ int) *Message { return prm.T() })
}

// OperationReplyAddress returns the address object of the operation reply or nil if no address is set.
func (o *OperationReply) OperationReplyAddress() *OperationReplyAddress {
	if o.OperationReplyAddressPromise != nil {
		return o.OperationReplyAddressPromise.T()
	}
	return nil
}

// boundMessages returns the list of messages that are bound to this operation reply. If "messages" field is not set,
// it returns all messages bound to the channel.
//
// If both fields aren't set, the bound messages are equal to the messages list of operation's channel.
// This function doesn't return it, use [Operation] methods instead.
func (o *OperationReply) boundMessages() []*Message {
	if o.UseAllChannelMessages && o.Channel() != nil {
		return o.Channel().BoundMessages()
	}
	return o.Messages()
}

func (o *OperationReply) Name() string {
	return o.OriginalName
}

func (o *OperationReply) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (o *OperationReply) Selectable() bool {
	return !o.Dummy
}

func (o *OperationReply) Visible() bool {
	return !o.Dummy && (o.ChannelPromise == nil || o.ChannelPromise.T().Visible())
}

func (o *OperationReply) String() string {
	return "OperationReply(" + o.OriginalName + ")"
}
