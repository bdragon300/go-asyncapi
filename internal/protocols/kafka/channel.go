package kafka

import (
	"encoding/json"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
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

func BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, channelKey string) (common.Renderer, error) {
	baseChan, err := protocols.BuildChannel(ctx, channel, channelKey, ProtoName, protoAbbr)
	if err != nil {
		return nil, err
	}

	baseChan.Struct.Fields = append(baseChan.Struct.Fields, render.StructField{Name: "topic", Type: &render.Simple{Name: "string"}})

	chanResult := &ProtoChannel{BaseProtoChannel: *baseChan}

	// Channel bindings
	ctx.Logger.Trace("Channel bindings")
	ctx.Logger.NextCallLevel()
	bindingsStruct := &render.Struct{ // TODO: remove in favor of parent channel
		BaseType: render.BaseType{
			Name:         ctx.GenerateObjName(channelKey, "Bindings"),
			DirectRender: true,
			PackageName:  ctx.TopPackageName(),
		},
	}
	method, err := buildChannelBindingsMethod(ctx, channel, bindingsStruct)
	ctx.Logger.PrevCallLevel()
	if err != nil {
		return nil, err
	}
	if method != nil {
		chanResult.BindingsStructNoRender = bindingsStruct
		chanResult.BindingsMethod = method
	}

	return chanResult, nil
}

func buildChannelBindingsMethod(ctx *common.CompileContext, channel *asyncapi.Channel, bindingsStruct *render.Struct) (*render.Func, error) {
	structValues := &render.StructInit{Type: &render.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}}
	var hasBindings bool

	if chBindings, ok := channel.Bindings.Get(ProtoName); ok {
		ctx.Logger.Trace("Channel bindings", "proto", ProtoName)
		hasBindings = true
		var bindings channelBindings
		if err := types.UnmarshalRawsUnion2(chBindings, &bindings); err != nil {
			return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
		}
		marshalFields := []string{"Topic", "Partitions", "Replicas"}
		if err := utils.StructToOrderedMap(bindings, &structValues.Values, marshalFields); err != nil {
			return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
		}

		if bindings.TopicConfiguration != nil {
			tc := &render.StructInit{
				Type: &render.Simple{Name: "TopicConfiguration", Package: ctx.RuntimePackage(ProtoName)},
			}
			marshalFields = []string{"RetentionMs", "RetentionBytes", "DeleteRetentionMs", "MaxMessageBytes"}
			if err := utils.StructToOrderedMap(*bindings.TopicConfiguration, &tc.Values, marshalFields); err != nil {
				return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
			}

			if len(bindings.TopicConfiguration.CleanupPolicy) > 0 {
				tcp := &render.StructInit{
					Type: &render.Simple{Name: "TopicCleanupPolicy", Package: ctx.RuntimePackage(ProtoName)},
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
	var publisherJSON types.OrderedMap[string, any]
	if channel.Publish != nil {
		ctx.Logger.Trace("Channel publish operation bindings")
		if b, ok := channel.Publish.Bindings.Get(ProtoName); ok {
			hasBindings = true
			var err error
			if publisherJSON, err = buildOperationBindings(b); err != nil {
				return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
			}
		}
	}

	// Subscribe channel bindings
	var subscriberJSON types.OrderedMap[string, any]
	if channel.Subscribe != nil {
		ctx.Logger.Trace("Channel subscribe operation bindings")
		if b, ok := channel.Subscribe.Bindings.Get(ProtoName); ok {
			hasBindings = true
			var err error
			if subscriberJSON, err = buildOperationBindings(b); err != nil {
				return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
			}
		}
	}

	if !hasBindings {
		return nil, nil
	}

	// Method Proto() proto.ChannelBindings
	return &render.Func{
		FuncSignature: render.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []render.FuncParam{
				{Type: render.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}},
			},
		},
		Receiver:     bindingsStruct,
		PackageName:  ctx.TopPackageName(),
		BodyRenderer: protocols.ChannelBindingsMethodBody(structValues, &publisherJSON, &subscriberJSON),
	}, nil
}

func buildOperationBindings(opBindings types.Union2[json.RawMessage, yaml.Node]) (res types.OrderedMap[string, any], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(opBindings, &bindings); err != nil {
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
	BindingsStructNoRender *render.Struct // nil if bindings not set FIXME: remove in favor of struct in parent channel
	BindingsMethod         *render.Func
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
	res = append(res, protocols.RenderChannelCommonMethods(ctx, p.Struct, p.Publisher, p.Subscriber)...)
	res = append(res, p.renderCommonMethods(ctx)...)
	if p.Publisher {
		res = append(res, protocols.RenderChannelPublisherMethods(ctx, p.Struct, ProtoName)...)
		res = append(res, p.renderPublisherMethods(ctx)...)
	}
	if p.Subscriber {
		res = append(res, protocols.RenderChannelSubscriberMethods(
			ctx, p.Struct, p.SubMessagePromise, p.FallbackMessageType, ProtoName, protoAbbr,
		)...)
	}
	return res
}

func (p ProtoChannel) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	return p.Struct.RenderUsage(ctx)
}

func (p ProtoChannel) String() string {
	return p.BaseProtoChannel.Name
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
				bg.Op("res.topic = res.name.String()")
				if p.BindingsStructNoRender != nil {
					bg.Id("bindings").Op(":=").Add(utils.ToCode(p.BindingsStructNoRender.RenderUsage(ctx))...).Values().Dot(protoAbbr).Call()
					bg.Op(`
						if bindings.Topic != "" {
							res.topic = bindings.Topic
						}`)
				}
				bg.Op(`return &res`)
			}),
	}
}

func (p ProtoChannel) renderCommonMethods(_ *common.RenderContext) []*j.Statement {
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

func (p ProtoChannel) renderPublisherMethods(ctx *common.RenderContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	var msgTyp common.GolangType = render.Pointer{Type: p.FallbackMessageType, DirectRender: true}
	if p.PubMessagePromise != nil {
		msgTyp = render.Pointer{Type: p.PubMessagePromise.Target().OutStruct, DirectRender: true}
	}

	var msgBindings *render.Struct
	if p.PubMessagePromise != nil {
		if _, ok := p.PubMessagePromise.Target().BindingsStructProtoMethods.Get(ProtoName); ok {
			msgBindings = p.PubMessagePromise.Target().BindingsStruct
		}
	}

	return []*j.Statement{
		// Method MakeEnvelope(envelope kafka.EnvelopeWriter, message *Message1Out) error
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			ParamsFunc(func(g *j.Group) {
				g.Id("envelope").Qual(ctx.RuntimePackage(ProtoName), "EnvelopeWriter")
				g.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...)
			}).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("envelope.ResetPayload()")
				if p.PubMessagePromise == nil { // No Message set for Channel in spec
					bg.Empty().Add(utils.QualSprintf(`
						enc := %Q(encoding/json,NewEncoder)(envelope)
						if err := enc.Encode(message); err != nil {
							return err
						}`))
				} else { // Message is set for Channel in spec
					bg.Op(`
						if err := message.MarshalKafkaEnvelope(envelope); err != nil {
							return err
						}`)
				}
				bg.Op("envelope.SetTopic").Call(j.Id(rn).Dot("topic"))
				if msgBindings != nil {
					bg.Op("envelope.SetBindings").Call(
						j.Add(utils.ToCode(msgBindings.RenderUsage(ctx))...).Values().Dot("Kafka()"),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}
