package kafka

import (
	"encoding/json"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/protocols"
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
	structValues := &assemble.StructInit{Type: &assemble.Simple{Type: "ChannelBindings", Package: ctx.RuntimePackage(protoName)}}
	var hasBindings bool

	if chBindings, ok := channel.Bindings.Get(protoName); ok {
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
				Type: &assemble.Simple{Type: "TopicConfiguration", Package: ctx.RuntimePackage(protoName)},
			}
			marshalFields = []string{"RetentionMs", "RetentionBytes", "DeleteRetentionMs", "MaxMessageBytes"}
			if err = utils.StructToOrderedMap(*bindings.TopicConfiguration, &tc.Values, marshalFields); err != nil {
				return
			}

			if len(bindings.TopicConfiguration.CleanupPolicy) > 0 {
				tcp := &assemble.StructInit{
					Type: &assemble.Simple{Type: "TopicCleanupPolicy", Package: ctx.RuntimePackage(protoName)},
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
		if b, ok := channel.Publish.Bindings.Get(protoName); ok {
			hasBindings = true
			if publisherJSON, err = buildOperationBindings(b); err != nil {
				return
			}
		}
	}

	// Subscribe channel bindings
	var subscriberJSON utils.OrderedMap[string, any]
	if channel.Subscribe != nil {
		if b, ok := channel.Subscribe.Bindings.Get(protoName); ok {
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
				{Type: assemble.Simple{Type: "ChannelBindings", Package: ctx.RuntimePackage(protoName)}},
			},
		},
		Receiver:      bindingsStruct,
		Package:       ctx.TopPackageName(),
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
		p.Publisher, p.Subscriber, protoName, protoAbbr,
	)...)
	res = append(res, p.assembleNewFunc(ctx)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, protocols.AssembleChannelCommonMethods(ctx, p.Struct, p.Publisher, p.Subscriber, protoAbbr)...)
	if p.Publisher {
		res = append(res, protocols.AssembleChannelPublisherMethods(
			ctx, p.Struct, p.PubMessageLink, p.FallbackMessageType, protoName, protoAbbr,
		)...)
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
