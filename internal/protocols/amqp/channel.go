package amqp

import (
	"fmt"
	"time"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/compile"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
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

func BuildChannel(ctx *common.CompileContext, channel *compile.Channel, channelKey string) (common.Renderer, error) {
	baseChan, err := protocols.BuildChannel(ctx, channel, channelKey, ProtoName, protoAbbr)
	if err != nil {
		return nil, err
	}

	baseChan.Struct.Fields = append(
		baseChan.Struct.Fields,
		render.StructField{Name: "exchange", Type: &render.Simple{Name: "string"}},
		render.StructField{Name: "queue", Type: &render.Simple{Name: "string"}},
	)

	chanResult := &ProtoChannel{BaseProtoChannel: *baseChan}

	// Channel bindings
	bindingsStruct := &render.Struct{ // TODO: remove in favor of parent channel
		BaseType: render.BaseType{
			Name:         ctx.GenerateObjName(channelKey, "Bindings"),
			DirectRender: true,
			PackageName:  ctx.TopPackageName(),
		},
	}
	method, chanType, err := buildChannelBindings(ctx, channel, bindingsStruct)
	if err != nil {
		return nil, err
	}
	if method != nil {
		chanResult.BindingsMethod = method
		chanResult.BindingsStructNoRender = bindingsStruct
		chanResult.BindingsChannelType = chanType
	}

	return chanResult, nil
}

func buildChannelBindings(ctx *common.CompileContext, channel *compile.Channel, bindingsStruct *render.Struct) (*render.Func, string, error) {
	structValues := &render.StructInit{Type: &render.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}}
	var hasBindings bool
	var chanType string

	if chBindings, ok := channel.Bindings.Get(ProtoName); ok {
		ctx.Logger.Trace("Channel bindings", "proto", ProtoName)
		hasBindings = true
		var bindings channelBindings
		if err := utils.UnmarshalRawsUnion2(chBindings, &bindings); err != nil {
			return nil, "", common.CompileError{Err: err, Path: ctx.PathRef(), Proto: ProtoName}
		}
		switch bindings.Is {
		case "routingKey":
			structValues.Values.Set("ChannelType", &render.Simple{Name: "ChannelTypeRoutingKey", Package: ctx.RuntimePackage(ProtoName)})
		case "queue":
			structValues.Values.Set("ChannelType", &render.Simple{Name: "ChannelTypeQueue", Package: ctx.RuntimePackage(ProtoName)})
		case "":
		default:
			return nil, "", common.CompileError{Err: fmt.Errorf("unknown channel type %q", bindings.Is), Path: ctx.PathRef(), Proto: ProtoName}
		}
		chanType = bindings.Is

		if bindings.Exchange != nil {
			ex := &render.StructInit{
				Type: &render.Simple{Name: "ExchangeConfiguration", Package: ctx.RuntimePackage(ProtoName)},
			}
			marshalFields := []string{"Name", "Durable", "AutoDelete", "VHost"}
			if err := utils.StructToOrderedMap(*bindings.Exchange, &ex.Values, marshalFields); err != nil {
				return nil, "", common.CompileError{Err: err, Path: ctx.PathRef(), Proto: ProtoName}
			}
			switch bindings.Exchange.Type {
			case "default":
				ex.Values.Set("Type", &render.Simple{Name: "ExchangeTypeDefault", Package: ctx.RuntimePackage(ProtoName)})
			case "topic":
				ex.Values.Set("Type", &render.Simple{Name: "ExchangeTypeTopic", Package: ctx.RuntimePackage(ProtoName)})
			case "direct":
				ex.Values.Set("Type", &render.Simple{Name: "ExchangeTypeDirect", Package: ctx.RuntimePackage(ProtoName)})
			case "fanout":
				ex.Values.Set("Type", &render.Simple{Name: "ExchangeTypeFanout", Package: ctx.RuntimePackage(ProtoName)})
			case "headers":
				ex.Values.Set("Type", &render.Simple{Name: "ExchangeTypeHeaders", Package: ctx.RuntimePackage(ProtoName)})
			case "":
			default:
				return nil, "", common.CompileError{
					Err:   fmt.Errorf("unknown exchange type %q", bindings.Exchange.Type),
					Path:  ctx.PathRef(),
					Proto: ProtoName,
				}
			}
			structValues.Values.Set("ExchangeConfiguration", ex)
		}
		if bindings.Queue != nil {
			ex := &render.StructInit{
				Type: &render.Simple{Name: "QueueConfiguration", Package: ctx.RuntimePackage(ProtoName)},
			}
			marshalFields := []string{"Name", "Durable", "Exclusive", "AutoDelete", "VHost"}
			if err := utils.StructToOrderedMap(*bindings.Exchange, &ex.Values, marshalFields); err != nil {
				return nil, "", common.CompileError{Err: err, Path: ctx.PathRef(), Proto: ProtoName}
			}
			structValues.Values.Set("QueueConfiguration", ex)
		}
	}

	// Publish channel bindings
	if channel.Publish != nil {
		if b, ok := channel.Publish.Bindings.Get(ProtoName); ok {
			ctx.Logger.Trace("Channel publish operation bindings", "proto", ProtoName)
			pob := &render.StructInit{
				Type: &render.Simple{Name: "PublishOperationBindings", Package: ctx.RuntimePackage(ProtoName)},
			}
			hasBindings = true
			var bindings publishOperationBindings
			if err := utils.UnmarshalRawsUnion2(b, &bindings); err != nil {
				return nil, "", common.CompileError{Err: err, Path: ctx.PathRef(), Proto: ProtoName}
			}
			marshalFields := []string{"Expiration", "UserID", "CC", "Priority", "Mandatory", "BCC", "ReplyTo", "Timestamp"}
			if err := utils.StructToOrderedMap(bindings, &pob.Values, marshalFields); err != nil {
				return nil, "", common.CompileError{Err: err, Path: ctx.PathRef(), Proto: ProtoName}
			}
			switch bindings.DeliveryMode {
			case 1:
				pob.Values.Set("DeliveryMode", &render.Simple{Name: "DeliveryModeTransient", Package: ctx.RuntimePackage(ProtoName)})
			case 2:
				pob.Values.Set("DeliveryMode", &render.Simple{Name: "DeliveryModePersistent", Package: ctx.RuntimePackage(ProtoName)})
			case 0:
			default:
				return nil, "", common.CompileError{
					Err:   fmt.Errorf("unknown delivery mode %v", bindings.DeliveryMode),
					Path:  ctx.PathRef(),
					Proto: ProtoName,
				}
			}

			structValues.Values.Set("PublisherBindings", pob)
		}
	}

	// Subscribe channel bindings
	if channel.Subscribe != nil {
		if b, ok := channel.Subscribe.Bindings.Get(ProtoName); ok {
			ctx.Logger.Trace("Channel subscribe operation bindings", "proto", ProtoName)
			sob := &render.StructInit{
				Type: &render.Simple{Name: "SubscribeOperationBindings", Package: ctx.RuntimePackage(ProtoName)},
			}
			hasBindings = true
			var bindings subscribeOperationBindings
			if err := utils.UnmarshalRawsUnion2(b, &bindings); err != nil {
				return nil, "", common.CompileError{Err: err, Path: ctx.PathRef(), Proto: ProtoName}
			}
			marshalFields := []string{"Expiration", "UserID", "CC", "Priority", "ReplyTo", "Timestamp", "Ack"}
			if err := utils.StructToOrderedMap(bindings, &sob.Values, marshalFields); err != nil {
				return nil, "", common.CompileError{Err: err, Path: ctx.PathRef(), Proto: ProtoName}
			}
			switch bindings.DeliveryMode {
			case 1:
				sob.Values.Set("DeliveryMode", &render.Simple{Name: "DeliveryModeTransient", Package: ctx.RuntimePackage(ProtoName)})
			case 2:
				sob.Values.Set("DeliveryMode", &render.Simple{Name: "DeliveryModePersistent", Package: ctx.RuntimePackage(ProtoName)})
			case 0:
			default:
				return nil, "", common.CompileError{
					Err:   fmt.Errorf("unknown delivery mode %v", bindings.DeliveryMode),
					Path:  ctx.PathRef(),
					Proto: ProtoName,
				}
			}

			structValues.Values.Set("SubscriberBindings", sob)
		}
	}

	if !hasBindings {
		return nil, "", nil
	}

	// Method Proto() proto.ChannelBindings
	res := &render.Func{
		FuncSignature: render.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []render.FuncParam{
				{Type: render.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}},
			},
		},
		Receiver:     bindingsStruct,
		PackageName:  ctx.TopPackageName(),
		BodyRenderer: protocols.ChannelBindingsMethodBody(structValues, nil, nil),
	}

	return res, chanType, nil
}

