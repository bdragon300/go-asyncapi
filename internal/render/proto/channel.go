package proto

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type BaseProtoChannel struct {
	Name            string // TODO: move fields to abstract channel
	Publisher       bool
	Subscriber      bool
	Struct          *render.Struct
	ServerIface     *render.Interface
	AbstractChannel *render.Channel

	PubMessagePromise   *render.Promise[*render.Message] // nil when message is not set
	SubMessagePromise   *render.Promise[*render.Message] // nil when message is not set
	FallbackMessageType common.GolangType

	ProtoName, ProtoAbbr string
}

func (pc BaseProtoChannel) RenderCommonSubscriberMethods(
	ctx *common.RenderContext,
	channelStruct *render.Struct,
	subMessagePromise *render.Promise[*render.Message],
	fallbackMessageType common.GolangType,
) []*j.Statement {
	rn := channelStruct.ReceiverName()
	receiver := j.Id(rn).Id(channelStruct.Name)
	var msgTyp common.GolangType = render.Pointer{Type: fallbackMessageType, DirectRender: true}
	if subMessagePromise != nil {
		msgTyp = render.Pointer{Type: subMessagePromise.Target().InStruct, DirectRender: true}
	}

	return []*j.Statement{
		// Method ExtractEnvelope(envelope proto.EnvelopeReader, message *Message1In) error
		j.Func().Params(receiver.Clone()).Id("ExtractEnvelope").
			Params(
				j.Id("envelope").Qual(ctx.RuntimePackage(pc.ProtoName), "EnvelopeReader"),
				j.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...),
			).
			Error().
			BlockFunc(func(bg *j.Group) {
				if subMessagePromise == nil {
					bg.Empty().Add(utils.QualSprintf(`
						enc := %Q(encoding/json,NewDecoder)(envelope)
						if err := enc.Decode(message); err != nil {
							return err
						}`))
				} else {
					bg.Op(fmt.Sprintf(`return message.Unmarshal%sEnvelope(envelope)`, pc.ProtoAbbr))
				}
			}),

		// Method Subscriber() proto.Subscriber
		j.Func().Params(receiver.Clone()).Id("Subscriber").
			Params().
			Qual(ctx.RuntimePackage(pc.ProtoName), "Subscriber").
			Block(
				j.Return(j.Id(rn).Dot("subscriber")),
			),

		// Method Subscribe(ctx context.Context, cb func(envelope proto.EnvelopeReader) error) error
		j.Func().Params(receiver.Clone()).Id("Subscribe").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("cb").Func().Params(j.Id("envelope").Qual(ctx.RuntimePackage(pc.ProtoName), "EnvelopeReader")).Error(), // FIXME: *any on fallback variant
			).
			Error().
			Block(
				j.Return(j.Id(rn).Dot("subscriber.Receive(ctx, cb)")),
			),
	}
}

func (pc BaseProtoChannel) RenderCommonPublisherMethods(
	ctx *common.RenderContext,
	channelStruct *render.Struct,
) []*j.Statement {
	rn := channelStruct.ReceiverName()
	receiver := j.Id(rn).Id(channelStruct.Name)

	return []*j.Statement{
		// Method Publisher() proto.Publisher
		j.Func().Params(receiver.Clone()).Id("Publisher").
			Params().
			Qual(ctx.RuntimePackage(pc.ProtoName), "Publisher").
			Block(
				j.Return(j.Id(rn).Dot("publisher")),
			),

		// Method Publish(ctx context.Context, envelopes ...proto.EnvelopeWriter) error
		j.Func().Params(receiver.Clone()).Id("Publish").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("envelopes").Op("...").Qual(ctx.RuntimePackage(pc.ProtoName), "EnvelopeWriter"),
			).
			Error().
			Block(
				j.Return(j.Id(rn).Dot("publisher.Send(ctx, envelopes...)")),
			),
	}
}

func (pc BaseProtoChannel) RenderCommonMethods(
	ctx *common.RenderContext,
	channelStruct *render.Struct,
	publisher, subscriber bool,
) []*j.Statement {
	rn := channelStruct.ReceiverName()
	receiver := j.Id(rn).Id(channelStruct.Name)

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			Qual(ctx.RuntimePackage(""), "ParamString").
			Block(
				j.Return(j.Id(rn).Dot("name")),
			),

		// Method Close() (err error)
		j.Func().Params(receiver.Clone()).Id("Close").
			Params().
			Params(j.Error()).
			BlockFunc(func(g *j.Group) {
				var args []j.Code
				if publisher {
					args = append(args, j.Id(rn).Dot("publisher").Dot("Close").Call())
				}
				if subscriber {
					args = append(args, j.Id(rn).Dot("subscriber").Dot("Close").Call())
				}
				g.Return(j.Qual("errors", "Join").Call(args...))
			}),
	}
}

