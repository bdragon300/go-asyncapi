package kafka

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type ProtoServerBindings struct {
	StructValues utils.OrderedMap[string, any]
}

type ProtoServerVariable struct {
	ArgName     string
	Enum        []string // TODO: implement validation
	Default     string
	Description string // TODO
}

type ProtoServer struct {
	Name            string
	URL             string
	ProtocolVersion string
	Struct          *assemble.Struct
	ChannelLinkList *assemble.LinkList[*assemble.Channel]
	Producer        bool
	Consumer        bool
	Bindings        *ProtoServerBindings
	Variables       utils.OrderedMap[string, ProtoServerVariable]
}

func (p ProtoServer) AllowRender() bool {
	return true
}

func (p ProtoServer) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	if p.Bindings != nil {
		res = append(res, p.assembleBindings(ctx, p.Bindings)...)
	}
	res = append(res, p.assembleConnArgsFunc(ctx)...)
	res = append(res, p.assembleNewFunc(ctx)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.assembleCommonMethods()...)
	res = append(res, p.assembleChannelMethods(ctx)...)
	if p.Producer {
		res = append(res, p.assembleProducerMethods(ctx)...)
	}
	if p.Consumer {
		res = append(res, p.assembleConsumerMethods(ctx)...)
	}
	return res
}

func (p ProtoServer) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoServer) assembleBindings(ctx *common.AssembleContext, bindings *ProtoServerBindings) []*j.Statement {
	vals := lo.FromEntries(lo.Map(bindings.StructValues.Entries(), func(item lo.Entry[string, any], index int) lo.Entry[j.Code, j.Code] {
		return lo.Entry[j.Code, j.Code]{Key: j.Id(item.Key), Value: j.Lit(item.Value)}
	}))
	return []*j.Statement{
		j.Func().Id(p.Struct.Name+"Bindings").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "ServerBindings").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage("kafka"), "ServerBindings").Values(j.Dict(vals))),
			),
	}
}

func (p ProtoServer) assembleConnArgsFunc(ctx *common.AssembleContext) []*j.Statement {
	var args []j.Code
	var body []j.Code
	if p.Variables.Len() > 0 {
		mapVals := j.Dict{}
		for _, v := range p.Variables.Entries() {
			mapVals[j.Lit(v.Key)] = j.Id(v.Value.ArgName)
			args = append(args, j.Id(v.Value.ArgName).String())
			if v.Value.Default != "" {
				body = append(body,
					j.If(j.Id(v.Value.ArgName).Op("==").Lit("")).
						Block(j.Id(v.Value.ArgName).Op("=").Lit(v.Value.Default)),
				)
			}
		}
		body = append(body,
			j.Op("p := map[string]string").Values(mapVals),
			j.Op("connArgs.ProtocolVersion = ").Lit(p.ProtocolVersion),
			j.Op("connArgs.URL, _, err =").Qual(ctx.RuntimePackage("3rdparty/uritemplates"), "Expand").Call(
				j.Lit(p.URL), j.Id("p"),
			),
			j.Return(),
		)
	} else {
		body = append(body,
			j.Op("connArgs.ProtocolVersion = ").Lit(p.ProtocolVersion),
			j.Op("connArgs.URL = ").Lit(p.URL),
			j.Return(),
		)
	}

	return []*j.Statement{
		j.Func().Id(p.Struct.Name+"ConnArgs").
			Params(args...).
			Params(j.Id("connArgs").Qual(ctx.RuntimePackage(""), "ServerConnArgs"), j.Err().Error()).
			Block(body...),
	}
}

func (p ProtoServer) assembleNewFunc(ctx *common.AssembleContext) []*j.Statement {
	var params []j.Code
	vals := j.Dict{}
	if p.Producer {
		params = append(params, j.Id("producer").Qual(ctx.RuntimePackage("kafka"), "Producer"))
		vals[j.Id("producer")] = j.Id("producer")
	}
	if p.Consumer {
		params = append(params, j.Id("consumer").Qual(ctx.RuntimePackage("kafka"), "Consumer"))
		vals[j.Id("consumer")] = j.Id("consumer")
	}
	return []*j.Statement{
		// NewServer1Server(producer kafka.Producer, consumer kafka.Consumer) *Server1Server
		j.Func().Id(p.Struct.NewFuncName()).
			Params(params...).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(vals)),
			),
	}
}

