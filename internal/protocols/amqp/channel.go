package amqp

import (
	"fmt"
	"time"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type channelBindings struct {
	Is       string          `json:"is" yaml:"is"`
	Exchange *exchangeParams `json:"exchange" yaml:"exchange"`
	Queue    *queueParams    `json:"queue" yaml:"queue"`
}

type exchangeParams struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	Durable    bool   `json:"durable" yaml:"durable"`
	AutoDelete bool   `json:"autoDelete" yaml:"autoDelete"`
	VHost      string `json:"vhost" yaml:"vhost"`
}

type queueParams struct {
	Name       string `json:"name" yaml:"name"`
	Durable    bool   `json:"durable" yaml:"durable"`
	Exclusive  bool   `json:"exclusive" yaml:"exclusive"`
	AutoDelete bool   `json:"autoDelete" yaml:"autoDelete"`
	VHost      string `json:"vhost" yaml:"vhost"`
}

type publishOperationBindings struct {
	Expiration   time.Duration `json:"expiration" yaml:"expiration"`
	UserID       string        `json:"userId" yaml:"userId"`
	CC           []string      `json:"cc" yaml:"cc"`
	Priority     int           `json:"priority" yaml:"priority"`
	DeliveryMode int           `json:"deliveryMode" yaml:"deliveryMode"`
	Mandatory    bool          `json:"mandatory" yaml:"mandatory"`
	BCC          []string      `json:"bcc" yaml:"bcc"`
	ReplyTo      string        `json:"replyTo" yaml:"replyTo"`
	Timestamp    bool          `json:"timestamp" yaml:"timestamp"`
}

type subscribeOperationBindings struct {
	Expiration   time.Duration `json:"expiration" yaml:"expiration"`
	UserID       string        `json:"userID" yaml:"userID"`
	CC           []string      `json:"cc" yaml:"cc"`
	Priority     int           `json:"priority" yaml:"priority"`
	DeliveryMode int           `json:"deliveryMode" yaml:"deliveryMode"`
	ReplyTo      string        `json:"replyTo" yaml:"replyTo"`
	Timestamp    bool          `json:"timestamp" yaml:"timestamp"`
	Ack          bool          `json:"ack" yaml:"ack"`
}

func BuildChannel(ctx *common.CompileContext, channel *compile.Channel, channelKey string) (common.Assembler, error) {
	baseChan, err := protocols.BuildChannel(ctx, channel, channelKey, protoName, protoAbbr)
	if err != nil {
		return nil, err
	}

	baseChan.Struct.Fields = append(baseChan.Struct.Fields, assemble.StructField{Name: "topic", Type: &assemble.Simple{Type: "string"}})

	chanResult := &ProtoChannel{BaseProtoChannel: *baseChan}

	// Channel bindings
	bindingsStruct := &assemble.Struct{ // TODO: remove in favor of parent channel
		BaseType: assemble.BaseType{
			Name:    ctx.GenerateObjName("", "Bindings"),
			Render:  true,
			Package: ctx.TopPackageName(),
		},
	}
	method, chanType, err := buildChannelBindings(ctx, channel, bindingsStruct)
	if err != nil {
		return nil, err
	}
	if method != nil {
		chanResult.BindingsMethod = method
		chanResult.BindingsStructNoAssemble = bindingsStruct
		chanResult.BindingsChannelType = chanType
	}

	return chanResult, nil
}

