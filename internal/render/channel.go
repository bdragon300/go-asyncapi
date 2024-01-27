package render

import (
	"sort"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"

	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Channel struct {
	Name                string // Channel name, typically equals to Channel key, can get overridden in x-go-name
	DirectRender        bool   // Typically, it's true if channel is defined in `channels` section, false if in `components` section
	Dummy               bool
	ChannelKey          string                     // Channel key
	ExplicitServerNames []string                   // List of servers the channel is linked with. Empty means "all servers"
	ServersPromises     []*Promise[*Server]        // Servers list this channel is applied to, either explicitly marked or "all servers"
	AllProtoChannels    map[string]common.Renderer // Proto channels for all supported protocols

	ParametersStruct *GoStruct // nil if no parameters

	BindingsStruct           *GoStruct           // nil if no bindings are set for channel at all
	BindingsChannelPromise   *Promise[*Bindings] // nil if channel bindings are not set
	BindingsSubscribePromise *Promise[*Bindings] // nil if subscribe operation bindings are not set
	BindingsPublishPromise   *Promise[*Bindings] // nil if publish operation bindings are not set
}

func (c Channel) DirectRendering() bool {
	return c.DirectRender && !c.Dummy
}

func (c Channel) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	ctx.LogRender("Channel", "", c.Name, "definition", c.DirectRendering())
	defer ctx.LogReturn()

	// Parameters
	if c.ParametersStruct != nil {
		res = append(res, c.ParametersStruct.RenderDefinition(ctx)...)
	}

	protocols := c.getServerProtocols(ctx)
	ctx.Logger.Debug("Channel protocols", "protocols", protocols)

	// Bindings
	if c.BindingsStruct != nil {
		ctx.Logger.Trace("Channel bindings")
		res = append(res, c.BindingsStruct.RenderDefinition(ctx)...)

		var chanBindings, pubBindings, subBindings *Bindings
		if c.BindingsChannelPromise != nil {
			chanBindings = c.BindingsChannelPromise.Target()
		}
		if c.BindingsPublishPromise != nil {
			pubBindings = c.BindingsPublishPromise.Target()
		}
		if c.BindingsSubscribePromise != nil {
			subBindings = c.BindingsSubscribePromise.Target()
		}
		for _, p := range protocols {
			protoTitle := ctx.ProtoRenderers[p].ProtocolTitle()
			res = append(res, renderChannelAndOperationBindingsMethod(
				ctx, c.BindingsStruct, chanBindings, pubBindings, subBindings, p, protoTitle,
			)...)
		}
	}

	res = append(res, c.renderChannelNameFunc(ctx)...)

	// Proto channels
	for _, p := range protocols {
		r, ok := c.AllProtoChannels[p]
		if !ok {
			ctx.Logger.Warnf("Skip protocol %q since it is not supported", p)
			continue
		}
		res = append(res, r.RenderDefinition(ctx)...)
	}
	return res
}

func (c Channel) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (c Channel) ID() string {
	return c.Name
}

func (c Channel) String() string {
	return "Channel " + c.Name
}

func (c Channel) renderChannelNameFunc(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderChannelNameFunc")

	// Channel1Name(params Chan1Parameters) runtime.ParamString
	return []*j.Statement{
		j.Func().Id(utils.ToGolangName(c.Name, true)+"Name").
			ParamsFunc(func(g *j.Group) {
				if c.ParametersStruct != nil {
					g.Id("params").Add(utils.ToCode(c.ParametersStruct.RenderUsage(ctx))...)
				}
			}).
			Qual(ctx.RuntimeModule(""), "ParamString").
			BlockFunc(func(bg *j.Group) {
				if c.ParametersStruct == nil {
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(c.ChannelKey),
					}))
				} else {
					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, f := range c.ParametersStruct.Fields {
							d[j.Id("params").Dot(f.Name).Dot("Name").Call()] = j.Id("params").Dot(f.Name).Dot("String").Call()
						}
					}))
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(c.ChannelKey),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				}
			}),
	}
}

func (c Channel) getServerProtocols(ctx *common.RenderContext) []string {
	res := lo.FilterMap(c.ServersPromises, func(item *Promise[*Server], index int) (string, bool) {
		_, ok := ctx.ProtoRenderers[item.Target().Protocol]
		if !ok {
			ctx.Logger.Warnf("Skip protocol %q since it is not supported", item.Target().Protocol)
		}
		return item.Target().Protocol, ok && !item.Target().Dummy
	})
	sort.Strings(res)
	return res
}
