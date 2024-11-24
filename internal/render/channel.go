package render

import (
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
	"sort"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

type Channel struct {
	Name                string // Channel name, typically equals to Channel key, can get overridden in x-go-name
	GolangName          string // Name of channel struct
	Dummy               bool
	RawName             string                     // Channel key
	ExplicitServerNames []string                   // List of servers the channel is linked with. Empty means "all servers"
	ServersPromises     []*lang.Promise[*Server]   // Servers list this channel is applied to, either explicitly marked or "all servers"

	Publisher  bool // true if channel has `publish` operation
	Subscriber bool // true if channel has `subscribe` operation

	ParametersStruct *lang.GoStruct // nil if no parameters

	PubMessagePromise   *lang.Promise[*Message] // nil when message is not set
	SubMessagePromise   *lang.Promise[*Message] // nil when message is not set
	FallbackMessageType common.GolangType       // Used in generated code when the message is not set, typically it's `any`

	BindingsStruct           *lang.GoStruct           // nil if no bindings are set for channel at all
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
//	if c.ParametersStruct != nil {
//		res = append(res, c.ParametersStruct.D(ctx)...)
//	}
//
//	protocols := c.ServersProtocols(ctx)
//	ctx.Logger.Debug("Channel protocols", "protocols", protocols)
//
//	// Bindings
//	if c.BindingsStruct != nil {
//		ctx.Logger.Trace("Channel bindings")
//		res = append(res, c.BindingsStruct.D(ctx)...)
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
//				ctx, c.BindingsStruct, chanBindings, pubBindings, subBindings, protoName, protoTitle,
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
//func (c Channel) String() string {
//	return "Channel " + c.Name
//}

//func (c Channel) renderChannelNameFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderChannelNameFunc")
//
//	// Channel1Name(params Chan1Parameters) runtime.ParamString
//	return []*j.Statement{
//		j.Func().Id(c.GolangName+"Name").
//			ParamsFunc(func(g *j.Group) {
//				if c.ParametersStruct != nil {
//					g.Id("params").Add(utils.ToCode(c.ParametersStruct.U(ctx))...)
//				}
//			}).
//			Qual(ctx.RuntimeModule(""), "ParamString").
//			BlockFunc(func(bg *j.Group) {
//				if c.ParametersStruct == nil {
//					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
//						j.Id("Expr"): j.Lit(c.RawName),
//					}))
//				} else {
//					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
//						for _, f := range c.ParametersStruct.Fields {
//							d[j.Id("params").Dot(f.Name).Dot("Name").Call()] = j.Id("params").Dot(f.Name).Dot("String").Call()
//						}
//					}))
//					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
//						j.Id("Expr"):       j.Lit(c.RawName),
//						j.Id("Parameters"): j.Id("paramMap"),
//					}))
//				}
//			}),
//	}
//}

// ServersProtocols returns supported protocol list for the given servers, throwing out unsupported ones
// TODO: move to top-level template
func (c Channel) ServersProtocols() []string {
	res := lo.Uniq(lo.FilterMap(c.ServersPromises, func(item *lang.Promise[*Server], _ int) (string, bool) {
		_, ok := ctx.ProtoRenderers[item.Target().Protocol]
		if !ok {
			ctx.Logger.Warnf("Skip protocol %q since it is not supported", item.Target().Protocol)
		}
		return item.Target().Protocol, ok && !item.Target().Dummy
	}))
	sort.Strings(res)
	return res
}

func (c Channel) BindingsProtocols() []string {
	panic("not implemented")
}

func (c Channel) ProtoBindingsValue(protoName string) common.Renderer {
	res := &lang.GoValue{
		Type:             &lang.GoSimple{Name: "ChannelBindings", Import: context.Context.RuntimeModule(protoName)},
		NilCurlyBrackets: true,
	}
	if c.BindingsChannelPromise != nil {
		if b, ok := c.BindingsChannelPromise.Target().Values.Get(protoName); ok {
			ctx.Logger.Debug("Channel bindings", "proto", protoName)
			res = b
		}
	}
	if c.BindingsPublishPromise != nil {
		if v, ok := c.BindingsPublishPromise.Target().Values.Get(protoName); ok {
			ctx.Logger.Debug("Publish operation bindings", "proto", protoName)
			res.StructVals.Set("PublisherBindings", v)
		}
	}
	if c.BindingsSubscribePromise != nil {
		if v, ok := c.BindingsSubscribePromise.Target().Values.Get(protoName); ok {
			ctx.Logger.Debug("Subscribe operation bindings", "proto", protoName)
			res.StructVals.Set("SubscriberBindings", v)
		}
	}
	return res
}

type ProtoChannel struct {
	*Channel
	GolangNameProto string // Channel GolangName name concatenated with protocol name, e.g. Channel1Kafka
	Struct          *lang.GoStruct

	ProtoName string
}

