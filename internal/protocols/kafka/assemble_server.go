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

type ProtoServer struct {
	Name            string
	Struct          *assemble.Struct
	ChannelLinkList *assemble.LinkList[*assemble.Channel]
	Producer        bool
	Consumer        bool
	Bindings        *ProtoServerBindings
}

func (p ProtoServer) AllowRender() bool {
	return true
}

func (p ProtoServer) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	if p.Bindings != nil {
		res = append(res, p.assembleBindings(ctx, p.Bindings)...)
	}
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
		var params []j.Code
		var newFuncParams []j.Code
		var body []j.Code
		protoChan := ch.AllProtocols[protoName]
		protoChanKafka := protoChan.(*ProtoChannel)

		// TODO: bindings are optional
		if protoChanKafka.Publisher {
			callArg := j.Nil()
			if protoChanKafka.PubChannelBindings != nil {
				params = append(params, j.Id("publisherBindings").Qual(ctx.RuntimePackage("kafka"), "ChannelBindings"))
				callArg = j.Op("&").Id("publisherBindings")
			}
			body = append(body,
				j.Op("pub, err := ").Id(rn).Dot("producer.Publisher").Call(callArg),
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

			callArg := j.Nil()
			if protoChanKafka.PubChannelBindings != nil {
				params = append(params, j.Id("subscriberBindings").Qual(ctx.RuntimePackage("kafka"), "ChannelBindings"))
				callArg = j.Op("&").Id("subscriberBindings")
			}
			body = append(body,
				j.Op("sub, err := ").Id(rn).Dot("consumer.Subscriber").Call(callArg),
				j.If(j.Err().Op("!=").Nil().Block(ifBody...)),
			)
			newFuncParams = append(newFuncParams, j.Qual(ctx.RuntimePackage(""), "ToSlice").Call(j.Id("sub")))
		}
		body = append(body, j.Return(j.List(j.Add(utils.ToCode(protoChanKafka.Struct.NewFuncUsage(ctx))...).Call(newFuncParams...), j.Nil())))
		res = append(res,
			j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name+"Channel", true)).
				Params(params...).
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
