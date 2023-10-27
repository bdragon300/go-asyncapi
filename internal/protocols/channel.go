package protocols

import (
	"fmt"
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func BuildChannel(
	ctx *common.CompileContext,
	channel *compile.Channel,
	channelKey string,
	protoName, protoAbbr string,
) (*BaseProtoChannel, error) {
	paramsLnk := assemble.NewListCbLink[*assemble.Parameter](func(item common.Assembler, path []string) bool {
		par, ok := item.(*assemble.Parameter)
		if !ok {
			return false
		}
		_, ok = channel.Parameters.Get(par.Name)
		return ok
	})
	ctx.Linker.AddMany(paramsLnk)

	chanResult := &BaseProtoChannel{
		Name: channelKey,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", protoAbbr),
				Description: channel.Description,
				Render:      true,
				Package:     ctx.TopPackageName(),
			},
			Fields: []assemble.StructField{
				{Name: "name", Type: &assemble.Simple{Type: "ParamString", Package: ctx.RuntimePackage("")}},
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

	return chanResult, nil
}

type BaseProtoChannel struct {
	Name                       string
	Publisher                  bool
	Subscriber                 bool
	Struct                     *assemble.Struct
	ServerIface                *assemble.Interface
	ParametersStructNoAssemble *assemble.Struct // nil if parameters not set

	PubMessageLink      *assemble.Link[*assemble.Message] // nil when message is not set
	SubMessageLink      *assemble.Link[*assemble.Message] // nil when message is not set
	FallbackMessageType common.Assembler
}

func AssembleChannelSubscriberMethods(
	ctx *common.AssembleContext,
	channelStruct *assemble.Struct,
	subMessageLink *assemble.Link[*assemble.Message],
	fallbackMessageType common.Assembler,
	protoName, protoAbbr string,
) []*j.Statement {
	rn := channelStruct.ReceiverName()
	receiver := j.Id(rn).Id(channelStruct.Name)
	msgTyp := fallbackMessageType
	if subMessageLink != nil {
		msgTyp = subMessageLink.Target().InStruct
	}

	return []*j.Statement{
		// Method ExtractEnvelope(envelope proto.EnvelopeReader, message proto.EnvelopeUnmarshaler) error
		j.Func().Params(receiver.Clone()).Id("ExtractEnvelope").
			Params(
				j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeReader"),
				j.Id("message").Qual(ctx.RuntimePackage(protoName), "EnvelopeUnmarshaler"),
			).
			Error().
			Block(
				j.Op(fmt.Sprintf(`return message.Unmarshal%sEnvelope(envelope)`, protoAbbr)),
			),

		// Method Subscriber() proto.Subscriber
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
								return cb(buf)`, rn))
						}),
				)),
			),
	}
}

func AssembleChannelPublisherMethods(
	ctx *common.AssembleContext,
	channelStruct *assemble.Struct,
	pubMessageLink *assemble.Link[*assemble.Message],
	fallbackMessageType common.Assembler,
	protoName, protoAbbr string,
) []*j.Statement {
	rn := channelStruct.ReceiverName()
	receiver := j.Id(rn).Id(channelStruct.Name)
	msgTyp := fallbackMessageType
	var msgBindings *assemble.Struct
	if pubMessageLink != nil {
		msgTyp = pubMessageLink.Target().OutStruct
		if pubMessageLink.Target().BindingsStruct != nil {
			msgBindings = pubMessageLink.Target().BindingsStruct
		}
	}

	return []*j.Statement{
		// Method Publisher() proto.Publisher
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
				j.Id("msgs").Op("...").Op("*").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...), // FIXME: *any on fallback variant
			).
			Error().
			BlockFunc(func(blockGroup *j.Group) {
				call := "MakeEnvelope(buf, msgs[i])"
				if msgBindings != nil {
					blockGroup.Op("bindings :=").Add(utils.ToCode(msgBindings.AssembleUsage(ctx))...).Values().Dot(protoAbbr + "()")
					call = "MakeEnvelope(buf, msgs[i], bindings)"
				}
				// TODO: kafka.NewEnvelopeOut() depends on selected implementation
				blockGroup.Add(utils.QualSprintf(`
					envelopes := make([]%Q(%[2]s,EnvelopeWriter), 0, len(msgs))
					for i := 0; i < len(msgs); i++ {
						buf := %Q(%[2]s,NewEnvelopeOut)()
						if err := %[1]s.%[3]s; err != nil {
							return %Q(fmt,Errorf)("make envelope #%%d error: %%w", i, err)
						}
						envelopes = append(envelopes, buf)
					}
					return %[1]s.publisher.Send(ctx, envelopes...)`, rn, ctx.RuntimePackage(protoName), call),
				)
			}),
	}
}

func AssembleChannelCommonMethods(
	ctx *common.AssembleContext,
	channelStruct *assemble.Struct,
	publisher, subscriber bool,
	protoAbbr string,
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

		// Protocol() run.Protocol
		j.Func().Params(receiver.Clone()).Id("Protocol").
			Params().
			Qual(ctx.RuntimePackage(""), "Protocol").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage(""), "Protocol"+protoAbbr)),
			),

		// Method Close() (err error)
		j.Func().Params(receiver.Clone()).Id("Close").
			Params().
			Params(j.Err().Error()).
			BlockFunc(func(g *j.Group) {
				if publisher {
					g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, %[1]s.publisher.Close())", rn))
				}
				if subscriber {
					g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, %[1]s.subscriber.Close())", rn))
				}
				g.Return()
			}),
	}
}

func AssembleChannelOpenFunc(
	ctx *common.AssembleContext,
	channelStruct *assemble.Struct,
	channelName string,
	serverIface *assemble.Interface,
	parametersStructNoAssemble, bindingsStructNoAssemble *assemble.Struct,
	publisher, subscriber bool,
	protoName, protoAbbr string,
) []*j.Statement {
	return []*j.Statement{
		// OpenChannel1Proto(params Channel1Parameters, servers ...channel1ProtoServer) (*Channel1Proto, error)
		j.Func().Id("Open"+channelStruct.Name).
			ParamsFunc(func(g *j.Group) {
				if parametersStructNoAssemble != nil {
					g.Id("params").Add(utils.ToCode(parametersStructNoAssemble.AssembleUsage(ctx))...)
				}
				g.Id("servers").Op("...").Add(utils.ToCode(serverIface.AssembleUsage(ctx))...)
			}).
			Params(j.Op("*").Add(utils.ToCode(channelStruct.AssembleUsage(ctx))...), j.Error()).
			BlockFunc(func(bg *j.Group) {
				bg.Op("if len(servers) == 0").Block(j.Op("return nil, ").Qual(ctx.RuntimePackage(""), "ErrEmptyServers"))
				bg.Id("name").Op(":=").Id(utils.ToGolangName(channelName, true) + "Name").CallFunc(func(g *j.Group) {
					if parametersStructNoAssemble != nil {
						g.Id("params")
					}
				})
				if bindingsStructNoAssemble != nil {
					bg.Id("bindings").Op(":=").Id(bindingsStructNoAssemble.Name).Values().Dot(protoAbbr).Call()
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
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter"), j.Qual(ctx.RuntimePackage(protoName), "ChannelBindings")).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(bindingsStructNoAssemble != nil, "&bindings", "nil"))
							g.Id("prod")
						})
					bg.Op(`
						if err != nil {
							return nil, err
						}`)
					bg.Op("pub := ").Qual(ctx.RuntimePackage(""), "PublisherFanOut").
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter")).
						Op("{Publishers: pubs}")
				}
				if subscriber {
					bg.Op("subs, err := ").
						Qual(ctx.RuntimePackage(""), "GatherSubscribers").
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeReader"), j.Qual(ctx.RuntimePackage(protoName), "ChannelBindings")).
						CallFunc(func(g *j.Group) {
							g.Id("name")
							g.Id(lo.Ternary(bindingsStructNoAssemble != nil, "&bindings", "nil"))
							g.Id("cons")
						})
					bg.Op("if err != nil").BlockFunc(func(g *j.Group) {
						if publisher {
							g.Add(utils.QualSprintf("err = %Q(errors,Join)(err, pub.Close())"))
						}
						g.Op("return nil, err")
					})
					bg.Op("sub := ").Qual(ctx.RuntimePackage(""), "SubscriberFanIn").
						Types(j.Qual(ctx.RuntimePackage(protoName), "EnvelopeReader")).
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
	values *assemble.StructInit,
	publisherJSONValues *utils.OrderedMap[string, any],
	subscriberJSONValues *utils.OrderedMap[string, any],
) func(ctx *common.AssembleContext, p *assemble.Func) []*j.Statement {
	return func(ctx *common.AssembleContext, p *assemble.Func) []*j.Statement {
		var res []*j.Statement
		res = append(res,
			j.Id("b").Op(":=").Add(utils.ToCode(values.AssembleInit(ctx))...),
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
