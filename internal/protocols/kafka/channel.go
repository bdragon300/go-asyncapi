package kafka

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
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
	paramsLnk := assemble.NewListCbLink[*assemble.Parameter](func(item common.Assembler, path []string) bool {
		par, ok := item.(*assemble.Parameter)
		if !ok {
			return false
		}
		_, ok = channel.Parameters.Get(par.Name)
		return ok
	})
	ctx.Linker.AddMany(paramsLnk)

	chanResult := &ProtoChannel{
		Name: channelKey,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", "Kafka"),
				Description: channel.Description,
				Render:      true,
				Package:     ctx.TopPackageName(),
			},
			Fields: []assemble.StructField{
				{Name: "name", Type: &assemble.Simple{Type: "ParamString", Package: ctx.RuntimePackage("")}},
				{Name: "topic", Type: &assemble.Simple{Type: "string"}},
			},
		},
		FallbackMessageType: &assemble.Simple{Type: "any", IsIface: true},
	}

	// FIXME: remove in favor of the non-proto channel
	if channel.Parameters.Len() > 0 {
		chanResult.ParametersStructNoAssemble = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    ctx.GenerateObjName("", "Parameters"),
				Render:  true,
				Package: ctx.TopPackageName(),
			},
			Fields: nil,
		}
		for _, paramName := range channel.Parameters.Keys() {
			ref := path.Join(ctx.PathRef(), "parameters", paramName)
			lnk := assemble.NewRefLinkAsGolangType(ref)
			ctx.Linker.Add(lnk)
			chanResult.ParametersStructNoAssemble.Fields = append(chanResult.ParametersStructNoAssemble.Fields, assemble.StructField{
				Name: utils.ToGolangName(paramName, true),
				Type: lnk,
			})
		}
	}

	// Interface to match servers bound with a channel
	var ifaceFirstMethodParams []assemble.FuncParam
	if chanResult.ParametersStructNoAssemble != nil {
		ifaceFirstMethodParams = append(ifaceFirstMethodParams, assemble.FuncParam{
			Name: "params",
			Type: &assemble.Simple{Type: chanResult.ParametersStructNoAssemble.Name, Package: ctx.TopPackageName()},
		})
	}
	chanResult.ServerIface = &assemble.Interface{
		BaseType: assemble.BaseType{
			Name:    utils.ToLowerFirstLetter(chanResult.Struct.Name + "Server"),
			Render:  true,
			Package: ctx.TopPackageName(),
		},
		Methods: []assemble.FuncSignature{
			{
				Name: "Open" + chanResult.Struct.Name,
				Args: ifaceFirstMethodParams,
				Return: []assemble.FuncParam{
					{Type: &assemble.Simple{Type: chanResult.Struct.Name, Package: ctx.TopPackageName()}, Pointer: true},
					{Type: &assemble.Simple{Type: "error"}},
				},
			},
		},
	}

	// Publisher stuff
	if channel.Publish != nil {
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, assemble.StructField{
			Name:        "publisher",
			Description: channel.Publish.Description,
			Type: &assemble.Simple{
				Type:    "Publisher",
				Package: ctx.RuntimePackage(protoName),
				IsIface: true,
			},
		})
		chanResult.Publisher = true
		if channel.Publish.Message != nil {
			ref := path.Join(ctx.PathRef(), "publish/message")
			chanResult.PubMessageLink = assemble.NewRefLink[*assemble.Message](ref)
			ctx.Linker.Add(chanResult.PubMessageLink)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, assemble.FuncSignature{
			Name: "Producer",
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: &assemble.Simple{Type: "Producer", Package: ctx.RuntimePackage(protoName), IsIface: true}},
			},
		})
	}

	// Subscriber stuff
	if channel.Subscribe != nil {
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, assemble.StructField{
			Name:        "subscriber",
			Description: channel.Subscribe.Description,
			Type: &assemble.Simple{
				Type:    "Subscriber",
				Package: ctx.RuntimePackage(protoName),
				IsIface: true,
			},
		})
		chanResult.Subscriber = true
		if channel.Subscribe.Message != nil {
			ref := path.Join(ctx.PathRef(), "subscribe/message")
			chanResult.SubMessageLink = assemble.NewRefLink[*assemble.Message](ref)
			ctx.Linker.Add(chanResult.SubMessageLink)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, assemble.FuncSignature{
			Name: "Consumer",
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: &assemble.Simple{Type: "Consumer", Package: ctx.RuntimePackage(protoName), IsIface: true}},
			},
		})
	}

	if bindings, ok, err := buildChannelBindings(channel); err != nil {
		return nil, err
	} else if ok {
		chanResult.BindingsStructNoAssemble = &assemble.Struct{ // TODO: remove in favor of parent channel
			BaseType: assemble.BaseType{
				Name:    ctx.GenerateObjName("", "Bindings"),
				Render:  true,
				Package: ctx.TopPackageName(),
			},
		}
		chanResult.BindingsValues = &bindings
	}

	return chanResult, nil
}

