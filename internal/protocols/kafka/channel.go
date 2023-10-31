package kafka

import (
	"encoding/json"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/compile"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type channelBindings struct {
	Topic              *string             `json:"topic" yaml:"topic"`
	Partitions         *int                `json:"partitions" yaml:"partitions"`
	Replicas           *int                `json:"replicas" yaml:"replicas"`
	TopicConfiguration *topicConfiguration `json:"topicConfiguration" yaml:"topicConfiguration"`
}

type topicConfiguration struct {
	CleanupPolicy     []string `json:"cleanup.policy" yaml:"cleanup.policy"`
	RetentionMs       *int     `json:"retention.ms" yaml:"retention.ms"`
	RetentionBytes    *int     `json:"retention.bytes" yaml:"retention.bytes"`
	DeleteRetentionMs *int     `json:"delete.retention.ms" yaml:"delete.retention.ms"`
	MaxMessageBytes   *int     `json:"max.message.bytes" yaml:"max.message.bytes"`
}

type operationBindings struct {
	GroupID  any `json:"groupId" yaml:"groupId"`   // jsonschema object
	ClientID any `json:"clientId" yaml:"clientId"` // jsonschema object
}

func BuildChannel(ctx *common.CompileContext, channel *compile.Channel, channelKey string) (common.Assembler, error) {
	baseChan, err := protocols.BuildChannel(ctx, channel, channelKey, ProtoName, protoAbbr)
	if err != nil {
		return nil, err
	}

	baseChan.Struct.Fields = append(baseChan.Struct.Fields, assemble.StructField{Name: "topic", Type: &assemble.Simple{Name: "string"}})

	chanResult := &ProtoChannel{BaseProtoChannel: *baseChan}

	// Channel bindings
	bindingsStruct := &assemble.Struct{ // TODO: remove in favor of parent channel
		BaseType: assemble.BaseType{
			Name:        ctx.GenerateObjName(channelKey, "Bindings"),
			Render:      true,
			PackageName: ctx.TopPackageName(),
		},
	}
	method, err := buildChannelBindingsMethod(ctx, channel, bindingsStruct)
	if err != nil {
		return nil, err
	}
	if method != nil {
		chanResult.BindingsStructNoAssemble = bindingsStruct
		chanResult.BindingsMethod = method
	}

	return chanResult, nil
}

func buildChannelBindingsMethod(ctx *common.CompileContext, channel *compile.Channel, bindingsStruct *assemble.Struct) (res *assemble.Func, err error) {
	structValues := &assemble.StructInit{Type: &assemble.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}}
	var hasBindings bool

	if chBindings, ok := channel.Bindings.Get(ProtoName); ok {
		hasBindings = true
		var bindings channelBindings
		if err = utils.UnmarshalRawsUnion2(chBindings, &bindings); err != nil {
			return
		}
		marshalFields := []string{"Topic", "Partitions", "Replicas"}
		if err = utils.StructToOrderedMap(bindings, &structValues.Values, marshalFields); err != nil {
			return
		}

		if bindings.TopicConfiguration != nil {
			tc := &assemble.StructInit{
				Type: &assemble.Simple{Name: "TopicConfiguration", Package: ctx.RuntimePackage(ProtoName)},
			}
			marshalFields = []string{"RetentionMs", "RetentionBytes", "DeleteRetentionMs", "MaxMessageBytes"}
			if err = utils.StructToOrderedMap(*bindings.TopicConfiguration, &tc.Values, marshalFields); err != nil {
				return
			}

			if len(bindings.TopicConfiguration.CleanupPolicy) > 0 {
				tcp := &assemble.StructInit{
					Type: &assemble.Simple{Name: "TopicCleanupPolicy", Package: ctx.RuntimePackage(ProtoName)},
				}
				if lo.Contains(bindings.TopicConfiguration.CleanupPolicy, "delete") {
					tcp.Values.Set("Delete", true)
				}
				if lo.Contains(bindings.TopicConfiguration.CleanupPolicy, "compact") {
					tcp.Values.Set("Compact", true)
				}
				tc.Values.Set("CleanupPolicy", tcp)
			}

			structValues.Values.Set("TopicConfiguration", tc)
		}
	}

	// Publish channel bindings
	var publisherJSON utils.OrderedMap[string, any]
	if channel.Publish != nil {
		if b, ok := channel.Publish.Bindings.Get(ProtoName); ok {
			hasBindings = true
			if publisherJSON, err = buildOperationBindings(b); err != nil {
				return
			}
		}
	}

	// Subscribe channel bindings
	var subscriberJSON utils.OrderedMap[string, any]
	if channel.Subscribe != nil {
		if b, ok := channel.Subscribe.Bindings.Get(ProtoName); ok {
			hasBindings = true
			if subscriberJSON, err = buildOperationBindings(b); err != nil {
				return
			}
		}
	}

	if !hasBindings {
		return nil, nil
	}

	// Method Proto() proto.ChannelBindings
	res = &assemble.Func{
		FuncSignature: assemble.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: assemble.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}},
			},
		},
		Receiver:      bindingsStruct,
		PackageName:   ctx.TopPackageName(),
		BodyAssembler: protocols.ChannelBindingsMethodBody(structValues, &publisherJSON, &subscriberJSON),
	}

	return
}