type ProtoChannel struct {
	protocols.BaseProtoChannel
	BindingsStructNoRender *render.Struct // nil if bindings not set FIXME: remove in favor of struct in parent channel
	BindingsMethod         *render.Func
	BindingsChannelType    string
}

func (p ProtoChannel) DirectRendering() bool {
	return true
}

func (p ProtoChannel) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	if p.BindingsMethod != nil {
		res = append(res, p.BindingsMethod.RenderDefinition(ctx)...)
	}
	res = append(res, p.ServerIface.RenderDefinition(ctx)...)
	res = append(res, protocols.RenderChannelOpenFunc(
		ctx, p.Struct, p.Name, p.ServerIface, p.ParametersStructNoRender, p.BindingsStructNoRender,
		p.Publisher, p.Subscriber, ProtoName, protoAbbr,
	)...)
	res = append(res, p.renderNewFunc(ctx)...)
	res = append(res, p.Struct.RenderDefinition(ctx)...)
	res = append(res, protocols.RenderChannelCommonMethods(ctx, p.Struct, p.Publisher, p.Subscriber, protoAbbr)...)
	res = append(res, p.renderCommonMethods(ctx)...)
	if p.Publisher {
		res = append(res, protocols.RenderChannelPublisherMethods(ctx, p.Struct, ProtoName)...)
		res = append(res, p.renderPublisherMethods(ctx)...)
	}
	if p.Subscriber {
		res = append(res, protocols.RenderChannelSubscriberMethods(
			ctx, p.Struct, p.SubMessageLink, p.FallbackMessageType, ProtoName, protoAbbr,
		)...)
	}
	return res
}

