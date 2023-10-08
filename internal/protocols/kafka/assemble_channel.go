package kafka

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type ProtoChannelBindings struct {
	StructValues             utils.OrderedMap[string, any]
	CleanupPolicyStructValue utils.OrderedMap[string, bool]
	PublisherValues          utils.OrderedMap[string, any]
	SubscriberValues         utils.OrderedMap[string, any]
}

type ProtoChannel struct {
	Name                       string
	Publisher                  bool
	Subscriber                 bool
	Struct                     *assemble.Struct
	ServerIface                *assemble.Interface
	ParametersStructNoAssemble *assemble.Struct      // nil if parameters not set
	BindingsStructNoAssemble   *assemble.Struct      // nil if bindings not set FIXME: remove in favor of struct in parent channel
	BindingsValues             *ProtoChannelBindings // nil if bindings don't set particularly for this protocol

	PubMessageLink      *assemble.Link[*assemble.Message] // nil when message is not set
	SubMessageLink      *assemble.Link[*assemble.Message] // nil when message is not set
	FallbackMessageType common.Assembler
}

func (p ProtoChannel) AllowRender() bool {
	return true
}

func (p ProtoChannel) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	if p.BindingsStructNoAssemble != nil && p.BindingsValues != nil {
		res = append(res, p.assembleBindingsMethod(ctx)...)
	}
	res = append(res, p.ServerIface.AssembleDefinition(ctx)...)
	res = append(res, p.assembleOpenFunc(ctx)...)
	res = append(res, p.assembleNewFunc(ctx)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.assembleCommonMethods(ctx)...)
	if p.Publisher {
		res = append(res, p.assemblePublisherMethods(ctx)...)
	}
	if p.Subscriber {
		res = append(res, p.assembleSubscriberMethods(ctx)...)
	}
	return res
}

func (p ProtoChannel) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoChannel) assembleBindingsMethod(ctx *common.AssembleContext) []*j.Statement {
	rn := p.BindingsStructNoAssemble.ReceiverName()
	receiver := j.Id(rn).Id(p.BindingsStructNoAssemble.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Kafka").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "ChannelBindings").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage("kafka"), "ChannelBindings").Values(j.DictFunc(func(d j.Dict) {
					for _, v := range p.BindingsValues.StructValues.Entries() {
						d[j.Id(v.Key)] = j.Lit(v.Value)
					}
					d[j.Id("PublisherBindings")] = j.Qual(ctx.RuntimePackage("kafka"), "OperationBindings").Values(j.DictFunc(func(d2 j.Dict) {
						for _, v2 := range p.BindingsValues.PublisherValues.Entries() {
							d2[j.Id(v2.Key)] = j.Lit(v2.Value)
						}
					}))
					d[j.Id("SubscriberBindings")] = j.Qual(ctx.RuntimePackage("kafka"), "OperationBindings").Values(j.DictFunc(func(d2 j.Dict) {
						for _, v2 := range p.BindingsValues.SubscriberValues.Entries() {
							d2[j.Id(v2.Key)] = j.Lit(v2.Value)
						}
					}))
					d[j.Id("CleanupPolicy")] = j.Qual(ctx.RuntimePackage("kafka"), "TopicCleanupPolicy").Values(j.DictFunc(func(d2 j.Dict) {
						for _, v2 := range p.BindingsValues.CleanupPolicyStructValue.Entries() {
							d2[j.Id(v2.Key)] = j.Lit(v2.Value)
						}
					}))
				}))),
			),
	}
}

