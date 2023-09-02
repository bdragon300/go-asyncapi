package kafka

import (
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

const protoName = "kafka"

type ProtoChannel struct {
	Name             string
	Topic            string
	Struct           *assemble.Struct
	Message          common.Assembler
	MessageHasSchema bool
}

func (p ProtoChannel) AllowRender() bool {
	return true
}

func (p ProtoChannel) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoChannel) commonMethods(ctx *common.AssembleContext, clientField string) []*j.Statement {
	structName := p.Struct.Name
	messageType := utils.CastSliceItems[*j.Statement, j.Code](p.Message.AssembleUsage(ctx))
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	return []*j.Statement{
		// Method Name() -> string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().Block(
			j.Return(j.Lit(p.Name)),
		),
		// Method Topic() -> string
		j.Func().Params(receiver.Clone()).Id("Topic").
			Params().
			String().Block(
			j.Return(j.Lit(p.Topic)),
		),
		// Method MakeEnvelope(message MessageType, key []byte) -> *KafkaOutEnvelope
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			Params(j.Id("message").Add(messageType...), j.Id("key").Index().Byte()).
			Params(j.Op("*").Id("KafkaOutEnvelope"), j.Error()).Block(
			j.List(j.Id("payload"), j.Err()).Op(":=").Qual("encoding/json", "Marshal").Call(j.Id("message")),
			j.If(j.Err().Op("!=").Nil()).Block(
				j.Return(j.List(j.Nil(), j.Err())),
			),
			j.Return(j.List(j.Op("&").Qual(ctx.RuntimePackage(), "KafkaOutEnvelope").Values(j.Dict{
				j.Id("KafkaMeta"): j.Id("KafkaMeta").Values(j.Dict{
					j.Id("Key"):       j.Id("key"),
					j.Id("Topic"):     j.Id(receiverName).Dot("Topic").Call(),
					j.Id("Partition"): j.Nil(),
				}),
				j.Id("Payload"): j.Id("payload"),
				j.Id("Headers"): j.Qual(ctx.RuntimePackage(), "StructToMapByte").Call(j.Id("message").Dot("Headers")),
				j.Id("To"):      j.Id(receiverName).Dot(clientField),
			}), j.Nil())),
		),
	}
}

type ProtoChannelSub struct {
	ProtoChannel
}

func (p ProtoChannelSub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	res := p.Struct.AssembleDefinition(ctx)
	res = append(res, p.assembleMethods(ctx)...)
	return res
}

func (p ProtoChannelSub) assembleMethods(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	var res []*j.Statement
	res = append(res, p.ProtoChannel.commonMethods(ctx, "consumer")...)
	publishMethod := j.Func().Params(receiver.Clone()).Id("Subscribe").
		Params(j.Id("ctx").Qual("context", "Context"), j.Id("cb").Qual(ctx.RuntimePackage(), "KafkaConsumerCallback")).
		Error().Block(
		j.Return(j.Id(receiverName).Dot("consumer").Dot("Consume").Call(j.Id("ctx"), j.Id("cb"))),
	)

	return append(res, publishMethod)
}

type ProtoChannelPub struct {
	ProtoChannel
}

func (p ProtoChannelPub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	res := p.Struct.AssembleDefinition(ctx)
	res = append(res, p.assembleMethods(ctx)...)
	return res
}

func (p ProtoChannelPub) assembleMethods(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	var res []*j.Statement
	res = append(res, p.ProtoChannel.commonMethods(ctx, "producers")...)
	publishMethod := j.Func().Params(receiver.Clone()).Id("Publish").
		Params(j.Id("ctx").Qual("context", "Context"), j.Id("envelope").Op("*").Qual(ctx.RuntimePackage(), "KafkaOutEnvelope")).
		Error().Block(
		j.Id("p").Op(":=").Op("&").Qual(ctx.RuntimePackage(), "ErrorPool").Values(),
		j.Op(`
for i := 0; i < len(envelope.To); i++ {
	i := i
	p.Go(func() error {
		return envelope.To[i].Produce(ctx, envelope)
	})
}
return p.Wait(ctx)`),
	)

	return append(res, publishMethod)
}

type ProtoServer struct {
	Name          string
	Struct        *assemble.Struct
	ChannelsLinks *assemble.LinkQueryList[*assemble.Channel]
}

func (p ProtoServer) AllowRender() bool {
	return true
}

func (p ProtoServer) commonMethods() []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	return []*j.Statement{
		// Method Name() -> string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().Block(
			j.Return(j.Lit(p.Name)),
		),
	}
}

type ProtoServerSub struct {
	ProtoServer
}

func (p ProtoServerSub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	res := p.Struct.AssembleDefinition(ctx)
	res = append(res, p.ProtoServer.commonMethods()...)
	for _, ch := range p.ChannelsLinks.Links() {
		subscriber := ch.SupportedProtocols[protoName].Subscribe
		if subscriber == nil {
			continue // Channel has not subscriber
		}
		chanTyp := utils.CastSliceItems[*j.Statement, j.Code](subscriber.AssembleUsage(ctx))
		stmt := j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name + "SubChannel")).
			Params().
			Op("*").Add(chanTyp...).Block(
			j.Return(j.Op("&").Add(chanTyp...).Values(j.Id(receiverName))),
		)
		res = append(res, stmt)
	}
	return res
}

func (p ProtoServerSub) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

type ProtoServerPub struct {
	ProtoServer
}

func (p ProtoServerPub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	res := p.Struct.AssembleDefinition(ctx)
	res = append(res, p.ProtoServer.commonMethods()...)
	for _, ch := range p.ChannelsLinks.Links() {
		publisher := ch.SupportedProtocols[protoName].Publish
		if publisher == nil {
			continue // Channel has not publisher
		}
		chanTyp := utils.CastSliceItems[*j.Statement, j.Code](publisher.AssembleUsage(ctx))
		stmt := j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name + "PubChannel")).
			Params().
			Op("*").Add(chanTyp...).Block(
			j.Return(j.Op("&").Add(chanTyp...).Values(j.Id(receiverName))),
		)
		res = append(res, stmt)
	}
	return res
}

func (p ProtoServerPub) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

type ProtoServerCommon struct {
	ProtoServer
}

func (p ProtoServerCommon) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	res := p.Struct.AssembleDefinition(ctx)
	for _, ch := range p.ChannelsLinks.Links() {
		chanTyp := utils.CastSliceItems[*j.Statement, j.Code](ch.SupportedProtocols[protoName].Common.AssembleUsage(ctx))
		stmt := j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name + "Channel")).
			Params().
			Op("*").Add(chanTyp...).Block(
			j.Return(j.Op("&").Add(chanTyp...).Values(j.Id(receiverName))),
		)
		res = append(res, stmt)
	}
	return res
}

func (p ProtoServerCommon) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}