func buildOperationBindings(opBindings utils.Union2[json.RawMessage, yaml.Node]) (res utils.OrderedMap[string, any], err error) {
	var bindings operationBindings
	if err = utils.UnmarshalRawsUnion2(opBindings, &bindings); err != nil {
		return
	}
	if bindings.GroupID != nil {
		v, err := json.Marshal(bindings.GroupID)
		if err != nil {
			return res, err
		}
		res.Set("GroupID", string(v))
	}
	if bindings.ClientID != nil {
		v, err := json.Marshal(bindings.ClientID)
		if err != nil {
			return res, err
		}
		res.Set("ClientID", string(v))
	}

	return
}

type ProtoChannel struct {
	protocols.BaseProtoChannel
	BindingsStructNoAssemble *assemble.Struct // nil if bindings not set FIXME: remove in favor of struct in parent channel
	BindingsMethod           *assemble.Func
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
		p.Publisher, p.Subscriber, ProtoName, protoAbbr,
	)...)
	res = append(res, p.assembleNewFunc(ctx)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, protocols.AssembleChannelCommonMethods(ctx, p.Struct, p.Publisher, p.Subscriber, protoAbbr)...)
	res = append(res, p.assembleCommonMethods(ctx)...)
	if p.Publisher {
		res = append(res, protocols.AssembleChannelPublisherMethods(ctx, p.Struct, ProtoName)...)
		res = append(res, p.assemblePublisherMethods(ctx)...)
	}
	if p.Subscriber {
		res = append(res, protocols.AssembleChannelSubscriberMethods(
			ctx, p.Struct, p.SubMessageLink, p.FallbackMessageType, ProtoName, protoAbbr,
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
					g.Id("publisher").Qual(ctx.RuntimePackage(ProtoName), "Publisher")
				}
				if p.Subscriber {
					g.Id("subscriber").Qual(ctx.RuntimePackage(ProtoName), "Subscriber")
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
				bg.Op("res.topic = res.name.String()")
				if p.BindingsStructNoAssemble != nil {
					bg.Id("bindings").Op(":=").Add(utils.ToCode(p.BindingsStructNoAssemble.AssembleUsage(ctx))...).Values().Dot(protoAbbr).Call()
					bg.Op(`
						if bindings.Topic != "" {
							res.topic = bindings.Topic
						}`)
				}
				bg.Op(`return &res`)
			}),
	}
}

func (p ProtoChannel) assembleCommonMethods(_ *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	return []*j.Statement{
		// Method Topic() string
		j.Func().Params(receiver.Clone()).Id("Topic").
			Params().
			String().
			Block(
				j.Return(j.Id(rn).Dot("topic")),
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
				g.Id("envelope").Qual(ctx.RuntimePackage(ProtoName), "EnvelopeWriter")
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
						if err := message.MarshalKafkaEnvelope(envelope); err != nil {
							return err
						}`)
				}
				bg.Op("envelope.SetTopic").Call(j.Id(rn).Dot("topic"))
				if msgBindings != nil {
					bg.Op("envelope.SetBindings").Call(
						j.Add(utils.ToCode(msgBindings.AssembleUsage(ctx))...).Values().Dot("Kafka()"),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}