func buildChannelBindings(ctx *common.CompileContext, channel *compile.Channel, bindingsStruct *assemble.Struct) (res *assemble.Func, chanType string, err error) {
	structValues := &assemble.StructInit{Type: &assemble.Simple{Type: "ChannelBindings", Package: ctx.RuntimePackage(protoName)}}
	var hasBindings bool

	if chBindings, ok := channel.Bindings.Get(protoName); ok {
		hasBindings = true
		var bindings channelBindings
		if err = utils.UnmarshalRawsUnion2(chBindings, &bindings); err != nil {
			return
		}
		switch bindings.Is {
		case "routingKey":
			structValues.Values.Set("ChannelType", &assemble.Simple{Type: "ChannelTypeRoutingKey", Package: ctx.RuntimePackage(protoName)})
		case "queue":
			structValues.Values.Set("ChannelType", &assemble.Simple{Type: "ChannelTypeQueue", Package: ctx.RuntimePackage(protoName)})
		case "":
		default:
			panic(fmt.Sprintf("Unknown channel type %q", bindings.Is))
		}
		chanType = bindings.Is

		if bindings.Exchange != nil {
			ex := &assemble.StructInit{
				Type: &assemble.Simple{Type: "ExchangeConfiguration", Package: ctx.RuntimePackage(protoName)},
			}
			marshalFields := []string{"Name", "Durable", "AutoDelete", "VHost"}
			if err = utils.StructToOrderedMap(*bindings.Exchange, &ex.Values, marshalFields); err != nil {
				return
			}
			switch bindings.Exchange.Type {
			case "default":
				ex.Values.Set("Type", &assemble.Simple{Type: "ExchangeTypeDefault", Package: ctx.RuntimePackage(protoName)})
			case "topic":
				ex.Values.Set("Type", &assemble.Simple{Type: "ExchangeTypeTopic", Package: ctx.RuntimePackage(protoName)})
			case "direct":
				ex.Values.Set("Type", &assemble.Simple{Type: "ExchangeTypeDirect", Package: ctx.RuntimePackage(protoName)})
			case "fanout":
				ex.Values.Set("Type", &assemble.Simple{Type: "ExchangeTypeFanout", Package: ctx.RuntimePackage(protoName)})
			case "headers":
				ex.Values.Set("Type", &assemble.Simple{Type: "ExchangeTypeHeaders", Package: ctx.RuntimePackage(protoName)})
			case "":
			default:
				panic(fmt.Sprintf("Unknown exchange type %q", bindings.Is))
			}
			structValues.Values.Set("ExchangeConfiguration", ex)
		}
		if bindings.Queue != nil {
			ex := &assemble.StructInit{
				Type: &assemble.Simple{Type: "QueueConfiguration", Package: ctx.RuntimePackage(protoName)},
			}
			marshalFields := []string{"Name", "Durable", "Exclusive", "AutoDelete", "VHost"}
			if err = utils.StructToOrderedMap(*bindings.Exchange, &ex.Values, marshalFields); err != nil {
				return
			}
			structValues.Values.Set("QueueConfiguration", ex)
		}
	}

	// Publish channel bindings
	if channel.Publish != nil {
		if b, ok := channel.Publish.Bindings.Get(protoName); ok {
			pob := &assemble.StructInit{
				Type: &assemble.Simple{Type: "PublishOperationBindings", Package: ctx.RuntimePackage(protoName)},
			}
			hasBindings = true
			var bindings publishOperationBindings
			if err = utils.UnmarshalRawsUnion2(b, &bindings); err != nil {
				return
			}
			marshalFields := []string{"Expiration", "UserID", "CC", "Priority", "Mandatory", "BCC", "ReplyTo", "Timestamp"}
			if err = utils.StructToOrderedMap(bindings, &pob.Values, marshalFields); err != nil {
				return
			}
			switch bindings.DeliveryMode {
			case 1:
				pob.Values.Set("DeliveryMode", &assemble.Simple{Type: "DeliveryModeTransient", Package: ctx.RuntimePackage(protoName)})
			case 2:
				pob.Values.Set("DeliveryMode", &assemble.Simple{Type: "DeliveryModePersistent", Package: ctx.RuntimePackage(protoName)})
			case 0:
			default:
				panic(fmt.Sprintf("Unknown delivery mode %v", bindings.DeliveryMode))
			}

			structValues.Values.Set("PublisherBindings", pob)
		}
	}

	// Subscribe channel bindings
	if channel.Subscribe != nil {
		if b, ok := channel.Subscribe.Bindings.Get(protoName); ok {
			sob := &assemble.StructInit{
				Type: &assemble.Simple{Type: "SubscribeOperationBindings", Package: ctx.RuntimePackage(protoName)},
			}
			hasBindings = true
			var bindings subscribeOperationBindings
			if err = utils.UnmarshalRawsUnion2(b, &bindings); err != nil {
				return
			}
			marshalFields := []string{"Expiration", "UserID", "CC", "Priority", "ReplyTo", "Timestamp", "Ack"}
			if err = utils.StructToOrderedMap(bindings, &sob.Values, marshalFields); err != nil {
				return
			}
			switch bindings.DeliveryMode {
			case 1:
				sob.Values.Set("DeliveryMode", &assemble.Simple{Type: "DeliveryModeTransient", Package: ctx.RuntimePackage(protoName)})
			case 2:
				sob.Values.Set("DeliveryMode", &assemble.Simple{Type: "DeliveryModePersistent", Package: ctx.RuntimePackage(protoName)})
			case 0:
			default:
				panic(fmt.Sprintf("Unknown delivery mode %v", bindings.DeliveryMode))
			}

			structValues.Values.Set("SubscriberBindings", sob)
		}
	}

	if !hasBindings {
		return nil, "", nil
	}

	// Method Proto() proto.ChannelBindings
	res = &assemble.Func{
		FuncSignature: assemble.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: assemble.Simple{Type: "ChannelBindings", Package: ctx.RuntimePackage(protoName)}},
			},
		},
		Receiver:      bindingsStruct,
		Package:       ctx.TopPackageName(),
		BodyAssembler: protocols.ChannelBindingsMethodBody(structValues, nil, nil),
	}

	return
}

