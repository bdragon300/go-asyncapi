package protocols

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/compile"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type ProtoServerVariable struct {
	ArgName     string
	Enum        []string // TODO: implement validation
	Default     string
	Description string // TODO
}

func BuildServer(
	ctx *common.CompileContext,
	server *compile.Server,
	serverKey string,
	protoName string,
) (*BaseProtoServer, error) {
	const buildProducer = true
	const buildConsumer = true

	srvResult := &BaseProtoServer{
		Name:            serverKey,
		URL:             server.URL,
		ProtocolVersion: server.ProtocolVersion,
		Struct: &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(serverKey, ""),
				Description:  server.Description,
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
		},
	}

	// Server variables
	for _, v := range server.Variables.Entries() {
		ctx.LogDebug("Server variable", "name", v.Key, "proto", protoName)
		srvResult.Variables.Set(v.Key, ProtoServerVariable{
			ArgName:     utils.ToGolangName(v.Key, false),
			Enum:        v.Value.Enum,
			Default:     v.Value.Default,
			Description: v.Value.Description,
		})
	}

	// Channels which are connected to this server
	channelsLnks := render.NewListCbLink[*render.Channel](func(item common.Renderer, path []string) bool {
		ch, ok := item.(*render.Channel)
		if !ok {
			return false
		}
		if len(ch.AppliedServers) > 0 {
			return lo.Contains(ch.AppliedServers, serverKey)
		}
		return ch.AppliedToAllServersLinks != nil
	})
	srvResult.ChannelLinkList = channelsLnks
	ctx.Linker.AddMany(channelsLnks)

	// Producer/consumer
	if buildProducer {
		ctx.LogDebug("Server producer", "proto", protoName)
		fld := render.StructField{
			Name: "producer",
			Type: &render.Simple{Name: "Producer", Package: ctx.RuntimePackage(protoName), IsIface: true},
		}
		srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)
		srvResult.Producer = true
	}
	if buildConsumer {
		ctx.LogDebug("Server consumer", "proto", protoName)
		fld := render.StructField{
			Name: "consumer",
			Type: &render.Simple{Name: "Consumer", Package: ctx.RuntimePackage(protoName), IsIface: true},
		}
		srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)
		srvResult.Consumer = true
	}

	return srvResult, nil
}

type BaseProtoServer struct {
	Name            string
	URL             string
	ProtocolVersion string
	Producer        bool
	Consumer        bool
	Struct          *render.Struct
	ChannelLinkList *render.LinkList[*render.Channel]
	Variables       utils.OrderedMap[string, ProtoServerVariable]
}

func RenderServerConsumerMethods(
	ctx *common.RenderContext,
	serverStruct *render.Struct,
	protoName string,
) []*j.Statement {
	rn := serverStruct.ReceiverName()
	receiver := j.Id(rn).Id(serverStruct.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Consumer").
			Params().
			Qual(ctx.RuntimePackage(protoName), "Consumer").
			Block(
				j.Return(j.Id(rn).Dot("consumer")),
			),
	}
}

func RenderServerProducerMethods(
	ctx *common.RenderContext,
	serverStruct *render.Struct,
	protoName string,
) []*j.Statement {
	rn := serverStruct.ReceiverName()
	receiver := j.Id(rn).Id(serverStruct.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Producer").
			Params().
			Qual(ctx.RuntimePackage(protoName), "Producer").
			Block(
				j.Return(j.Id(rn).Dot("producer")),
			),
	}
}

func RenderServerChannelMethod(
	ctx *common.RenderContext,
	serverStruct, channelStruct *render.Struct,
	channel common.Renderer,
	channelParametersStructNoRender *render.Struct,
) []*j.Statement {
	rn := serverStruct.ReceiverName()
	receiver := j.Id(rn).Id(serverStruct.Name)

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
				j.Return(j.Qual(ctx.GeneratedPackage(channelStruct.PackageName), "Open"+channelStruct.Name).CallFunc(func(g *j.Group) {
					if channelParametersStructNoRender != nil {
						g.Id("params")
					}
					g.Id(rn)
				})),
			),
	}
}

