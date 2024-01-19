package mqtt

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/render/proto"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type ProtoChannel struct {
	proto.BaseProtoChannel
}

func (pc ProtoChannel) DirectRendering() bool {
	return true
}

func (pc ProtoChannel) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	ctx.LogRender("Channel", "", pc.Name, "definition", pc.DirectRendering(), "proto", pc.ProtoName)
	defer ctx.LogReturn()
	var res []*j.Statement
	res = append(res, pc.ServerIface.RenderDefinition(ctx)...)
	res = append(res, pc.RenderOpenFunc(
		ctx, pc.Struct, pc.Name, pc.ServerIface, pc.AbstractChannel.ParametersStruct, pc.AbstractChannel.BindingsStruct,
		pc.Publisher, pc.Subscriber,
	)...)
	res = append(res, pc.renderNewFunc(ctx)...)
	res = append(res, pc.Struct.RenderDefinition(ctx)...)
	res = append(res, pc.RenderCommonMethods(ctx, pc.Struct, pc.Publisher, pc.Subscriber)...)
	res = append(res, pc.renderMQTTMethods(ctx)...)
	if pc.Publisher {
		res = append(res, pc.RenderCommonPublisherMethods(ctx, pc.Struct)...)
		res = append(res, pc.renderMQTTPublisherMethods(ctx)...)
	}
	if pc.Subscriber {
		res = append(res, pc.RenderCommonSubscriberMethods(
			ctx, pc.Struct, pc.SubMessagePromise, pc.FallbackMessageType,
		)...)
	}
	return res
}

func (pc ProtoChannel) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	ctx.LogRender("Channel", "", pc.Name, "usage", pc.DirectRendering(), "proto", pc.ProtoName)
	defer ctx.LogReturn()
	return pc.Struct.RenderUsage(ctx)
}

func (pc ProtoChannel) ID() string {
	return pc.Name
}

func (pc ProtoChannel) String() string {
	return "MQTT ProtoChannel " + pc.Name
}

func (pc ProtoChannel) renderNewFunc(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderNewFunc", "proto", pc.ProtoName)

	return []*j.Statement{
		// NewChannel1Proto(params Channel1Parameters, publisher proto.Publisher, subscriber proto.Subscriber) *Channel1Proto
		j.Func().Id(pc.Struct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				if pc.AbstractChannel.ParametersStruct != nil {
					g.Id("params").Add(utils.ToCode(pc.AbstractChannel.ParametersStruct.RenderUsage(ctx))...)
				}
				if pc.Publisher {
					g.Id("publisher").Qual(ctx.RuntimeModule(pc.ProtoName), "Publisher")
				}
				if pc.Subscriber {
					g.Id("subscriber").Qual(ctx.RuntimeModule(pc.ProtoName), "Subscriber")
				}
			}).
			Op("*").Add(utils.ToCode(pc.Struct.RenderUsage(ctx))...).
			BlockFunc(func(bg *j.Group) {
				bg.Op("res := ").Add(utils.ToCode(pc.Struct.RenderUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
					d[j.Id("name")] = j.Id(utils.ToGolangName(pc.Name, true) + "Name").CallFunc(func(g *j.Group) {
						if pc.AbstractChannel.ParametersStruct != nil {
							g.Id("params")
						}
					})
					if pc.Publisher {
						d[j.Id("publisher")] = j.Id("publisher")
					}
					if pc.Subscriber {
						d[j.Id("subscriber")] = j.Id("subscriber")
					}
				}))
				bg.Op("res.topic = res.name.String()")
				bg.Op(`return &res`)
			}),
	}
}

func (pc ProtoChannel) renderMQTTMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderMQTTMethods", "proto", pc.ProtoName)

	rn := pc.Struct.ReceiverName()
	receiver := j.Id(rn).Id(pc.Struct.Name)

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

func (pc ProtoChannel) renderMQTTPublisherMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderMQTTPublisherMethods", "proto", pc.ProtoName)

	rn := pc.Struct.ReceiverName()
	receiver := j.Id(rn).Id(pc.Struct.Name)

	var msgTyp common.GolangType = render.GoPointer{Type: pc.FallbackMessageType, DirectRender: true}
	if pc.PubMessagePromise != nil {
		msgTyp = render.GoPointer{Type: pc.PubMessagePromise.Target().OutStruct, DirectRender: true}
	}

	return []*j.Statement{
		// Method MakeEnvelope(envelope proto.EnvelopeWriter, message *Message1Out) error
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			ParamsFunc(func(g *j.Group) {
				g.Id("envelope").Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeWriter")
				g.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...)
			}).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("envelope.ResetPayload()")
				if pc.PubMessagePromise == nil { // No Message set for Channel in spec
					bg.Empty().Add(utils.QualSprintf(`
						enc := %Q(encoding/json,NewEncoder)(envelope)
						if err := enc.Encode(message); err != nil {
							return err
						}`))
				} else { // Message is set for Channel in spec
					bg.Op(`
						if err := message.MarshalMQTTEnvelope(envelope); err != nil {
							return err
						}`)
				}
				bg.Op("envelope.SetTopic").Call(j.Id(rn).Dot("topic"))
				// Message SetBindings
				if pc.PubMessagePromise != nil && pc.PubMessagePromise.Target().HasProtoBindings(pc.ProtoName) {
					bg.Op("envelope.SetBindings").Call(
						j.Add(utils.ToCode(pc.PubMessagePromise.Target().BindingsStruct.RenderUsage(ctx))...).Values().Dot("MQTT()"),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}
