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
	BoundServerNames []string  // List of servers the channel is linked with. Empty list means "all servers"
	ServersPromise   *lang.ListPromise[*Server] // Servers list this channel is bound with. Empty list means "no servers bound".

	IsPublisher  bool // true if channel has `publish` operation
	IsSubscriber bool // true if channel has `subscribe` operation
	IsComponent bool // true if channel is defined in `components` section

	ParametersType *lang.GoStruct // nil if no parameters

	PublisherMessageTypePromise  *lang.Promise[*Message] // nil when message is not set
	SubscriberMessageTypePromise *lang.Promise[*Message] // nil when message is not set

	BindingsType             *lang.GoStruct           // nil if no bindings are set for channel at all
	BindingsChannelPromise   *lang.Promise[*Bindings] // nil if channel bindings are not set
	BindingsSubscribePromise *lang.Promise[*Bindings] // nil if subscribe operation bindings are not set
	BindingsPublishPromise   *lang.Promise[*Bindings] // nil if publish operation bindings are not set

	ProtoChannels []*ProtoChannel // Proto channels for each supported protocol
}

func (c *Channel) Kind() common.ObjectKind {
	return common.ObjectKindChannel
}

func (c *Channel) Selectable() bool {
	return !c.Dummy && !c.IsComponent // Select only the channels defined in the `channels` section`
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

func (c *Channel) PublisherMessageType() *Message {
	if c.PublisherMessageTypePromise != nil {
		return c.PublisherMessageTypePromise.T()
	}
	return nil
}

func (c *Channel) SubscriberMessageType() *Message {
	if c.SubscriberMessageTypePromise != nil {
		return c.SubscriberMessageTypePromise.T()
	}
	return nil
}

func (c *Channel) BindingsChannel() *Bindings {
	if c.BindingsChannelPromise != nil {
		return c.BindingsChannelPromise.T()
	}
	return nil
}

func (c *Channel) BindingsPublish() *Bindings {
	if c.BindingsPublishPromise != nil {
		return c.BindingsPublishPromise.T()
	}
	return nil
}

func (c *Channel) BindingsSubscribe() *Bindings {
	if c.BindingsSubscribePromise != nil {
		return c.BindingsSubscribePromise.T()
	}
	return nil
}

func (c *Channel) BindingsProtocols() (res []string) {
	if c.BindingsType == nil {
		return nil
	}
	if c.BindingsChannelPromise != nil {
		res = append(res, c.BindingsChannelPromise.T().Values.Keys()...)
		res = append(res, c.BindingsChannelPromise.T().JSONValues.Keys()...)
	}
	if c.BindingsPublishPromise != nil {
		res = append(res, c.BindingsPublishPromise.T().Values.Keys()...)
		res = append(res, c.BindingsPublishPromise.T().JSONValues.Keys()...)
	}
	if c.BindingsSubscribePromise != nil {
		res = append(res, c.BindingsSubscribePromise.T().Values.Keys()...)
		res = append(res, c.BindingsSubscribePromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

func (c *Channel) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ChannelBindings", Import: common.GetContext().RuntimeModule(protoName)},
		EmptyCurlyBrackets: true,
	}
	if c.BindingsChannelPromise != nil {
		if b, ok := c.BindingsChannelPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	if c.BindingsPublishPromise != nil {
		if v, ok := c.BindingsPublishPromise.T().Values.Get(protoName); ok {
			res.StructValues.Set("PublisherBindings", v)
		}
	}
	if c.BindingsSubscribePromise != nil {
		if v, ok := c.BindingsSubscribePromise.T().Values.Get(protoName); ok {
			res.StructValues.Set("SubscriberBindings", v)
		}
	}
	return res
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
	protos := lo.Map(p.ServersPromise.T(), func(s *Server, _ int) string { return s.Protocol })
	r := lo.Contains(
		protos,
		p.Protocol,
	)
	return r
}
