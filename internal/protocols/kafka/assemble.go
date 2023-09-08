package kafka

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type ProtoChannel struct {
	Name   string
	Topic  string
	Iface  *assemble.Interface
	Struct *assemble.Struct

	MessageLink *assemble.Link[*assemble.Message] // nil when message is not set
}

func (p ProtoChannel) AllowRender() bool {
	return true
}

func (p ProtoChannel) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoChannel) commonMethods(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	return []*j.Statement{
		// Method Name() -> string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Name)),
			),

		// Method Topic() -> string
		j.Func().Params(receiver.Clone()).Id("Topic").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Topic)),
			),

		// Method ChannelParams() KafkaChannelParams
		j.Func().Params(receiver.Clone()).Id("ChannelParams").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "ChannelParams").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage("kafka"), "ChannelParams").Values(j.Dict{
					j.Id("Topic"): j.Id(receiverName).Dot("Topic").Call(),
				})),
			),
	}
}

type ProtoChannelSub struct {
	ProtoChannel
}

func (p ProtoChannelSub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	res := p.Iface.AssembleDefinition(ctx)

	// NewSubChannel1(servers ...SubChannel1Server) *SubChannel1
	allocVals := j.Dict{j.Id("servers"): j.Id("servers")}
	res = append(res,
		j.Func().Id(p.Struct.NewFuncName()).
			Params(j.Id("servers").Op("...").Add(utils.ToCode(p.Iface.AssembleUsage(ctx))...)).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(allocVals)),
			),
	)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.assembleMethods(ctx)...)
	return res
}

func (p ProtoChannelSub) assembleMethods(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	messageType := []j.Code{j.Any()}
	headersType := []j.Code{j.Any()}
	if p.MessageLink != nil {
		messageType = utils.ToCode(p.MessageLink.Link().AssembleUsage(ctx))
		headersType = utils.ToCode(p.MessageLink.Link().HeadersType.AssembleUsage(ctx))
	}
	rcvName := strings.ToLower(string(structName[0]))
	receiver := j.Id(rcvName).Op("*").Id(structName)
	kafkaPkg := ctx.RuntimePackage("kafka")

	res := p.ProtoChannel.commonMethods(ctx)
	methods := []*j.Statement{
		// Method ExtractEnvelope(envelope *kafka.InEnvelope, msg *Message1) error
		j.Func().Params(receiver.Clone()).Id("ExtractEnvelope").
			Params(
				j.Id("envelope").Op("*").Qual(kafkaPkg, "InEnvelope"),
				j.Id("msg").Op("*").Add(messageType...),
			).
			Error().
			Block(
				j.If(
					j.Err().Op(":=").Qual("encoding/json", "Unmarshal").Call(j.Op("envelope.Payload"), j.Op("&msg.Payload")),
					j.Err().Op("!=").Nil(),
				).Block(j.Return(j.Err())),
				j.Op("msg.Headers =").Op("*").New(j.Add(headersType...)),
				j.Op("msg.ID = \"\""), // TODO
				j.Return(j.Nil()),
			),

		// Method Subscribe(ctx context.Context) (<-chan *Message1, <-chan error, error)
		j.Func().Params(receiver.Clone()).Id("Subscribe").
			Params(j.Id("ctx").Qual("context", "Context")).
			Params(
				j.Op("<-chan *").Add(messageType...),
				j.Op("<-chan error"),
				j.Error(),
			).
			Block(
				j.Op(fmt.Sprintf(`
					rcv, err := %[1]s.Receive(ctx)
					if err != nil {
						return nil, nil, err
					}`, rcvName),
				),
				j.Op("resCh, errCh :=").Qual(ctx.RuntimePackage(""), "StreamBy").Call(
					j.Lit(16), // FIXME
					j.Id("rcv"),
					j.Func().Params(j.Id("item").Op("*").Qual(ctx.RuntimePackage("kafka"), "InEnvelope")).Params(j.Op("*").Add(messageType...), j.Error()).Block(
						j.Op("buf :=").New(j.Add(messageType...)),
						j.Op(fmt.Sprintf(`
							if err := %[1]s.ExtractEnvelope(item, buf); err != nil {
								return nil, err
							}
							return buf, nil`, rcvName),
						),
					),
				),
				j.Return(j.Op("resCh, errCh, nil")),
			),

		// Method Receive(ctx context.Context) (<-chan *kafka.InEnvelope, error)
		j.Func().Params(receiver.Clone()).Id("Receive").
			Params(j.Id("ctx").Qual("context", "Context")).
			Params(
				j.Op("<-chan *").Qual(ctx.RuntimePackage("kafka"), "InEnvelope"),
				j.Error(),
			).
			Block(
				j.Op("var chans []<-chan *").Qual(ctx.RuntimePackage("kafka"), "InEnvelope"),
				j.Op(fmt.Sprintf(`
					for i := 0; i < len(%[1]s.servers); i++ {
						ch, err := %[1]s.servers[i].Consumer().Consume(ctx, %[1]s.ChannelParams())
						if err != nil {
							return nil, err
						}
						chans = append(chans, ch)
					}`, rcvName),
				),
				j.Return(j.Qual(ctx.RuntimePackage(""), "FanIn").Call(j.Op("16, chans...")), j.Nil()), // TODO: buffer capacity
			),
	}

	return append(res, methods...)
}