func (pc BaseProtoChannel) RenderOpenFunc(
	ctx *common.RenderContext,
	channelStruct *render.Struct,
	channelName string,
	serverIface *render.Interface,
	parametersStruct, bindingsStruct *render.Struct,
	publisher, subscriber bool,
) []*j.Statement {
	return []*j.Statement{
		// OpenChannel1Proto(params Channel1Parameters, servers ...channel1ProtoServer) (*Channel1Proto, error)
		j.Func().Id("Open"+channelStruct.Name).
			ParamsFunc(func(g *j.Group) {
				if parametersStruct != nil {
					g.Id("params").Add(utils.ToCode(parametersStruct.RenderUsage(ctx))...)
				}
				g.Id("servers").Op("...").Add(utils.ToCode(serverIface.RenderUsage(ctx))...)
			}).
			Params(j.Op("*").Add(utils.ToCode(channelStruct.RenderUsage(ctx))...), j.Error()).
			BlockFunc(func(bg *j.Group) {
				bg.Op("if len(servers) == 0").Block(j.Op("return nil, ").Qual(ctx.RuntimePackage(""), "ErrEmptyServers"))
				if publisher || subscriber {
					bg.Id("name").Op(":=").Id(utils.ToGolangName(channelName, true) + "Name").CallFunc(func(g *j.Group) {
						if parametersStruct != nil {
							g.Id("params")
						}
					})
					if bindingsStruct != nil {
						bg.Id("bindings").Op(":=").Id(bindingsStruct.Name).Values().Dot(pc.ProtoAbbr).Call()
					}
					if publisher {
						bg.Var().Id("prod").Index().Qual(ctx.RuntimePackage(pc.ProtoName), "Producer")
					}
					if subscriber {
						bg.Var().Id("cons").Index().Qual(ctx.RuntimePackage(pc.ProtoName), "Consumer")
					}
					bg.Op("for _, srv := range servers").BlockFunc(func(g *j.Group) {
						if publisher {
							g.Op("prod = append(prod, srv.Producer())")
						}
						if subscriber {
							g.Op("cons = append(cons, srv.Consumer())")
						}
					})
				}
				if publisher {
					bg.Op("pubs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherPublishers").
						Types(
							j.Qual(ctx.RuntimePackage(pc.ProtoName), "EnvelopeWriter"),
							j.Qual(ctx.RuntimePackage(pc.ProtoName), "Publisher"),
							j.Qual(ctx.RuntimePackage(pc.ProtoName), "ChannelBindings"),
						).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(bindingsStruct != nil, "&bindings", "nil"))
							g.Id("prod")
						})
					bg.Op(`
						if err != nil {
							return nil, err
						}`)
					bg.Op("pub := ").Qual(ctx.RuntimePackage(""), "PublisherFanOut").
						Types(j.Qual(ctx.RuntimePackage(pc.ProtoName), "EnvelopeWriter"), j.Qual(ctx.RuntimePackage(pc.ProtoName), "Publisher")).
						Op("{Publishers: pubs}")
				}
				if subscriber {
					bg.Op("subs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherSubscribers").
						Types(
							j.Qual(ctx.RuntimePackage(pc.ProtoName), "EnvelopeReader"),
							j.Qual(ctx.RuntimePackage(pc.ProtoName), "Subscriber"),
							j.Qual(ctx.RuntimePackage(pc.ProtoName), "ChannelBindings"),
						).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(bindingsStruct != nil, "&bindings", "nil"))
							g.Id("cons")
						})
					bg.Op("if err != nil").BlockFunc(func(g *j.Group) {
						if publisher {
							g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, pub.Close())"))
						}
						g.Op("return nil, err")
					})
					bg.Op("sub := ").Qual(ctx.RuntimePackage(""), "SubscriberFanIn").
						Types(j.Qual(ctx.RuntimePackage(pc.ProtoName), "EnvelopeReader"), j.Qual(ctx.RuntimePackage(pc.ProtoName), "Subscriber")).
						Op("{Subscribers: subs}")
				}
				bg.Op("ch := ").Id(channelStruct.NewFuncName()).CallFunc(func(g *j.Group) {
					g.Id("params")
					if publisher {
						g.Id("pub")
					}
					if subscriber {
						g.Id("sub")
					}
				})
				bg.Op("return ch, nil")
			}),
	}
}