func (p ProtoChannel) assembleOpenFunc(ctx *common.AssembleContext) []*j.Statement {
	return []*j.Statement{
		j.Func().Id("Open"+p.Struct.Name).
			ParamsFunc(func(g *j.Group) {
				if p.ParametersStructNoAssemble != nil {
					g.Id("params").Add(utils.ToCode(p.ParametersStructNoAssemble.AssembleUsage(ctx))...)
				}
				g.Id("servers").Op("...").Add(utils.ToCode(p.ServerIface.AssembleUsage(ctx))...)
			}).
			Params(j.Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...), j.Error()).
			BlockFunc(func(bodyGroup *j.Group) {
				bodyGroup.Op("if len(servers) == 0").Block(j.Op("return nil, ").Qual(ctx.RuntimePackage(""), "ErrEmptyServers"))
				bodyGroup.Id("name").Op(":=").Id(utils.ToGolangName(p.Name, true) + "Name").CallFunc(func(g *j.Group) {
					if p.ParametersStructNoAssemble != nil {
						g.Id("params")
					}
				})
				if p.BindingsStructNoAssemble != nil {
					bodyGroup.Op(fmt.Sprintf("bindings := %s{}.Kafka()", p.BindingsStructNoAssemble.Name))
				}
				if p.Publisher {
					bodyGroup.Var().Id("prod").Index().Qual(ctx.RuntimePackage("kafka"), "Producer")
				}
				if p.Subscriber {
					bodyGroup.Var().Id("cons").Index().Qual(ctx.RuntimePackage("kafka"), "Consumer")
				}
				bodyGroup.Op("for _, srv := range servers").BlockFunc(func(g *j.Group) {
					if p.Publisher {
						g.Op("prod = append(prod, srv.Producer())")
					}
					if p.Subscriber {
						g.Op("cons = append(cons, srv.Consumer())")
					}
				})
				if p.Publisher {
					bodyGroup.Op("pubs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherPublishers").
						Types(j.Qual(ctx.RuntimePackage("kafka"), "EnvelopeWriter"), j.Qual(ctx.RuntimePackage("kafka"), "ChannelBindings")).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(p.BindingsStructNoAssemble != nil, "&bindings", "nil"))
							g.Id("prod")
						})
					bodyGroup.Op(`
						if err != nil {
							return nil, err
						}`)
					bodyGroup.Op("pub := ").Qual(ctx.RuntimePackage(""), "PublisherFanOut").
						Types(j.Qual(ctx.RuntimePackage("kafka"), "EnvelopeWriter")).
						Op("{Publishers: pubs}")
				}
				if p.Subscriber {
					bodyGroup.Op("subs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherSubscribers").
						Types(j.Qual(ctx.RuntimePackage("kafka"), "EnvelopeReader"), j.Qual(ctx.RuntimePackage("kafka"), "ChannelBindings")).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(p.BindingsStructNoAssemble != nil, "&bindings", "nil"))
							g.Id("cons")
						})
					bodyGroup.Op("if err != nil").BlockFunc(func(g *j.Group) {
						if p.Publisher {
							g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, pub.Close())"))
						}
						g.Op("return nil, err")
					})
					bodyGroup.Op("sub := ").Qual(ctx.RuntimePackage(""), "SubscriberFanIn").
						Types(j.Qual(ctx.RuntimePackage("kafka"), "EnvelopeReader")).
						Op("{Subscribers: subs}")
				}
				bodyGroup.Op("ch := ").Id(p.Struct.NewFuncName()).CallFunc(func(g *j.Group) {
					g.Id("params")
					if p.Publisher {
						g.Id("pub")
					}
					if p.Subscriber {
						g.Id("sub")
					}
				})
				bodyGroup.Op("return ch, nil")
			}),
	}
}

func (p ProtoChannel) assembleNewFunc(ctx *common.AssembleContext) []*j.Statement {
	return []*j.Statement{
		j.Func().Id(p.Struct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				if p.ParametersStructNoAssemble != nil {
					g.Id("params").Add(utils.ToCode(p.ParametersStructNoAssemble.AssembleUsage(ctx))...)
				}
				if p.Publisher {
					g.Id("publisher").Qual(ctx.RuntimePackage("kafka"), "Publisher")
				}
				if p.Subscriber {
					g.Id("subscriber").Qual(ctx.RuntimePackage("kafka"), "Subscriber")
				}
			}).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			BlockFunc(func(bodyGroup *j.Group) {
				bodyGroup.Op("res := ").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
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
				bodyGroup.Op("res.topic = res.name.String()")
				if p.BindingsStructNoAssemble != nil {
					bodyGroup.Op(fmt.Sprintf("bindings := %s{}.Kafka()", p.BindingsStructNoAssemble.Name))
					bodyGroup.Op(`
						if bindings.Topic != "" {
							res.topic = bindings.Topic
						}`)
				}
				bodyGroup.Op(`
					if res.topic == "" {
						res.topic = res.name.String()
					}
					return &res`)
			}),
	}
}

func (p ProtoChannel) assembleCommonMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			Qual(ctx.RuntimePackage(""), "ParamString").
			Block(
				j.Return(j.Id(rn).Dot("name")),
			),

		// Method Topic() string
		j.Func().Params(receiver.Clone()).Id("Topic").
			Params().
			String().
			Block(
				j.Return(j.Id(rn).Dot("topic")),
			),

		// Protocol() runtime.Protocol
		j.Func().Params(receiver.Clone()).Id("Protocol").
			Params().
			Qual(ctx.RuntimePackage(""), "Protocol").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage(""), "ProtocolKafka")),
			),

		// Method Close() (err error)
		j.Func().Params(receiver.Clone()).Id("Close").
			Params().
			Params(j.Err().Error()).
			BlockFunc(func(g *j.Group) {
				if p.Publisher {
					g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, %[1]s.publisher.Close())", rn))
				}
				if p.Subscriber {
					g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, %[1]s.subscriber.Close())", rn))
				}
				g.Return()
			}),
	}
}

