package protocols

import (
	"fmt"
	"path"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func BuildChannel(
	ctx *common.CompileContext,
	channel *asyncapi.Channel,
	channelKey string,
	protoName, protoAbbr string,
) (*BaseProtoChannel, error) {
	paramsLnk := render.NewListCbPromise[*render.Parameter](func(item common.Renderer, path []string) bool {
		par, ok := item.(*render.Parameter)
		if !ok {
			return false
		}
		_, ok = channel.Parameters.Get(par.Name)
		return ok
	})
	ctx.PutListPromise(paramsLnk)

	chanResult := &BaseProtoChannel{
		Name: channelKey,
		Struct: &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(channelKey, protoAbbr),
				Description:  channel.Description,
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
			Fields: []render.StructField{
				{Name: "name", Type: &render.Simple{Name: "ParamString", Package: ctx.RuntimePackage("")}},
			},
		},
		FallbackMessageType: &render.Simple{Name: "any", IsIface: true},
	}

	// FIXME: remove in favor of the non-proto channel
	if channel.Parameters.Len() > 0 {
		ctx.Logger.Trace("Channel parameters", "proto", protoName)
		ctx.Logger.NextCallLevel()
		chanResult.ParametersStructNoRender = &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(channelKey, "Parameters"),
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
			Fields: nil,
		}
		for _, paramName := range channel.Parameters.Keys() {
			ctx.Logger.Trace("Channel parameter", "name", paramName, "proto", protoName)
			ref := path.Join(ctx.PathRef(), "parameters", paramName)
			lnk := render.NewGolangTypePromise(ref, common.PromiseOriginInternal)
			ctx.PutPromise(lnk)
			chanResult.ParametersStructNoRender.Fields = append(chanResult.ParametersStructNoRender.Fields, render.StructField{
				Name: utils.ToGolangName(paramName, true),
				Type: lnk,
			})
		}
		ctx.Logger.PrevCallLevel()
	}

	// Interface to match servers bound with a channel
	var ifaceFirstMethodParams []render.FuncParam
	if chanResult.ParametersStructNoRender != nil {
		ifaceFirstMethodParams = append(ifaceFirstMethodParams, render.FuncParam{
			Name: "params",
			Type: &render.Simple{Name: chanResult.ParametersStructNoRender.Name, Package: ctx.TopPackageName()},
		})
	}
	chanResult.ServerIface = &render.Interface{
		BaseType: render.BaseType{
			Name:         utils.ToLowerFirstLetter(chanResult.Struct.Name + "Server"),
			DirectRender: true,
			PackageName:  ctx.TopPackageName(),
		},
		Methods: []render.FuncSignature{
			{
				Name: "Open" + chanResult.Struct.Name,
				Args: ifaceFirstMethodParams,
				Return: []render.FuncParam{
					{Type: &render.Simple{Name: chanResult.Struct.Name, Package: ctx.TopPackageName()}, Pointer: true},
					{Type: &render.Simple{Name: "error"}},
				},
			},
		},
	}

	// Publisher stuff
	if channel.Publish != nil {
		ctx.Logger.Trace("Channel publish operation", "proto", protoName)
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, render.StructField{
			Name:        "publisher",
			Description: channel.Publish.Description,
			Type: &render.Simple{
				Name:    "Publisher",
				Package: ctx.RuntimePackage(protoName),
				IsIface: true,
			},
		})
		chanResult.Publisher = true
		if channel.Publish.Message != nil {
			ctx.Logger.Trace("Channel publish operation message", "proto", protoName)
			ref := path.Join(ctx.PathRef(), "publish/message")
			chanResult.PubMessageLink = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(chanResult.PubMessageLink)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, render.FuncSignature{
			Name: "Producer",
			Args: nil,
			Return: []render.FuncParam{
				{Type: &render.Simple{Name: "Producer", Package: ctx.RuntimePackage(protoName), IsIface: true}},
			},
		})
	}

	// Subscriber stuff
	if channel.Subscribe != nil {
		ctx.Logger.Trace("Channel subscribe operation", "proto", protoName)
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, render.StructField{
			Name:        "subscriber",
			Description: channel.Subscribe.Description,
			Type: &render.Simple{
				Name:    "Subscriber",
				Package: ctx.RuntimePackage(protoName),
				IsIface: true,
			},
		})
		chanResult.Subscriber = true
		if channel.Subscribe.Message != nil {
			ctx.Logger.Trace("Channel subscribe operation message", "proto", protoName)
			ref := path.Join(ctx.PathRef(), "subscribe/message")
			chanResult.SubMessageLink = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(chanResult.SubMessageLink)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, render.FuncSignature{
			Name: "Consumer",
			Args: nil,
			Return: []render.FuncParam{
				{Type: &render.Simple{Name: "Consumer", Package: ctx.RuntimePackage(protoName), IsIface: true}},
			},
		})
	}

	return chanResult, nil
}

