package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

type Channel struct {
	OriginalName     string // Channel name, typically equals to Channel key, can get overridden in x-go-name
	Dummy            bool
	IsComponent bool // true if channel is defined in `components` section
	IsPublisher bool
	IsSubscriber bool

	ServersPromises   []*lang.Promise[*Server] // Servers that this channel is bound with. Empty list means "no servers bound".
	AllActiveServersPromise *lang.ListPromise[common.Renderable]

	ParametersType *lang.GoStruct // nil if no parameters

	MessagesPromises []*lang.Ref

	// All operations we know about for further selecting ones that are bound to this channel
	// We can't collect here just the operations already bound with this channel, because the channel in operation
	// is also a promise, and the order of promises resolving is not guaranteed. So we just collect all operations
	// and then filter them by the channel on render stage.
	AllActiveOperationsPromise *lang.ListPromise[common.Renderable]

	BindingsType             *lang.GoStruct           // nil if no bindings are set for channel at all
	BindingsPromise          *lang.Promise[*Bindings] // nil if channel bindings are not set

	ProtoChannels []*ProtoChannel // Proto channels for each supported protocol
}

func (c *Channel) Kind() common.ObjectKind {
	return common.ObjectKindChannel
}

func (c *Channel) Selectable() bool {
	return !c.Dummy && !c.IsComponent // Select only the channels defined in the `channels` section
}

func (c *Channel) Visible() bool {
	return !c.Dummy
}

func (c *Channel) Name() string {
	return utils.CapitalizeUnchanged(c.OriginalName)
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
	if len(c.ServersPromises) == 0 {
		return c.AllActiveServersPromise.T()
	}
	return lo.Map(c.ServersPromises, func(s *lang.Promise[*Server], _ int) common.Renderable { return s.T() })
}

func (c *Channel) BoundMessages() []common.Renderable {
	ops := c.BoundOperations()
	opMsgs := lo.FlatMap(ops, func(o common.Renderable, _ int) []common.Renderable {
		op := common.DerefRenderable(o).(*Operation)
		return op.Messages()
	})
	r := lo.Without(c.Messages(), opMsgs...)
	return r
}

func (c *Channel) BoundOperations() []common.Renderable {
	r := lo.Filter(c.AllActiveOperationsPromise.T(), func(o common.Renderable, _ int) bool {
		op := common.DerefRenderable(o).(*Operation)
		return common.CheckSameRenderables(op.Channel(), c)
	})
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
		Type:               &lang.GoSimple{TypeName: "ChannelBindings", Import: common.GetContext().RuntimeModule(protoName)},
		EmptyCurlyBrackets: true,
	}
	if c.BindingsPromise != nil {
		if b, ok := c.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

func (c *Channel) Messages() []common.Renderable {
	return lo.Map(c.MessagesPromises, func(prm *lang.Ref, _ int) common.Renderable { return prm.T() })
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