func RenderServerCommonMethods(
	ctx *common.RenderContext,
	serverStruct *render.Struct,
	serverName string,
	protoAbbr string,
) []*j.Statement {
	receiver := j.Id(serverStruct.ReceiverName()).Id(serverStruct.Name)

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(serverName)),
			),

		// Protocol() run.Protocol
		j.Func().Params(receiver.Clone()).Id("Protocol").
			Params().
			Qual(ctx.RuntimePackage(""), "Protocol").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage(""), "Protocol"+protoAbbr)),
			),
	}
}

func RenderServerNewFunc(
	ctx *common.RenderContext,
	serverStruct *render.Struct,
	producer, consumer bool,
	protoName string,
) []*j.Statement {
	return []*j.Statement{
		// NewServer1(producer proto.Producer, consumer proto.Consumer) *Server1
		j.Func().Id(serverStruct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				if producer {
					g.Id("producer").Qual(ctx.RuntimePackage(protoName), "Producer")
				}
				if consumer {
					g.Id("consumer").Qual(ctx.RuntimePackage(protoName), "Consumer")
				}
			}).
			Op("*").Add(utils.ToCode(serverStruct.RenderUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(serverStruct.RenderUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
					if producer {
						d[j.Id("producer")] = j.Id("producer")
					}
					if consumer {
						d[j.Id("consumer")] = j.Id("consumer")
					}
				}))),
			),
	}
}

func RenderServerURLFunc(
	ctx *common.RenderContext,
	serverStruct *render.Struct,
	serverVariables utils.OrderedMap[string, ProtoServerVariable],
	url string,
) []*j.Statement {
	// Server1URL(param1 string, param2 string) run.ParamString
	return []*j.Statement{
		j.Func().Id(serverStruct.Name+"URL").
			ParamsFunc(func(g *j.Group) {
				for _, entry := range serverVariables.Entries() {
					g.Id(entry.Value.ArgName).String()
				}
			}).
			Qual(ctx.RuntimePackage(""), "ParamString").
			BlockFunc(func(bg *j.Group) {
				if serverVariables.Len() > 0 {
					for _, entry := range serverVariables.Entries() {
						if entry.Value.Default != "" {
							bg.If(j.Id(entry.Value.ArgName).Op("==").Lit("")).
								Block(
									j.Id(entry.Value.ArgName).Op("=").Lit(entry.Value.Default),
								)
						}
					}
					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, entry := range serverVariables.Entries() {
							d[j.Lit(entry.Key)] = j.Id(entry.Value.ArgName)
						}
					}))
					bg.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(url),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				} else {
					bg.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(url),
					}))
				}
			}),
	}
}

func RenderServerProtocolVersionConst(serverStruct *render.Struct, protocolVersion string) []*j.Statement {
	return []*j.Statement{
		j.Const().Id(serverStruct.Name + "ProtocolVersion").Op("=").Lit(protocolVersion),
	}
}

func ServerBindingsMethodBody(values *render.StructInit, jsonValues *utils.OrderedMap[string, any]) func(ctx *common.RenderContext, p *render.Func) []*j.Statement {
	return func(ctx *common.RenderContext, p *render.Func) []*j.Statement {
		var res []*j.Statement
		res = append(res,
			j.Id("b").Op(":=").Add(utils.ToCode(values.RenderInit(ctx))...),
		)
		if jsonValues != nil {
			for _, e := range jsonValues.Entries() {
				n := utils.ToLowerFirstLetter(e.Key)
				res = append(res,
					j.Id(n).Op(":=").Lit(e.Value),
					j.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.%[2]s)", n, e.Key)),
				)
			}
		}
		res = append(res, j.Return(j.Id("b")))
		return res
	}
}
