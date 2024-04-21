package proto

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type BaseProtoChannel struct {
	Parent          *render.Channel
	GolangNameProto string // Channel GolangName name concatenated with protocol name, e.g. Channel1Kafka
	Struct          *render.GoStruct
	ServerIface     *render.GoInterface

	ProtoName, ProtoTitle string
}

func (pc BaseProtoChannel) RenderCommonSubscriberMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderCommonSubscriberMethods", "proto", pc.ProtoName)

	rn := pc.Struct.ReceiverName()
	receiver := j.Id(rn).Id(pc.Struct.Name)
	var msgTyp common.GolangType = render.GoPointer{Type: pc.Parent.FallbackMessageType, DirectRender: true}
	if pc.Parent.SubMessagePromise != nil {
		msgTyp = render.GoPointer{Type: pc.Parent.SubMessagePromise.Target().InStruct, DirectRender: true}
	}

	return []*j.Statement{
		// Method ExtractEnvelope(envelope proto.EnvelopeReader, message *Message1In) error
		j.Func().Params(receiver.Clone()).Id("ExtractEnvelope").
			Params(
				j.Id("envelope").Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeReader"),
				j.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...),
			).
			Error().
			BlockFunc(func(bg *j.Group) {
				if pc.Parent.SubMessagePromise == nil {
					bg.Empty().Add(utils.QualSprintf(`
						enc := %Q(encoding/json,NewDecoder)(envelope)
						if err := enc.Decode(message); err != nil {
							return err
						}`))
				} else {
					bg.Op(fmt.Sprintf(`return message.Unmarshal%sEnvelope(envelope)`, pc.ProtoTitle))
				}
			}),

		// Method Subscriber() proto.Subscriber
		j.Func().Params(receiver.Clone()).Id("Subscriber").
			Params().
			Qual(ctx.RuntimeModule(pc.ProtoName), "Subscriber").
			Block(
				j.Return(j.Id(rn).Dot("subscriber")),
			),

		// Method Subscribe(ctx context.Context, cb func(envelope proto.EnvelopeReader)) error
		j.Func().Params(receiver.Clone()).Id("Subscribe").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("cb").Func().Params(j.Id("envelope").Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeReader")), // FIXME: *any on fallback variant
			).
			Error().
			Block(
				j.Return(j.Id(rn).Dot("subscriber.Receive(ctx, cb)")),
			),
	}
}

func (pc BaseProtoChannel) RenderCommonPublisherMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderCommonPublisherMethods", "proto", pc.ProtoName)

	rn := pc.Struct.ReceiverName()
	receiver := j.Id(rn).Id(pc.Struct.Name)

	return []*j.Statement{
		// Method Publisher() proto.Publisher
		j.Func().Params(receiver.Clone()).Id("Publisher").
			Params().
			Qual(ctx.RuntimeModule(pc.ProtoName), "Publisher").
			Block(
				j.Return(j.Id(rn).Dot("publisher")),
			),

		// Method Publish(ctx context.Context, envelopes ...proto.EnvelopeWriter) error
		j.Func().Params(receiver.Clone()).Id("Publish").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("envelopes").Op("...").Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeWriter"),
			).
			Error().
			Block(
				j.Return(j.Id(rn).Dot("publisher.Send(ctx, envelopes...)")),
			),
	}
}

func (pc BaseProtoChannel) RenderCommonMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderCommonMethods", "proto", pc.ProtoName)

	rn := pc.Struct.ReceiverName()
	receiver := j.Id(rn).Id(pc.Struct.Name)

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			Qual(ctx.RuntimeModule(""), "ParamString").
			Block(
				j.Return(j.Id(rn).Dot("name")),
			),

		// Method Close() (err error)
		j.Func().Params(receiver.Clone()).Id("Close").
			Params().
			Params(j.Err().Error()).
			BlockFunc(func(g *j.Group) {
				if pc.Parent.Publisher {
					g.If(j.Id(rn).Dot("publisher").Op("!=").Nil()).Block(
						j.Add(utils.QualSprintf("err = %Q(errors,Join)(err, %[1]s.publisher.Close())", rn)),
					)
				}
				if pc.Parent.Subscriber {
					g.If(j.Id(rn).Dot("subscriber").Op("!=").Nil()).Block(
						j.Add(utils.QualSprintf("err = %Q(errors,Join)(err, %[1]s.subscriber.Close())", rn)),
					)
				}
				g.Return()
			}),
	}
}