type ProtoChannel struct {
	protocols.BaseProtoChannel
	BindingsStructNoAssemble *assemble.Struct // nil if bindings not set FIXME: remove in favor of struct in parent channel
	BindingsMethod           *assemble.Func
	BindingsChannelType      string
}

func (p ProtoChannel) AllowRender() bool {
	return true
}

func (p ProtoChannel) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	if p.BindingsMethod != nil {
		res = append(res, p.BindingsMethod.AssembleDefinition(ctx)...)
	}
	res = append(res, p.ServerIface.AssembleDefinition(ctx)...)
	res = append(res, protocols.AssembleChannelOpenFunc(
		ctx, p.Struct, p.Name, p.ServerIface, p.ParametersStructNoAssemble, p.BindingsStructNoAssemble,
		p.Publisher, p.Subscriber, protoName, protoAbbr,
	)...)
	res = append(res, p.assembleNewFunc(ctx)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, protocols.AssembleChannelCommonMethods(ctx, p.Struct, p.Publisher, p.Subscriber, protoAbbr)...)
	res = append(res, p.assembleCommonMethods(ctx)...)
	if p.Publisher {
		res = append(res, protocols.AssembleChannelPublisherMethods(ctx, p.Struct, protoName)...)
		res = append(res, p.assemblePublisherMethods(ctx)...)
	}
	if p.Subscriber {
		res = append(res, protocols.AssembleChannelSubscriberMethods(
			ctx, p.Struct, p.SubMessageLink, p.FallbackMessageType, protoName, protoAbbr,
		)...)
	}
	return res
}

