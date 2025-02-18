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

type Channel struct {
	OriginalName string // Channel name, typically equals to Channel key, can get overridden in x-go-name
	Address      string
	Dummy        bool
	IsSelectable bool // true if channel should get to selections
	IsPublisher  bool
	IsSubscriber bool

	ServersPromises         []*lang.Promise[*Server] // Servers that this channel is bound with. Empty list means "no servers bound".
	AllActiveServersPromise *lang.ListPromise[common.Renderable]

	ParametersPromises []*lang.Ref    // nil if no parameters
	ParametersType     *lang.GoStruct // nil if no parameters

	MessagesRefs []*lang.Ref

	// All operations we know about for further selecting ones that are bound to this channel
	// We can't collect here just the operations already bound with this channel, because the channel in operation
	// is also a promise, and the order of promises resolving is not guaranteed. So we just collect all operations
	// and then filter them by the channel on render stage.
	AllActiveOperationsPromise *lang.ListPromise[common.Renderable]

	BindingsType    *lang.GoStruct           // nil if no bindings are set for channel at all
	BindingsPromise *lang.Promise[*Bindings] // nil if channel bindings are not set

	ProtoChannels []*ProtoChannel // Proto channels for each supported protocol
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

func (c *Channel) Name() string {
	return c.OriginalName
}

func (c *Channel) String() string {
	return "Channel " + c.OriginalName
}

func (c *Channel) SelectProtoObject(protocol string) common.Renderable {
	res := lo.Filter(c.ProtoChannels, func(p *ProtoChannel, _ int) bool {
		return p.Selectable() && p.Protocol == protocol
	})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}

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

func (c *Channel) BoundMessages() []common.Renderable {
	ops := c.BoundOperations()
	opMsgs := lo.FlatMap(ops, func(o common.Renderable, _ int) []common.Renderable {
		op := common.DerefRenderable(o).(*Operation)
		return op.Messages()
	})
	r := utils.WithoutBy(c.Messages(), opMsgs, common.CheckSameRenderables)
	return r
}

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

func (c *Channel) Bindings() *Bindings {
	if c.BindingsPromise != nil {
		return c.BindingsPromise.T()
	}
	return nil
}

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

func (c *Channel) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ChannelBindings", Import: protoName, RuntimeImport: true},
		EmptyCurlyBrackets: true,
	}
	if c.BindingsPromise != nil {
		if b, ok := c.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

func (c *Channel) Parameters() []common.Renderable {
	r := lo.Map(c.ParametersPromises, func(prm *lang.Ref, _ int) common.Renderable { return prm })
	return r
}

func (c *Channel) Messages() []common.Renderable {
	return lo.Map(c.MessagesRefs, func(prm *lang.Ref, _ int) common.Renderable { return prm })
}

type ProtoChannel struct {
	*Channel
	Type *lang.GoStruct

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