type ProtoChannelPub struct {
	ProtoChannel
}

func (p ProtoChannelPub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	res := p.Iface.AssembleDefinition(ctx)

	// NewPubChannel1(servers ...PubChannel1Server) *PubChannel1
	allocVals := j.Dict{j.Id("servers"): j.Id("servers")}
	res = append(res,
		j.Func().Id(p.Struct.NewFuncName()).
			Params(j.Id("servers").Op("...").Add(utils.ToCode(p.Iface.AssembleUsage(ctx))...)).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(allocVals)),
			),
	)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.assembleMethods(ctx)...)
	return res
}

func (p ProtoChannelPub) assembleMethods(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	messageType := utils.ToCode(p.MessageLink.Link().AssembleUsage(ctx))
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Id(structName)
	kafkaPkg := ctx.RuntimePackage("kafka")

	res := p.ProtoChannel.commonMethods(ctx)
	methods := []*j.Statement{
		// Method MakeEnvelope(message MessageType, key []byte) *KafkaOutEnvelope
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			Params(
				j.Id("envelope").Op("*").Qual(kafkaPkg, "OutEnvelope"),
				j.Id("payload").Index().Byte(),
			).
			Block(
				j.Op("envelope.Payload = payload"),
				j.Op("envelope.Headers = nil"), // TODO
				j.Op(fmt.Sprintf("envelope.Metadata.Topic = %[1]s.Topic()", receiverName)),
				j.Op("envelope.Metadata.Partition = -1"),
			),

		// Method Publish(ctx context.Context, messages ...Message1) error
		j.Func().Params(receiver.Clone()).Id("Publish").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("messages").Op("...").Add(messageType...),
			).
			Error().
			Block(
				j.Id("res").Op(":=").Make(j.Index().Qual(kafkaPkg, "OutEnvelope"), j.Lit(0), j.Len(j.Id("messages"))),
				j.Op("for i, msg := range messages").Block(
					j.Op("buf :=").Qual(kafkaPkg, "OutEnvelope").Values(),
					j.Op("payload, err := ").Qual("encoding/json", "Marshal").Call(j.Id("msg")),
					j.If(j.Err().Op("!=").Nil()).Block(
						j.Return(j.Qual("fmt", "Errorf").Call(j.Lit("unable to marshal msg #%d: %w"), j.Id("i"), j.Err())),
					),
					j.Op(fmt.Sprintf("%[1]s.MakeEnvelope(&buf, payload)", receiverName)),
					j.Op("res = append(res, buf)"),
				),
				j.Return(j.Op(fmt.Sprintf("%[1]s.Send(ctx, res)", receiverName))),
			),

		// Method Send(ctx context.Context, envelopes []kafka.OutEnvelope) error
		j.Func().Params(receiver.Clone()).Id("Send").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("envelopes").Index().Qual(kafkaPkg, "OutEnvelope"),
			).
			Error().
			Block(
				j.Op("p :=").Qual(ctx.RuntimePackage(""), "NewErrorPool").Call(),
				j.Op(fmt.Sprintf(`
					for i := 0; i < len(%[1]s.servers); i++ {
						i := i
						p.Go(func() error {
							return %[1]s.servers[i].Producer().Produce(ctx, %[1]s.ChannelParams(), envelopes)
						})
					}
					return p.Wait()`, receiverName),
				),
			),
	}

	return append(res, methods...)
}

type ProtoChannelCommon struct {
	assemble.Struct
}

type ProtoServer struct {
	Name          string
	Struct        *assemble.Struct
	ChannelsLinks *assemble.LinkList[*assemble.Channel]
}

func (p ProtoServer) AllowRender() bool {
	return true
}

func (p ProtoServer) commonMethods() []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Id(structName)

	return []*j.Statement{
		// Method Name() -> string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Name)),
			),
	}
}

func (p ProtoServer) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

type ProtoServerSub struct {
	ProtoServer
}

