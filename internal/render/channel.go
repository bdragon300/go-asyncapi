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

//ServersProtocols returns supported protocol list for the given servers, throwing out unsupported ones
//func (c Channel) serversProtocols() []string {
//	res := lo.Uniq(lo.FilterMap(c.ServersPromise.T(), func(item *Server, _ int) (string, bool) {
//		_, ok := ctx.ProtoRenderers[item.Protocol]
//		if !ok {
//			ctx.Logger.Warnf("Skip protocol %q since it is not supported", item.Protocol)
//		}
//		return item.Protocol, ok && !item.Dummy
//	}))
//	sort.Strings(res)
//	return res
//}

//func (c Channel) Selectable() bool {
//	return c.HasDefinition && !c.Dummy
//}

//func (c Channel) D(ctx *common.RenderContext) []*j.Statement {
//	var res []*j.Statement
//	ctx.LogStartRender("Channel", "", c.GetOriginalName, "definition", c.Selectable())
//	defer ctx.LogFinishRender()
//
//	// Parameters
//	if c.ParametersType != nil {
//		res = append(res, c.ParametersType.D(ctx)...)
//	}
//
//	protocols := c.ServersProtocols(ctx)
//	ctx.Logger.Debug("Channel protocols", "protocols", protocols)
//
//	// Bindings
//	if c.BindingsType != nil {
//		ctx.Logger.Trace("Channel bindings")
//		res = append(res, c.BindingsType.D(ctx)...)
//
//		var chanBindings, pubBindings, subBindings *Bindings
//		if c.BindingsChannelPromise != nil {
//			chanBindings = c.BindingsChannelPromise.Target()
//		}
//		if c.BindingsPublishPromise != nil {
//			pubBindings = c.BindingsPublishPromise.Target()
//		}
//		if c.BindingsSubscribePromise != nil {
//			subBindings = c.BindingsSubscribePromise.Target()
//		}
//		for _, p := range protocols {
//			protoTitle := ctx.ProtoRenderers[p].ProtocolTitle()
//			// Protocol renderer can maintain multiple Server protocols, so get the protocol name from renderer
//			protoName := ctx.ProtoRenderers[p].ProtocolName()
//			res = append(res, renderChannelAndOperationBindingsMethod(
//				ctx, c.BindingsType, chanBindings, pubBindings, subBindings, protoName, protoTitle,
//			)...)
//		}
//	}
//
//	res = append(res, c.renderChannelNameFunc(ctx)...)
//
//	// Proto channels
//	for _, p := range protocols {
//		r, ok := c.AllProtoChannels[p]
//		if !ok {
//			ctx.Logger.Warnf("Skip protocol %q since it is not supported", p)
//			continue
//		}
//		res = append(res, r.D(ctx)...)
//	}
//	if len(protocols) == 0 {
//		res = append(res, j.Comment(fmt.Sprintf("Channel %q is not assigned to any server with supported protocol, so no code to generate", c.GetOriginalName)))
//		ctx.Logger.Info("Channel is not assigned to any server with supported protocol, so no code to generate", "channel", c.GetOriginalName)
//	}
//	return res
//}

//func (c Channel) U(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}
//
//func (c Channel) ID() string {
//	return c.GetOriginalName
//}
//
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

//func (c Channel) renderChannelNameFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderChannelNameFunc")
//
//	// Channel1Name(params Chan1Parameters) runtime.ParamString
//	return []*j.Statement{
//		j.Func().Id(c.TypeNamePrefix+"GetOriginalName").
//			ParamsFunc(func(g *j.Group) {
//				if c.ParametersType != nil {
//					g.Id("params").Add(utils.ToCode(c.ParametersType.U(ctx))...)
//				}
//			}).
//			Qual(ctx.RuntimeModule(""), "ParamString").
//			BlockFunc(func(bg *j.Group) {
//				if c.ParametersType == nil {
//					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
//						j.Id("Expr"): j.Lit(c.SpecKey),
//					}))
//				} else {
//					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
//						for _, f := range c.ParametersType.Fields {
//							d[j.Id("params").Dot(f.GetOriginalName).Dot("GetOriginalName").Call()] = j.Id("params").Dot(f.GetOriginalName).Dot("String").Call()
//						}
//					}))
//					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
//						j.Id("Expr"):       j.Lit(c.SpecKey),
//						j.Id("Parameters"): j.Id("paramMap"),
//					}))
//				}
//			}),
//	}
//}

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
			//ctx.Logger.Debug("Channel bindings", "proto", protoName)
			res = b
		}
	}
	if c.BindingsPublishPromise != nil {
		if v, ok := c.BindingsPublishPromise.T().Values.Get(protoName); ok {
			//ctx.Logger.Debug("Publish operation bindings", "proto", protoName)
			res.StructValues.Set("PublisherBindings", v)
		}
	}
	if c.BindingsSubscribePromise != nil {
		if v, ok := c.BindingsSubscribePromise.T().Values.Get(protoName); ok {
			//ctx.Logger.Debug("Subscribe operation bindings", "proto", protoName)
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