func (pc BaseProtoChannel) RenderOpenFunc(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderOpenFunc", "proto", pc.ProtoName)

	return []*j.Statement{
		// OpenChannel1Proto(ctx context.Context, params Channel1Parameters, servers ...channel1ProtoServer) (*Channel1Proto, error)
		j.Func().Id("Open"+pc.GolangNameProto).
			ParamsFunc(func(g *j.Group) {
				g.Id("ctx").Qual("context", "Context")
				if pc.Parent.ParametersStruct != nil {
					g.Id("params").Add(utils.ToCode(pc.Parent.ParametersStruct.RenderUsage(ctx))...)
				}
				g.Id("servers").Op("...").Add(utils.ToCode(pc.ServerIface.RenderUsage(ctx))...)
			}).
			Params(j.Op("*").Add(utils.ToCode(pc.Struct.RenderUsage(ctx))...), j.Error()).
			BlockFunc(func(bg *j.Group) {
				bg.Op("if len(servers) == 0").Block(j.Op("return nil, ").Qual(ctx.RuntimeModule(""), "ErrEmptyServers"))
				if pc.Parent.Publisher || pc.Parent.Subscriber {
					bg.Id("name").Op(":=").Id(pc.Parent.GolangName + "Name").CallFunc(func(g *j.Group) {
						if pc.Parent.ParametersStruct != nil {
							g.Id("params")
						}
					})
					if pc.Parent.BindingsStruct != nil {
						bg.Id("bindings").Op(":=").Id(pc.Parent.BindingsStruct.Name).Values().Dot(pc.ProtoTitle).Call()
					}
					if pc.Parent.Publisher {
						bg.Var().Id("prod").Index().Qual(ctx.RuntimeModule(pc.ProtoName), "Producer")
					}
					if pc.Parent.Subscriber {
						bg.Var().Id("cons").Index().Qual(ctx.RuntimeModule(pc.ProtoName), "Consumer")
					}
					bg.Op("for _, srv := range servers").BlockFunc(func(g *j.Group) {
						if pc.Parent.Publisher {
							g.Op(`
								if srv.Producer() != nil {
									prod = append(prod, srv.Producer())
								}`)
						}
						if pc.Parent.Subscriber {
							g.Op(`
								if srv.Consumer() != nil {
									cons = append(cons, srv.Consumer())
								}`)
						}
					})
				}
				if pc.Parent.Publisher {
					bg.Op("pubs, err := ").
						Qual(ctx.RuntimeModule(""), "GatherPublishers").
						Types(
							j.Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeWriter"),
							j.Qual(ctx.RuntimeModule(pc.ProtoName), "Publisher"),
							j.Qual(ctx.RuntimeModule(pc.ProtoName), "ChannelBindings"),
						).
						CallFunc(func(g *j.Group) {
							g.Id("ctx")
							g.Id("name")
							g.Id(lo.Ternary(pc.Parent.BindingsStruct != nil, "&bindings", "nil"))
							g.Id("prod")
						})
					bg.Op(`
						if err != nil {
							return nil, err
						}`)
					bg.Op("pub := ").Qual(ctx.RuntimeModule(""), "PublisherFanOut").
						Types(j.Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeWriter"), j.Qual(ctx.RuntimeModule(pc.ProtoName), "Publisher")).
						Op("{Publishers: pubs}")
				}
				if pc.Parent.Subscriber {
					bg.Op("subs, err := ").
						Qual(ctx.RuntimeModule(""), "GatherSubscribers").
						Types(
							j.Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeReader"),
							j.Qual(ctx.RuntimeModule(pc.ProtoName), "Subscriber"),
							j.Qual(ctx.RuntimeModule(pc.ProtoName), "ChannelBindings"),
						).
						CallFunc(func(g *j.Group) {
							g.Id("ctx")
							g.Id("name")
							g.Id(lo.Ternary(pc.Parent.BindingsStruct != nil, "&bindings", "nil"))
							g.Id("cons")
						})
					bg.Op("if err != nil").BlockFunc(func(g *j.Group) {
						if pc.Parent.Publisher {
							g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, pub.Close())"))
						}
						g.Op("return nil, err")
					})
					bg.Op("sub := ").Qual(ctx.RuntimeModule(""), "SubscriberFanIn").
						Types(j.Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeReader"), j.Qual(ctx.RuntimeModule(pc.ProtoName), "Subscriber")).
						Op("{Subscribers: subs}")
				}
				bg.Op("ch := ").Id(pc.Struct.NewFuncName()).CallFunc(func(g *j.Group) {
					if pc.Parent.ParametersStruct != nil {
						g.Id("params")
					}
					if pc.Parent.Publisher {
						g.Id("pub")
					}
					if pc.Parent.Subscriber {
						g.Id("sub")
					}
				})
				bg.Op("return ch, nil")
			}),
	}
}
