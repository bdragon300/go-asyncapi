package render

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

// Channel represents the channel object.
type Channel struct {
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
	AllActiveServersPromise *lang.ListPromise[common.Renderable]

	// ParametersPromises is a list of refs to channel parameters.
	ParametersPromises []*lang.Ref
	// ParametersType is a Go struct for channel parameters. Nil if no parameters are set.
	ParametersType *lang.GoStruct

	// MessagesRefs is a list of messages listed in the channel definition in document.
	MessagesRefs []*lang.Ref

	// AllActiveOperationsPromise contains all active operations in the document.
	//
	// On compiling stage we don't know which operations are bound to a particular channel.
	// So we just collect all operations to this field and postpone filtering them until the rendering stage.
	//
	// We could use a promise callback to filter operations by channel, but the channel in operation is also a promise,
	// and the order of promises resolving is not guaranteed.
	AllActiveOperationsPromise *lang.ListPromise[common.Renderable]

	// BindingsType is a Go struct for channel bindings. Nil if no bindings are set.
	BindingsType *lang.GoStruct
	// BindingsPromise is a promise to channel bindings contents. Nil if no bindings are set.
	BindingsPromise *lang.Promise[*Bindings]

	// ProtoChannels is a list of prebuilt ProtoChannel objects for each supported protocol
	ProtoChannels []*ProtoChannel
}

// Parameters returns a list of [lang.Ref] to Parameter.
func (c *Channel) Parameters() []common.Renderable {
	r := lo.Map(c.ParametersPromises, func(prm *lang.Ref, _ int) common.Renderable { return prm })
	return r
}

// Messages returns a list of [lang.Ref] to Message.
func (c *Channel) Messages() []common.Renderable {
	return lo.Map(c.MessagesRefs, func(prm *lang.Ref, _ int) common.Renderable { return prm })
}

// Bindings returns the [Bindings] object or nil if no bindings are set.
func (c *Channel) Bindings() *Bindings {
	if c.BindingsPromise != nil {
		return c.BindingsPromise.T()
	}
	return nil
}

// SelectProtoObject returns a selectable ProtoChannel object for the given protocol or nil if not found or if
// ProtoChannel is not selectable.
func (c *Channel) SelectProtoObject(protocol string) common.Renderable {
	res := lo.Filter(c.ProtoChannels, func(p *ProtoChannel, _ int) bool {
		return p.Selectable() && p.Protocol == protocol
	})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}

// BoundServers returns a list of Server or lang.Ref to Server objects that this channel is bound with.
func (c *Channel) BoundServers() []common.Renderable {
	if c.Dummy {
		return nil
	}

	var res []common.Renderable
	if len(c.ServersPromises) == 0 {
		res = c.AllActiveServersPromise.T()
		// ListPromise is filled up by linker, which doesn't guarantee the order. So, sort items by name
		slices.SortFunc(res, func(a, b common.Renderable) int { return cmp.Compare(a.Name(), b.Name()) })
	} else {
		res = lo.Map(c.ServersPromises, func(s *lang.Promise[*Server], _ int) common.Renderable { return s.T() })
	}

	return res
}

// BoundMessages returns a list of Message or lang.Ref to Message objects that this channel is bound with.
func (c *Channel) BoundMessages() []common.Renderable {
	ops := c.BoundOperations()
	opMsgs := lo.FlatMap(ops, func(o common.Renderable, _ int) []common.Renderable {
		op := common.DerefRenderable(o).(*Operation)
		return op.Messages()
	})
	r := utils.WithoutBy(c.Messages(), opMsgs, common.CheckSameRenderables)
	return r
}

// BoundOperations returns a list of Operation or lang.Ref to Operation objects that this channel is bound with.
func (c *Channel) BoundOperations() []common.Renderable {
	if c.Dummy {
		return nil
	}
	r := lo.Filter(c.AllActiveOperationsPromise.T(), func(o common.Renderable, _ int) bool {
		op := common.DerefRenderable(o).(*Operation)
		return common.CheckSameRenderables(op.Channel(), c)
	})
	// ListPromise is filled up by linker, which doesn't guarantee the order. So, sort items by name
	slices.SortFunc(r, func(a, b common.Renderable) int { return cmp.Compare(a.Name(), b.Name()) })
	return r
}

// BindingsProtocols returns a list of protocols that have bindings defined for this channel.
func (c *Channel) BindingsProtocols() (res []string) {
	if c.BindingsType == nil {
		return nil
	}
	if c.BindingsPromise != nil {
		res = append(res, c.BindingsPromise.T().Values.Keys()...)
		res = append(res, c.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

// ProtoBindingsValue returns the struct initialization [lang.GoValue] of BindingsType for the given protocol.
// The returned value contains all constant bindings values defined in document for the protocol.
// If no bindings are set for the protocol, returns an empty [lang.GoValue].
func (c *Channel) ProtoBindingsValue(protoName string) common.Renderable {
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

func (c *Channel) Kind() common.ObjectKind {
	return common.ObjectKindChannel
}

func (c *Channel) Selectable() bool {
	return !c.Dummy && c.IsSelectable // Select only the channels defined in the `channels` section
}

func (c *Channel) Visible() bool {
	return !c.Dummy
}

func (c *Channel) String() string {
	return "Channel " + c.OriginalName
}

type ProtoChannel struct {
	*Channel
	// Type is a protocol-specific channel's Go struct
	Type     *lang.GoStruct
	Protocol string
}

func (p *ProtoChannel) Selectable() bool {
	return !p.Dummy && p.isBound()
}

func (p *ProtoChannel) String() string {
	return fmt.Sprintf("ProtoChannel[%s] %s", p.Protocol, p.OriginalName)
}

// isBound returns true if channel is bound to at least one server with supported protocol
func (p *ProtoChannel) isBound() bool {
	protos := lo.Map(p.BoundServers(), func(s common.Renderable, _ int) string {
		srv := common.DerefRenderable(s).(*Server)
		return srv.Protocol
	})
	r := lo.Contains(protos, p.Protocol)
	return r
}
