package render

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"

	j "github.com/dave/jennifer/jen"
)

type Channel struct {
	Name                string // Channel name, typically equals to Channel key, can get overridden in x-go-name
	Address             string // Channel address
	GolangName          string // Name of channel struct
	DirectRender        bool   // Typically, it's true if channel is defined in `channels` section, false if in `components` section
	Dummy               bool
	RawName             string                     // Channel key
	ExplicitServerNames []string                   // List of servers the channel is linked with. Empty means "all servers"
	ServersPromises     []*Promise[*Server]        // Servers list this channel is applied to, either explicitly marked or "all servers"
	AllProtoChannels    map[string]common.Renderer // Proto channels for all supported protocols

	Publisher  bool // true if channel has `publish` operation
	Subscriber bool // true if channel has `subscribe` operation

	ParametersStruct *GoStruct // nil if no parameters

	PubMessagePromise   *Promise[*Message] // nil when message is not set
	SubMessagePromise   *Promise[*Message] // nil when message is not set
	FallbackMessageType common.GolangType  // Used in generated code when the message is not set, typically it's `any`

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
	ctx.LogStartRender("Channel", "", c.Name, "definition", c.DirectRendering())
	defer ctx.LogFinishRender()

	// Parameters
	if c.ParametersStruct != nil {
		res = append(res, c.ParametersStruct.RenderDefinition(ctx)...)
	}

	protocols := getServerProtocols(ctx, c.ServersPromises)
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
			// Protocol renderer can maintain multiple Server protocols, so get the protocol name from renderer
			protoName := ctx.ProtoRenderers[p].ProtocolName()
			res = append(res, renderChannelAndOperationBindingsMethod(
				ctx, c.BindingsStruct, chanBindings, pubBindings, subBindings, protoName, protoTitle,
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
	if len(protocols) == 0 {
		res = append(res, j.Comment(fmt.Sprintf("Channel %q is not assigned to any server with supported protocol, so no code to generate", c.Name)))
		ctx.Logger.Info("Channel is not assigned to any server with supported protocol, so no code to generate", "channel", c.Name)
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

	address := c.Address
	if address == "" {
		address = c.RawName
	}

	// Channel1Name(params Chan1Parameters) runtime.ParamString
	return []*j.Statement{
		j.Func().Id(c.GolangName+"Name").
			ParamsFunc(func(g *j.Group) {
				if c.ParametersStruct != nil {
					g.Id("params").Add(utils.ToCode(c.ParametersStruct.RenderUsage(ctx))...)
				}
			}).
			Qual(ctx.RuntimeModule(""), "ParamString").
			BlockFunc(func(bg *j.Group) {
				if c.ParametersStruct == nil {
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(address),
					}))
				} else {
					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, f := range c.ParametersStruct.Fields {
							d[j.Id("params").Dot(f.Name).Dot("Name").Call()] = j.Id("params").Dot(f.Name).Dot("String").Call()
						}
					}))
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(address),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				}
			}),
	}
}