func (p ProtoChannel) assemblePublisherMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	msgTyp := p.FallbackMessageType
	if p.PubMessageLink != nil {
		msgTyp = p.PubMessageLink.Target().OutStruct
	}

	return []*j.Statement{
		// Method MakeEnvelope(envelope kafka.EnvelopeWriter, message kafka.EnvelopeMarshaler) error
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			Params(
				j.Id("envelope").Qual(ctx.RuntimePackage("kafka"), "EnvelopeWriter"),
				j.Id("message").Qual(ctx.RuntimePackage("kafka"), "EnvelopeMarshaler"),
			).
			Error().
			Block(utils.QualSprintf(`
				envelope.ResetPayload()
				if err := message.MarshalKafkaEnvelope(envelope); err != nil {
					return err
				}

				envelope.SetMetadata(kafka.EnvelopeMeta{
					Topic:     %[1]s.topic,
					Partition: -1, // not set
					Timestamp: %Q(time,Time){},
				})
				return nil`, rn)),

		// Method Publisher() kafka.Publisher
		j.Func().Params(receiver.Clone()).Id("Publisher").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "Publisher").
			Block(
				j.Return(j.Id(rn).Dot("publisher")),
			),

		// Method Publish(ctx context.Context, messages ...*Message2Out) (err error)
		j.Func().Params(receiver.Clone()).Id("Publish").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("messages").Op("...").Op("*").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...), // FIXME: *any on fallback variant
			).
			Error().
			Block(
				utils.QualSprintf(`
					envelopes := make([]%Q(%[2]s,EnvelopeWriter), 0, len(messages))
					for i := 0; i < len(messages); i++ {
						buf := new(%Q(%[2]s,EnvelopeOut))
						if err := %[1]s.MakeEnvelope(buf, messages[i]); err != nil {
							return %Q(fmt,Errorf)("make envelope #%%d error: %%w", i, err)
						}
						envelopes = append(envelopes, buf)
					}
					return %[1]s.publisher.Send(ctx, envelopes...)`, rn, ctx.RuntimePackage("kafka")),
			),
	}
}

func (p ProtoChannel) assembleSubscriberMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	msgTyp := p.FallbackMessageType
	if p.SubMessageLink != nil {
		msgTyp = p.SubMessageLink.Target().InStruct
	}

	return []*j.Statement{
		// Method ExtractEnvelope(envelope kafka.EnvelopeReader, message kafka.EnvelopeUnmarshaler) error
		j.Func().Params(receiver.Clone()).Id("ExtractEnvelope").
			Params(
				j.Id("envelope").Qual(ctx.RuntimePackage("kafka"), "EnvelopeReader"),
				j.Id("message").Qual(ctx.RuntimePackage("kafka"), "EnvelopeUnmarshaler"),
			).
			Error().
			Block(
				j.Op(`
					if err := message.UnmarshalKafkaEnvelope(envelope); err != nil {
						return err
					}
					return nil`),
			),

		// Method Subscriber() kafka.Subscriber
		j.Func().Params(receiver.Clone()).Id("Subscriber").
			Params().
			Qual(ctx.RuntimePackage("kafka"), "Subscriber").
			Block(
				j.Return(j.Id(rn).Dot("subscriber")),
			),

		// Method Subscribe(ctx context.Context, cb func(msg *Message2In) error) (err error)
		j.Func().Params(receiver.Clone()).Id("Subscribe").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("cb").Func().Params(j.Id("message").Op("*").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...)).Error(), // FIXME: *any on fallback variant
			).
			Error().
			Block(
				j.Return(j.Id(rn).Dot("subscriber.Receive").Call(
					j.Id("ctx"),
					j.Func().
						Params(j.Id("envelope").Qual(ctx.RuntimePackage("kafka"), "EnvelopeReader")).
						Error().
						BlockFunc(func(g *j.Group) {
							g.Op("buf := new").Call(j.Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...))
							g.Add(utils.QualSprintf(`
								if err := %[1]s.ExtractEnvelope(envelope, buf); err != nil {
									return %Q(fmt,Errorf)("envelope extraction error: %%w", err)
								}
								envelope.Commit()
								return cb(buf)`, rn))
						}),
				)),
			),
	}
}
