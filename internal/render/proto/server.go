package proto

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type ServerVariable struct {
	ArgName     string
	Enum        []string // TODO: implement validation
	Default     string
	Description string // TODO
}

type BaseProtoServer struct {
	Name            string // TODO: move fields to abstract server
	URL             string
	ProtocolVersion string
	Struct          *render.GoStruct
	ChannelsPromise *render.ListPromise[*render.Channel]
	Variables       types.OrderedMap[string, ServerVariable]

	ProtoName, ProtoAbbr string
}

func (ps BaseProtoServer) RenderConsumerMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderConsumerMethods", "proto", ps.ProtoName)

	rn := ps.Struct.ReceiverName()
	receiver := j.Id(rn).Id(ps.Struct.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Consumer").
			Params().
			Qual(ctx.RuntimeModule(ps.ProtoName), "Consumer").
			Block(
				j.Return(j.Id(rn).Dot("consumer")),
			),
	}
}

func (ps BaseProtoServer) RenderProducerMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderProducerMethods", "proto", ps.ProtoName)

	rn := ps.Struct.ReceiverName()
	receiver := j.Id(rn).Id(ps.Struct.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Producer").
			Params().
			Qual(ctx.RuntimeModule(ps.ProtoName), "Producer").
			Block(
				j.Return(j.Id(rn).Dot("producer")),
			),
	}
}

func (ps BaseProtoServer) RenderOpenChannelMethod(ctx *common.RenderContext, channelStruct *render.GoStruct, channel common.Renderer, channelParametersStructNoRender *render.GoStruct) []*j.Statement {
	ctx.Logger.Trace("RenderOpenChannelMethod", "proto", ps.ProtoName)

	rn := ps.Struct.ReceiverName()
	receiver := j.Id(rn).Id(ps.Struct.Name)

	return []*j.Statement{
		// Method OpenChannel1Proto(params Channel1Parameters) (*Channel1Proto, error)
		j.Func().Params(receiver.Clone()).Id("Open"+channelStruct.Name).
			ParamsFunc(func(g *j.Group) {
				if channelParametersStructNoRender != nil {
					g.Id("params").Add(utils.ToCode(channelParametersStructNoRender.RenderUsage(ctx))...)
				}
			}).
			Params(j.Op("*").Add(utils.ToCode(channel.RenderUsage(ctx))...), j.Error()).
			Block(
				j.Return(j.Qual(ctx.GeneratedModule(channelStruct.Import), "Open"+channelStruct.Name).CallFunc(func(g *j.Group) {
					if channelParametersStructNoRender != nil {
						g.Id("params")
					}
					g.Id(rn)
				})),
			),
	}
}

func (ps BaseProtoServer) RenderCommonMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderCommonMethods", "proto", ps.ProtoName)

	receiver := j.Id(ps.Struct.ReceiverName()).Id(ps.Struct.Name)

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(ps.Name)),
			),
	}
}

func (ps BaseProtoServer) RenderNewFunc(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderNewFunc", "proto", ps.ProtoName)

	return []*j.Statement{
		// NewServer1(producer proto.Producer, consumer proto.Consumer) *Server1
		j.Func().Id(ps.Struct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				g.Id("producer").Qual(ctx.RuntimeModule(ps.ProtoName), "Producer")
				g.Id("consumer").Qual(ctx.RuntimeModule(ps.ProtoName), "Consumer")
			}).
			Op("*").Add(utils.ToCode(ps.Struct.RenderUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(ps.Struct.RenderUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
					d[j.Id("producer")] = j.Id("producer")
					d[j.Id("consumer")] = j.Id("consumer")
				}))),
			),
	}
}

func (ps BaseProtoServer) RenderURLFunc(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderURLFunc", "proto", ps.ProtoName)

	// Server1URL(param1 string, param2 string) run.ParamString
	return []*j.Statement{
		j.Func().Id(ps.Struct.Name+"URL").
			ParamsFunc(func(g *j.Group) {
				for _, entry := range ps.Variables.Entries() {
					g.Id(entry.Value.ArgName).String()
				}
			}).
			Qual(ctx.RuntimeModule(""), "ParamString").
			BlockFunc(func(bg *j.Group) {
				if ps.Variables.Len() > 0 {
					for _, entry := range ps.Variables.Entries() {
						if entry.Value.Default != "" {
							bg.If(j.Id(entry.Value.ArgName).Op("==").Lit("")).
								Block(
									j.Id(entry.Value.ArgName).Op("=").Lit(entry.Value.Default),
								)
						}
					}
					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, entry := range ps.Variables.Entries() {
							d[j.Lit(entry.Key)] = j.Id(entry.Value.ArgName)
						}
					}))
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(ps.URL),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				} else {
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(ps.URL),
					}))
				}
			}),
	}
}

func (ps BaseProtoServer) RenderProtocolVersionConst(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderProtocolVersionConst", "proto", ps.ProtoName)

	return []*j.Statement{
		j.Const().Id(ps.Struct.Name + "ProtocolVersion").Op("=").Lit(ps.ProtocolVersion),
	}
}