func (p ProtoServer) assembleCommonMethods() []*j.Statement {
	receiver := j.Id(p.Struct.ReceiverName()).Id(p.Struct.Name)

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Name)),
			),
	}
}

func (p ProtoServer) assembleChannelMethods(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	for _, ch := range p.ChannelLinkList.Targets() {
		// Method Channel1PubChannel(publisherParams, subscriberParams kafka.ChannelParams) *Channel1KafkaPubChannel
		var funcArgs []j.Code
		var newFuncParams []j.Code
		var body []j.Code
		protoChan := ch.AllProtocols[protoName]
		protoChanKafka := protoChan.(*ProtoChannel)

		if len(protoChanKafka.ParameterLinks.Targets()) > 0 {
			funcArgs = append(funcArgs, j.Id("channelName").String())
			newFuncParams = append(newFuncParams, j.Id("channelName"))
		} else {
			body = append(body, j.Const().Id("channelName").Op("=").Lit(protoChanKafka.Name))
		}

		if protoChanKafka.Publisher {
			args := []j.Code{j.Id("channelName")}
			if protoChanKafka.PubChannelBindings != nil {
				funcArgs = append(funcArgs, j.Id("publisherBindings").Qual(ctx.RuntimePackage("kafka"), "ChannelBindings"))
				args = append(args, j.Op("&").Id("publisherBindings"))
			} else {
				args = append(args, j.Nil())
			}
			body = append(body,
				j.Op("pub, err := ").Id(rn).Dot("producer.Publisher").Call(args...),
				j.If(j.Err().Op("!=").Nil().Block(j.Op("return nil, err"))),
			)
			newFuncParams = append(newFuncParams, j.Qual(ctx.RuntimePackage(""), "ToSlice").Call(j.Id("pub")))
		}
		if protoChanKafka.Subscriber {
			var ifBody []j.Code
			if protoChanKafka.Publisher {
				ifBody = append(ifBody, j.Op("err = errors.Join(err, pub.Close())")) // Close publisher
			}
			ifBody = append(ifBody, j.Op("return nil, err"))

			args := []j.Code{j.Id("channelName")}
			if protoChanKafka.PubChannelBindings != nil {
				funcArgs = append(funcArgs, j.Id("subscriberBindings").Qual(ctx.RuntimePackage("kafka"), "ChannelBindings"))
				args = append(args, j.Op("&").Id("subscriberBindings"))
			} else {
				args = append(args, j.Nil())
			}
			body = append(body,
				j.Op("sub, err := ").Id(rn).Dot("consumer.Subscriber").Call(args...),
				j.If(j.Err().Op("!=").Nil().Block(ifBody...)),
			)
			newFuncParams = append(newFuncParams, j.Qual(ctx.RuntimePackage(""), "ToSlice").Call(j.Id("sub")))
		}
		body = append(body, j.Return(j.List(j.Add(utils.ToCode(protoChanKafka.Struct.NewFuncUsage(ctx))...).Call(newFuncParams...), j.Nil())))
		res = append(res,
			j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name+"Channel", true)).
				Params(funcArgs...).
				Params(j.Op("*").Add(utils.ToCode(protoChan.AssembleUsage(ctx))...), j.Error()).
				Block(body...),
		)
	}
	return res
}

func (p ProtoServer) assembleProducerMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Producer").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "Producer").
			Block(
				j.Return(j.Id(rn).Dot("producer")),
			),
	}
}

func (p ProtoServer) assembleConsumerMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Consumer").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "Consumer").
			Block(
				j.Return(j.Id(rn).Dot("consumer")),
			),
	}
}