type BaseProtoChannel struct {
	Name                     string
	Publisher                bool
	Subscriber               bool
	Struct                   *render.Struct
	ServerIface              *render.Interface
	ParametersStructNoRender *render.Struct // nil if parameters not set

	PubMessageLink      *render.Link[*render.Message] // nil when message is not set
	SubMessageLink      *render.Link[*render.Message] // nil when message is not set
	FallbackMessageType common.GolangType
}

func RenderChannelSubscriberMethods(
	ctx *common.RenderContext,
	channelStruct *render.Struct,
	subMessageLink *render.Link[*render.Message],
	fallbackMessageType common.GolangType,
	protoName, protoAbbr string,
) []*j.Statement {
	rn := channelStruct.ReceiverName()
	receiver := j.Id(rn).Id(channelStruct.Name)
	var msgTyp common.GolangType = render.Pointer{Type: fallbackMessageType, DirectRender: true}
	if subMessageLink != nil {
		msgTyp = render.Pointer{Type: subMessageLink.Target().InStruct, DirectRender: true}
	}

	return []*j.Statement{
		// Method ExtractEnvelope(envelope proto.EnvelopeReader, message *Message1In) error
		j.Func().Params(receiver.Clone()).Id("ExtractEnvelope").
			Params(
				j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeReader"),
				j.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...),
			).
			Error().
			BlockFunc(func(bg *j.Group) {
				if subMessageLink == nil {
					bg.Empty().Add(utils.QualSprintf(`
						enc := %Q(encoding/json,NewDecoder)(envelope)
						if err := enc.Decode(message); err != nil {
							return err
						}`))
				} else {
					bg.Op(fmt.Sprintf(`return message.Unmarshal%sEnvelope(envelope)`, protoAbbr))
				}
			}),

		// Method Subscriber() proto.Subscriber
		j.Func().Params(receiver.Clone()).Id("Subscriber").
			Params().
			Qual(ctx.RuntimePackage(protoName), "Subscriber").
			Block(
				j.Return(j.Id(rn).Dot("subscriber")),
			),

		// Method Subscribe(ctx context.Context, cb func(envelope proto.EnvelopeReader) error) error
		j.Func().Params(receiver.Clone()).Id("Subscribe").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("cb").Func().Params(j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeReader")).Error(), // FIXME: *any on fallback variant
			).
			Error().
			Block(
				j.Return(j.Id(rn).Dot("subscriber.Receive(ctx, cb)")),
			),
	}
}

func RenderChannelPublisherMethods(
	ctx *common.RenderContext,
	channelStruct *render.Struct,
	protoName string,
) []*j.Statement {
	rn := channelStruct.ReceiverName()
	receiver := j.Id(rn).Id(channelStruct.Name)

	return []*j.Statement{
		// Method Publisher() proto.Publisher
		j.Func().Params(receiver.Clone()).Id("Publisher").
			Params().
			Qual(ctx.RuntimePackage(protoName), "Publisher").
			Block(
				j.Return(j.Id(rn).Dot("publisher")),
			),

		// Method Publish(ctx context.Context, envelopes ...proto.EnvelopeWriter) error
		j.Func().Params(receiver.Clone()).Id("Publish").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("envelopes").Op("...").Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter"),
			).
			Error().
			Block(
				j.Return(j.Id(rn).Dot("publisher.Send(ctx, envelopes...)")),
			),
	}
}