func buildChannelBindings(channel *compile.Channel) (res ProtoChannelBindings, hasBindings bool, err error) {
	if chBindings, ok := channel.Bindings.Get(protoName); ok {
		var bindings channelBindings
		hasBindings = true
		if err = utils.UnmarshalRawsUnion2(chBindings, &bindings); err != nil {
			return
		}
		marshalFields := []string{"Topic", "Partitions", "Replicas"}
		if err = utils.StructToOrderedMap(bindings, &res.StructValues, marshalFields); err != nil {
			return
		}

		if bindings.TopicConfiguration != nil {
			tc := bindings.TopicConfiguration
			marshalFields = []string{"RetentionMs", "RetentionBytes", "DeleteRetentionMs", "MaxMessageBytes"}
			if err = utils.StructToOrderedMap(*bindings.TopicConfiguration, &res.StructValues, marshalFields); err != nil {
				return
			}
			if lo.Contains(tc.CleanupPolicy, "delete") {
				res.CleanupPolicyStructValue.Set("Delete", true)
			}
			if lo.Contains(tc.CleanupPolicy, "compact") {
				res.CleanupPolicyStructValue.Set("Compact", true)
			}
		}
	}

	if channel.Publish != nil {
		if opBindings, ok := channel.Publish.Bindings.Get(protoName); ok {
			hasBindings = true
			if res.PublisherJSONValues, err = buildOperationBindings(opBindings); err != nil {
				return
			}
		}
	}

	if channel.Subscribe != nil {
		if opBindings, ok := channel.Subscribe.Bindings.Get(protoName); ok {
			hasBindings = true
			if res.SubscriberJSONValues, err = buildOperationBindings(opBindings); err != nil {
				return
			}
		}
	}

	return
}

func buildOperationBindings(opBindings utils.Union2[json.RawMessage, yaml.Node]) (res utils.OrderedMap[string, string], err error) {
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

type ProtoChannelBindings struct {
	StructValues             utils.OrderedMap[string, any]
	CleanupPolicyStructValue utils.OrderedMap[string, bool]
	PublisherJSONValues      utils.OrderedMap[string, string]
	SubscriberJSONValues     utils.OrderedMap[string, string]
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
		// Method Kafka() kafka.ChannelBindings
		j.Func().Params(receiver.Clone()).Id("Kafka").
			Params().
			Qual(ctx.RuntimePackage(protoName), "ChannelBindings").
			BlockFunc(func(blockGroup *j.Group) {
				blockGroup.Id("b").Op(":=").Qual(ctx.RuntimePackage(protoName), "ChannelBindings").Values(j.DictFunc(func(d j.Dict) {
					for _, v := range p.BindingsValues.StructValues.Entries() {
						d[j.Id(v.Key)] = j.Lit(v.Value)
					}
					d[j.Id("CleanupPolicy")] = j.Qual(ctx.RuntimePackage(protoName), "TopicCleanupPolicy").Values(j.DictFunc(func(d2 j.Dict) {
						for _, v2 := range p.BindingsValues.CleanupPolicyStructValue.Entries() {
							d2[j.Id(v2.Key)] = j.Lit(v2.Value)
						}
					}))
				}))
				for _, e := range p.BindingsValues.PublisherJSONValues.Entries() {
					n := utils.ToLowerFirstLetter(e.Key)
					blockGroup.Id(n).Op(":=").Lit(e.Value)
					blockGroup.Empty().Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.PublisherBindings.%[2]s)", n, e.Key))
				}
				for _, e := range p.BindingsValues.SubscriberJSONValues.Entries() {
					n := utils.ToLowerFirstLetter(e.Key)
					blockGroup.Id(n).Op(":=").Lit(e.Value)
					blockGroup.Empty().Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.SubscriberBindings.%[2]s)", n, e.Key))
				}
				blockGroup.Return(j.Id("b"))
			},
			),
	}
}

func (p ProtoChannel) assembleOpenFunc(ctx *common.AssembleContext) []*j.Statement {
	return []*j.Statement{
		// OpenChannel1Kafka(params Channel1Parameters, servers ...channel1KafkaServer) (*Channel1Kafka, error)
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
					bodyGroup.Var().Id("prod").Index().Qual(ctx.RuntimePackage(protoName), "Producer")
				}
				if p.Subscriber {
					bodyGroup.Var().Id("cons").Index().Qual(ctx.RuntimePackage(protoName), "Consumer")
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
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter"), j.Qual(ctx.RuntimePackage(protoName), "ChannelBindings")).
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
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter")).
						Op("{Publishers: pubs}")
				}
				if p.Subscriber {
					bodyGroup.Op("subs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherSubscribers").
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeReader"), j.Qual(ctx.RuntimePackage(protoName), "ChannelBindings")).
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
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeReader")).
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
		// NewChannel1Kafka(params Channel1Parameters, publisher kafka.Publisher, subscriber kafka.Subscriber) *Channel1Kafka
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

		// Protocol() run.Protocol
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
				j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter"),
				j.Id("message").Qual(ctx.RuntimePackage(protoName), "EnvelopeMarshaler"),
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
			Qual(ctx.RuntimePackage(protoName), "Publisher").
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
					return %[1]s.publisher.Send(ctx, envelopes...)`, rn, ctx.RuntimePackage(protoName)),
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
				j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeReader"),
				j.Id("message").Qual(ctx.RuntimePackage(protoName), "EnvelopeUnmarshaler"),
			).
			Error().
			Block(
				j.Op(`return message.UnmarshalKafkaEnvelope(envelope)`),
			),

		// Method Subscriber() kafka.Subscriber
		j.Func().Params(receiver.Clone()).Id("Subscriber").
			Params().
			Qual(ctx.RuntimePackage(protoName), "Subscriber").
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
						Params(j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeReader")).
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