func (p ProtoChannel) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	return p.Struct.RenderUsage(ctx)
}

func (p ProtoChannel) String() string {
	return "AMQP channel " + p.BaseProtoChannel.Name
}

func (p ProtoChannel) renderNewFunc(ctx *common.RenderContext) []*j.Statement {
	return []*j.Statement{
		// NewChannel1Proto(params Channel1Parameters, publisher proto.Publisher, subscriber proto.Subscriber) *Channel1Proto
		j.Func().Id(p.Struct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				if p.ParametersStructNoRender != nil {
					g.Id("params").Add(utils.ToCode(p.ParametersStructNoRender.RenderUsage(ctx))...)
				}
				if p.Publisher {
					g.Id("publisher").Qual(ctx.RuntimePackage(ProtoName), "Publisher")
				}
				if p.Subscriber {
					g.Id("subscriber").Qual(ctx.RuntimePackage(ProtoName), "Subscriber")
				}
			}).
			Op("*").Add(utils.ToCode(p.Struct.RenderUsage(ctx))...).
			BlockFunc(func(bg *j.Group) {
				bg.Op("res := ").Add(utils.ToCode(p.Struct.RenderUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
					d[j.Id("name")] = j.Id(utils.ToGolangName(p.Name, true) + "Name").CallFunc(func(g *j.Group) {
						if p.ParametersStructNoRender != nil {
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
					if p.BindingsStructNoRender != nil {
						bg.Id("bindings").Op(":=").Add(utils.ToCode(p.BindingsStructNoRender.RenderUsage(ctx))...).Values().Dot(protoAbbr).Call()
						bg.Op(`
							if bindings.ExchangeConfiguration.Name != "" {
								res.exchange = bindings.ExchangeConfiguration.Name
							}
							res.queue = bindings.QueueConfiguration.Name`)
					}
				case "queue":
					bg.Op("res.queue = res.name.String()")
					if p.BindingsStructNoRender != nil {
						bg.Id("bindings").Op(":=").Add(utils.ToCode(p.BindingsStructNoRender.RenderUsage(ctx))...).Values().Dot(protoAbbr).Call()
						bg.Op(`
							if bindings.QueueConfiguration.Name != "" {
								res.queue = bindings.QueueConfiguration.Name
							}
							res.exchange = bindings.ExchangeConfiguration.Name`)
					}
				default:
					ctx.Logger.Fatalf("Unknown channel type: %q", p.BindingsChannelType)
				}
				bg.Op(`return &res`)
			}),
	}
}

func (p ProtoChannel) renderCommonMethods(_ *common.RenderContext) []*j.Statement {
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

func (p ProtoChannel) renderPublisherMethods(ctx *common.RenderContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	var msgTyp common.GolangType = render.NullableType{Type: p.FallbackMessageType, Render: true}
	if p.PubMessageLink != nil {
		msgTyp = render.NullableType{Type: p.PubMessageLink.Target().OutStruct, Render: true}
	}

	var msgBindings *render.Struct
	if p.PubMessageLink != nil {
		if _, ok := p.PubMessageLink.Target().BindingsStructProtoMethods.Get(ProtoName); ok {
			msgBindings = p.PubMessageLink.Target().BindingsStruct
		}
	}

	return []*j.Statement{
		// Method MakeEnvelope(envelope proto.EnvelopeWriter, message *Message1Out) error
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			ParamsFunc(func(g *j.Group) {
				g.Id("envelope").Qual(ctx.RuntimePackage(ProtoName), "EnvelopeWriter")
				g.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...)
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
						j.Add(utils.ToCode(msgBindings.RenderUsage(ctx))...).Values().Dot("AMQP()"),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}
