package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

type Channel struct {
	Name            string // Channel name, typically equals to Channel key, can get overridden in x-go-name
	TypeNamePrefix  string // Prefix for a proto channel type name
	Dummy           bool
	SpecKey         string                     // Key in the source document
	SpecServerNames []string                   // List of servers the channel is linked with. Empty means "all servers"
	ServersPromise  *lang.ListPromise[*Server] // Servers list this channel is applied to, either explicitly marked or "all servers"

	IsPublisher  bool // true if channel has `publish` operation
	IsSubscriber bool // true if channel has `subscribe` operation

	ParametersType *lang.GoStruct // nil if no parameters

	PublisherMessageTypePromise *lang.Promise[*Message] // nil when message is not set
	SubscribeMessageTypePromise *lang.Promise[*Message] // nil when message is not set

	BindingsType             *lang.GoStruct           // nil if no bindings are set for channel at all
	BindingsChannelPromise   *lang.Promise[*Bindings] // nil if channel bindings are not set
	BindingsSubscribePromise *lang.Promise[*Bindings] // nil if subscribe operation bindings are not set
	BindingsPublishPromise   *lang.Promise[*Bindings] // nil if publish operation bindings are not set
}

func (c Channel) Kind() common.ObjectKind {
	return common.ObjectKindChannel
}

func (c Channel) Selectable() bool {
	return !c.Dummy
}

//func (c Channel) Selectable() bool {
//	return c.HasDefinition && !c.Dummy
//}

//func (c Channel) D(ctx *common.RenderContext) []*j.Statement {
//	var res []*j.Statement
//	ctx.LogStartRender("Channel", "", c.Name, "definition", c.Selectable())
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
//		res = append(res, j.Comment(fmt.Sprintf("Channel %q is not assigned to any server with supported protocol, so no code to generate", c.Name)))
//		ctx.Logger.Info("Channel is not assigned to any server with supported protocol, so no code to generate", "channel", c.Name)
//	}
//	return res
//}

//func (c Channel) U(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}
//
//func (c Channel) ID() string {
//	return c.Name
//}
//
func (c Channel) String() string {
	return "Channel " + c.Name
}

//func (c Channel) renderChannelNameFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderChannelNameFunc")
//
//	// Channel1Name(params Chan1Parameters) runtime.ParamString
//	return []*j.Statement{
//		j.Func().Id(c.TypeNamePrefix+"Name").
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
//							d[j.Id("params").Dot(f.Name).Dot("Name").Call()] = j.Id("params").Dot(f.Name).Dot("String").Call()
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

// ServersProtocols returns supported protocol list for the given servers, throwing out unsupported ones
//func (c Channel) ServersProtocols() []string {
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

func (c Channel) BindingsProtocols() (res []string) {
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

func (c Channel) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{Name: "ChannelBindings", Import: context.Context.RuntimeModule(protoName)},
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

	ProtoName string
}

func (p ProtoChannel) String() string {
	return "ProtoChannel " + p.Name
}