func RenderChannelCommonMethods(ctx *common.RenderContext, channelStruct *render.Struct, publisher, subscriber bool) []*j.Statement {
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

func RenderChannelOpenFunc(
	ctx *common.RenderContext,
	channelStruct *render.Struct,
	channelName string,
	serverIface *render.Interface,
	parametersStructNoRender, bindingsStructNoRender *render.Struct,
	publisher, subscriber bool,
	protoName, protoAbbr string,
) []*j.Statement {
	return []*j.Statement{
		// OpenChannel1Proto(params Channel1Parameters, servers ...channel1ProtoServer) (*Channel1Proto, error)
		j.Func().Id("Open"+channelStruct.Name).
			ParamsFunc(func(g *j.Group) {
				if parametersStructNoRender != nil {
					g.Id("params").Add(utils.ToCode(parametersStructNoRender.RenderUsage(ctx))...)
				}
				g.Id("servers").Op("...").Add(utils.ToCode(serverIface.RenderUsage(ctx))...)
			}).
			Params(j.Op("*").Add(utils.ToCode(channelStruct.RenderUsage(ctx))...), j.Error()).
			BlockFunc(func(bg *j.Group) {
				bg.Op("if len(servers) == 0").Block(j.Op("return nil, ").Qual(ctx.RuntimePackage(""), "ErrEmptyServers"))
				bg.Id("name").Op(":=").Id(utils.ToGolangName(channelName, true) + "Name").CallFunc(func(g *j.Group) {
					if parametersStructNoRender != nil {
						g.Id("params")
					}
				})
				if bindingsStructNoRender != nil {
					bg.Id("bindings").Op(":=").Id(bindingsStructNoRender.Name).Values().Dot(protoAbbr).Call()
				}
				if publisher {
					bg.Var().Id("prod").Index().Qual(ctx.RuntimePackage(protoName), "Producer")
				}
				if subscriber {
					bg.Var().Id("cons").Index().Qual(ctx.RuntimePackage(protoName), "Consumer")
				}
				bg.Op("for _, srv := range servers").BlockFunc(func(g *j.Group) {
					if publisher {
						g.Op("prod = append(prod, srv.Producer())")
					}
					if subscriber {
						g.Op("cons = append(cons, srv.Consumer())")
					}
				})
				if publisher {
					bg.Op("pubs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherPublishers").
						Types(
							j.Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter"),
							j.Qual(ctx.RuntimePackage(protoName), "Publisher"),
							j.Qual(ctx.RuntimePackage(protoName), "ChannelBindings"),
						).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(bindingsStructNoRender != nil, "&bindings", "nil"))
							g.Id("prod")
						})
					bg.Op(`
						if err != nil {
							return nil, err
						}`)
					bg.Op("pub := ").Qual(ctx.RuntimePackage(""), "PublisherFanOut").
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter"), j.Qual(ctx.RuntimePackage(protoName), "Publisher")).
						Op("{Publishers: pubs}")
				}
				if subscriber {
					bg.Op("subs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherSubscribers").
						Types(
							j.Qual(ctx.RuntimePackage(protoName), "EnvelopeReader"),
							j.Qual(ctx.RuntimePackage(protoName), "Subscriber"),
							j.Qual(ctx.RuntimePackage(protoName), "ChannelBindings"),
						).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(bindingsStructNoRender != nil, "&bindings", "nil"))
							g.Id("cons")
						})
					bg.Op("if err != nil").BlockFunc(func(g *j.Group) {
						if publisher {
							g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, pub.Close())"))
						}
						g.Op("return nil, err")
					})
					bg.Op("sub := ").Qual(ctx.RuntimePackage(""), "SubscriberFanIn").
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeReader"), j.Qual(ctx.RuntimePackage(protoName), "Subscriber")).
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

func ChannelBindingsMethodBody(
	values *render.StructInit,
	publisherJSONValues *types.OrderedMap[string, any],
	subscriberJSONValues *types.OrderedMap[string, any],
) func(ctx *common.RenderContext, p *render.Func) []*j.Statement {
	return func(ctx *common.RenderContext, p *render.Func) []*j.Statement {
		var res []*j.Statement
		res = append(res,
			j.Id("b").Op(":=").Add(utils.ToCode(values.RenderInit(ctx))...),
		)
		if publisherJSONValues != nil {
			for _, e := range subscriberJSONValues.Entries() {
				n := utils.ToLowerFirstLetter(e.Key)
				res = append(res,
					j.Id(n).Op(":=").Lit(e.Value),
					j.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.SubscriberBindings.%[2]s)", n, e.Key)),
				)
			}
		}
		if subscriberJSONValues != nil {
			for _, e := range publisherJSONValues.Entries() {
				n := utils.ToLowerFirstLetter(e.Key)
				res = append(res,
					j.Id(n).Op(":=").Lit(e.Value),
					j.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.PublisherBindings.%[2]s)", n, e.Key)),
				)
			}
		}
		res = append(res, j.Return(j.Id("b")))
		return res
	}
}
