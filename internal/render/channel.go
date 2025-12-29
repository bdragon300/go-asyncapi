package render

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

// Channel represents the channel object.
type Channel struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the channel as it was defined in the AsyncAPI document.
	OriginalName string
	// Address is channel address raw value.
	Address string

	// Dummy is true when channel is ignored (x-ignore: true)
	Dummy bool
	// IsSelectable is true if channel should get to selections
	IsSelectable bool
	// IsPublisher is true if the generation of publisher code is enabled
	IsPublisher bool
	// IsSubscriber is true if the generation of subscriber code is enabled
	IsSubscriber bool

	// ServersPromises is a list of servers that this channel is bound with. Empty if no servers are set.
	ServersPromises []*lang.Promise[*Server]
	// AllActiveServersPromise contains all active servers in the document. It is used when servers field is not set
	// for the channel, which means that the channel is bound to all active servers.
	AllActiveServersPromise *lang.ListPromise[common.Artifact]

	// ParametersRefs is a list of refs to channel parameters.
	ParametersRefs []*lang.Ref
	// ParametersType is a Go struct for channel parameters. Nil if no parameters are set.
	ParametersType *lang.GoStruct

	// MessagesRefs is a list of references to messages that this channel is bound with, both for ones referenced
	// by $ref or inlined definitions.
	MessagesRefs []*lang.Ref

	// AllActiveOperationsPromise contains all active operations in the document.
	//
	// On compiling stage we don't know which operations are bound to a particular channel.
	// So we just collect all operations to this field and postpone filtering them until the rendering stage.
	//
	// We could use a promise callback to filter operations by channel, but the channel in operation is also a promise,
	// and the order of promises resolving is not guaranteed.
	AllActiveOperationsPromise *lang.ListPromise[common.Artifact]

	// BindingsType is a Go struct for channel bindings. Nil if channel bindings are not set at all.
	BindingsType *lang.GoStruct
	// BindingsPromise is a promise to channel bindings contents. Nil if no bindings are set.
	BindingsPromise *lang.Promise[*Bindings]
}

// Parameters returns a map of channel's Parameter objects by names which they defined in channel.
//
// It may return lang.Ref objects, because a key the parameter object is defined with in channel matters and
// is used in channel address template, therefore, it should appear in the generated code as well.
func (c *Channel) Parameters() (res types.OrderedMap[string, *Parameter]) {
	for _, ref := range c.ParametersRefs {
		res.Set(ref.Name(), common.DerefArtifact(ref).(*Parameter))
	}
	return
}

// Messages returns a list of Message.
func (c *Channel) Messages() []*Message {
	return lo.Map(c.MessagesRefs, func(prm *lang.Ref, _ int) *Message { return common.DerefArtifact(prm).(*Message) })
}

// Bindings returns the [Bindings] object or nil if no bindings are set.
func (c *Channel) Bindings() *Bindings {
	if c.BindingsPromise != nil {
		return c.BindingsPromise.T()
	}
	return nil
}

// ProtoChannel returns a selectable ProtoChannel object for the given protocol.
func (c *Channel) ProtoChannel(protocol string) *ProtoChannel {
	return &ProtoChannel{Channel: c, Protocol: protocol}
}

// BoundServers returns a list of Server objects that this channel is bound with.
func (c *Channel) BoundServers() []*Server {
	if c.Dummy {
		return nil
	}

	var res []*Server
	if len(c.ServersPromises) == 0 {
		res = lo.Map(c.AllActiveServersPromise.T(), func(a common.Artifact, _ int) *Server { return common.DerefArtifact(a).(*Server) })
		// ListPromise is filled up by linker, which doesn't guarantee the order. So, sort items by name
		slices.SortFunc(res, func(a, b *Server) int { return cmp.Compare(a.Name(), b.Name()) })
	} else {
		res = lo.Map(c.ServersPromises, func(s *lang.Promise[*Server], _ int) *Server { return s.T() })
	}

	return res
}

// BoundMessages returns a list of Message objects that this channel is bound with.
func (c *Channel) BoundMessages() []*Message {
	res := c.Messages()
	return res
}

// BoundOperations returns a list of Operation objects that this channel is bound with.
func (c *Channel) BoundOperations() []*Operation {
	if c.Dummy {
		return nil
	}
	r := lo.FilterMap(c.AllActiveOperationsPromise.T(), func(o common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact(o).(*Operation)
		return op, op.Channel() == c
	})
	// ListPromise is filled up by linker, which doesn't guarantee the order. So, sort items by name
	slices.SortFunc(r, func(a, b *Operation) int { return cmp.Compare(a.Name(), b.Name()) })
	return r
}

// ActiveProtocols returns a unique list of protocols that this channel is bound. This function considers only
// selectable and visible servers.
func (c *Channel) ActiveProtocols() (res []string) {
	protocols := lo.FilterMap(c.BoundServers(), func(s *Server, _ int) (string, bool) {
		return s.Protocol, s.Selectable() && s.Visible()
	})
	return lo.Uniq(protocols)
}

// BindingsProtocols returns a list of protocols that have bindings defined for this channel.
func (c *Channel) BindingsProtocols() (res []string) {
	if c.BindingsType == nil {
		return nil
	}
	if c.BindingsPromise != nil {
		res = append(res, c.BindingsPromise.T().Protocols()...)
	}
	return lo.Uniq(res)
}

// ProtoBindingsValue returns the struct initialization [lang.GoValue] of BindingsType for the given protocol.
// The returned value contains all constant bindings values defined in document for the protocol.
// If no bindings are set for the protocol, returns an empty [lang.GoValue].
func (c *Channel) ProtoBindingsValue(protoName string) common.Artifact {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ChannelBindings", Import: protoName, IsRuntimeImport: true},
		EmptyCurlyBrackets: true,
	}
	if c.BindingsPromise != nil {
		if b, ok := c.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

func (c *Channel) Name() string {
	return c.OriginalName
}

func (c *Channel) Kind() common.ArtifactKind {
	return common.ArtifactKindChannel
}

func (c *Channel) Selectable() bool {
	return !c.Dummy && c.IsSelectable // Select only the channels defined in the `channels` section
}

func (c *Channel) Visible() bool {
	return !c.Dummy
}

func (c *Channel) String() string {
	return "Channel(" + c.OriginalName + ")"
}

func (c *Channel) Pinnable() bool {
	return true
}

type ProtoChannel struct {
	*Channel
	Protocol string
}

func (p *ProtoChannel) Pinnable() bool {
	return false
}

func (p *ProtoChannel) String() string {
	return fmt.Sprintf("ProtoChannel[%s](%s)", p.Protocol, p.OriginalName)
}