func (p ProtoServerSub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Id(structName)
	var res []*j.Statement

	// NewServer1SubServer(consumer kafka.Consumer) *Server1SubServer
	allocVals := j.Dict{j.Id("consumer"): j.Id("consumer")}
	res = append(res,
		j.Func().Id(p.Struct.NewFuncName()).
			Params(j.Id("consumer").Qual(ctx.RuntimePackage("kafka"), "Consumer")).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(allocVals)),
			),
	)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.ProtoServer.commonMethods()...)
	// Method Consumer() kafka.Consumer
	res = append(res,
		j.Func().Params(receiver.Clone()).Id("Consumer").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "Consumer").
			Block(
				j.Return(j.Id(receiverName).Dot("consumer")),
			),
	)
	for _, ch := range p.ChannelsLinks.Links() {
		subscriber := ch.SupportedProtocols[protoName].Subscribe
		if subscriber == nil {
			continue // Channel has not subscriber
		}
		subscriberKafka := subscriber.(*ProtoChannelSub)

		res = append(res,
			// Method Channel1SubChannel() *Channel1KafkaSubChannel
			j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name+"SubChannel", true)).
				Params().
				Op("*").Add(utils.ToCode(subscriber.AssembleUsage(ctx))...).
				Block(
					j.Return(j.Add(utils.ToCode(subscriberKafka.Struct.NewFuncUsage(ctx))...).Call(j.Op("&").Id(receiverName))),
				),
		)
	}
	return res
}

type ProtoServerPub struct {
	ProtoServer
}

func (p ProtoServerPub) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Id(structName)
	var res []*j.Statement

	// NewServer1PubServer(producer kafka.Producer) *Server1PubServer
	allocVals := j.Dict{j.Id("producer"): j.Id("producer")}
	res = append(res,
		j.Func().Id(p.Struct.NewFuncName()).
			Params(j.Id("producer").Qual(ctx.RuntimePackage("kafka"), "Producer")).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(allocVals)),
			),
	)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.ProtoServer.commonMethods()...)

	// Method Consumer() kafka.Consumer
	res = append(res,
		j.Func().Params(receiver.Clone()).Id("Producer").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "Producer").
			Block(
				j.Return(j.Id(receiverName).Dot("producer")),
			),
	)

	for _, ch := range p.ChannelsLinks.Links() {
		publisher := ch.SupportedProtocols[protoName].Publish
		if publisher == nil {
			continue // Channel has not publisher
		}
		publisherKafka := publisher.(*ProtoChannelPub)
		// Method Channel1PubChannel() *Channel1KafkaPubChannel
		res = append(res,
			j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name+"PubChannel", true)).
				Params().
				Op("*").Add(utils.ToCode(publisher.AssembleUsage(ctx))...).
				Block(
					j.Return(j.Add(utils.ToCode(publisherKafka.Struct.NewFuncUsage(ctx))...).Call(j.Op("&").Id(receiverName))),
				),
		)
	}
	return res
}

type ProtoServerCommon struct {
	ProtoServer
	PubStruct *assemble.Struct
	SubStruct *assemble.Struct
}

func (p ProtoServerCommon) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	structName := p.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Id(structName)
	var res []*j.Statement

	// NewServer1Server(producer, consumer) *Server1Server
	allocVals := j.Dict{
		j.Id(p.PubStruct.Name): j.Add(utils.ToCode(p.PubStruct.AssembleUsage(ctx))...).Values(j.Dict{j.Id("producer"): j.Id("producer")}),
		j.Id(p.SubStruct.Name): j.Add(utils.ToCode(p.SubStruct.AssembleUsage(ctx))...).Values(j.Dict{j.Id("consumer"): j.Id("consumer")}),
	}
	res = append(res,
		j.Func().Id(p.Struct.NewFuncName()).
			Params(
				j.Id("producer").Qual(ctx.RuntimePackage("kafka"), "Producer"),
				j.Id("consumer").Qual(ctx.RuntimePackage("kafka"), "Consumer"),
			).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(allocVals)),
			),
	)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)

	for _, ch := range p.ChannelsLinks.Links() {
		commonChan := ch.SupportedProtocols[protoName].Common
		chanKafka := commonChan.(*ProtoChannelCommon)
		vals := j.Dict{}
		if ch.SupportedProtocols[protoName].Publish != nil {
			kafkaID := j.Id(utils.ToGolangName(ch.Name+"KafkaPubChannel", true))
			vals[kafkaID] = j.Op("*").Id(receiverName).Dot(utils.ToGolangName(ch.Name+"PubChannel", true)).Call()
		}
		if ch.SupportedProtocols[protoName].Subscribe != nil {
			kafkaID := j.Id(utils.ToGolangName(ch.Name+"KafkaSubChannel", true))
			vals[kafkaID] = j.Op("*").Id(receiverName).Dot(utils.ToGolangName(ch.Name+"SubChannel", true)).Call()
		}

		// Method Channel1Channel() *Channel1KafkaChannel
		stmt := j.Func().Params(receiver.Clone()).Id(utils.ToGolangName(ch.Name+"Channel", true)).
			Params().
			Op("*").Add(utils.ToCode(chanKafka.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(chanKafka.Struct.AssembleUsage(ctx))...).Values(vals)),
			)
		res = append(res, stmt)
	}
	return res
}