func (p ProtoChannel) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoChannel) assembleNewFunc(ctx *common.AssembleContext) []*j.Statement {
	return []*j.Statement{
		// NewChannel1Proto(params Channel1Parameters, publisher proto.Publisher, subscriber proto.Subscriber) *Channel1Proto
		j.Func().Id(p.Struct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				if p.ParametersStructNoAssemble != nil {
					g.Id("params").Add(utils.ToCode(p.ParametersStructNoAssemble.AssembleUsage(ctx))...)
				}
				if p.Publisher {
					g.Id("publisher").Qual(ctx.RuntimePackage(protoName), "Publisher")
				}
				if p.Subscriber {
					g.Id("subscriber").Qual(ctx.RuntimePackage(protoName), "Subscriber")
				}
			}).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			BlockFunc(func(bg *j.Group) {
				bg.Op("res := ").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
					d[j.Id("name")] = j.Id(utils.ToGolangName(p.Name, true) + "Name").CallFunc(func(g *j.Group) {
						if p.ParametersStructNoAssemble != nil {
							g.Id("params")
						}
					})
					if p.Publisher {
						d[j.Id("publisher")] = j.Id("publisher")
					}
					if p.Subscriber {
						d[j.Id("subscriber")] = j.Id("subscriber")
					}
				}))

				switch p.BindingsChannelType {
				case "routingKey", "":
					bg.Op("res.exchange = res.name.String()")
					if p.BindingsStructNoAssemble != nil {
						bg.Id("bindings").Op(":=").Add(utils.ToCode(p.BindingsStructNoAssemble.AssembleUsage(ctx))...).Values().Dot(protoAbbr).Call()
						bg.Op(`
							if bindings.ExchangeConfiguration.Name != "" {
								res.exchange = bindings.ExchangeConfiguration.Name
							}
							res.queue = bindings.QueueConfiguration.Name`)
					}
				case "queue":
					bg.Op("res.queue = res.name.String()")
					if p.BindingsStructNoAssemble != nil {
						bg.Id("bindings").Op(":=").Add(utils.ToCode(p.BindingsStructNoAssemble.AssembleUsage(ctx))...).Values().Dot(protoAbbr).Call()
						bg.Op(`
							if bindings.QueueConfiguration.Name != "" {
								res.queue = bindings.QueueConfiguration.Name
							}
							res.exchange = bindings.ExchangeConfiguration.Name`)
					}
				default:
					panic(fmt.Sprintf("Unknown channel type: %q", p.BindingsChannelType))
				}
				bg.Op(`return &res`)
			}),
	}
}

func (p ProtoChannel) assembleCommonMethods(_ *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	return []*j.Statement{
		// Method Exchange() string
		j.Func().Params(receiver.Clone()).Id("Exchange").
			Params().
			String().
			Block(
				j.Return(j.Id(rn).Dot("exchange")),
			),

		// Method Queue() string
		j.Func().Params(receiver.Clone()).Id("Queue").
			Params().
			String().
			Block(
				j.Return(j.Id(rn).Dot("queue")),
			),
	}
}

func (p ProtoChannel) assemblePublisherMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	var msgTyp common.GolangType = assemble.NullableType{Type: p.FallbackMessageType, Render: true}
	if p.PubMessageLink != nil {
		msgTyp = assemble.NullableType{Type: p.PubMessageLink.Target().OutStruct, Render: true}
	}

	var msgBindings *assemble.Struct
	if p.PubMessageLink != nil && p.PubMessageLink.Target().BindingsStruct != nil {
		msgBindings = p.PubMessageLink.Target().BindingsStruct
	}

	return []*j.Statement{
		// Method MakeEnvelope(envelope kafka.EnvelopeWriter, message *Message1Out) error
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			ParamsFunc(func(g *j.Group) {
				g.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter")
				g.Id("message").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...)
			}).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("envelope.ResetPayload()")
				if p.PubMessageLink == nil {
					bg.Empty().Add(utils.QualSprintf(`
						enc := %Q(encoding/json,NewEncoder)(envelope)
						if err := enc.Encode(message); err != nil {
							return err
						}`))
				} else {
					bg.Op(`
						if err := message.MarshalAMQPEnvelope(envelope); err != nil {
							return err
						}`)
				}
				bg.Op("envelope.SetExchange").Call(j.Id(rn).Dot("exchange"))
				bg.Op("envelope.SetQueue").Call(j.Id(rn).Dot("queue"))
				if msgBindings != nil {
					bg.Op("envelope.SetBindings").Call(
						j.Add(utils.ToCode(msgBindings.AssembleUsage(ctx))...).Values().Dot("AMQP()"),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}
