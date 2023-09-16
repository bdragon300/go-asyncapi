package kafka

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type ProtoServer struct {
	Name            string
	Struct          *assemble.Struct
	ChannelLinkList *assemble.LinkList[*assemble.Channel]
	Producer        bool
	Consumer        bool
}

func (p ProtoServer) AllowRender() bool {
	return true
}

func (p ProtoServer) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
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
		protoChan := ch.SupportedProtocols[protoName]
		protoChanKafka := protoChan.(*ProtoChannel)

		if protoChanKafka.Publisher {
			params = append(params, j.Id("publisherParams").Qual(ctx.RuntimePackage("kafka"), "ChannelParams"))
			body = append(body,
				j.Op("pub, err := ").Id(rn).Dot("producer.Publisher(publisherParams)"),
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

			params = append(params, j.Id("subscriberParams").Qual(ctx.RuntimePackage("kafka"), "ChannelParams"))
			body = append(body,
				j.Op("sub, err := ").Id(rn).Dot("consumer.Subscriber(subscriberParams)"),
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
