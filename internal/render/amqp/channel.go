package amqp

//func (pc ProtoChannel) Selectable() bool {
//	return true
//}

//func (pc ProtoChannel) D(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Channel", "", pc.Parent.Name, "definition", pc.Selectable(), "proto", pc.ProtoName)
//	defer ctx.LogFinishRender()
//
//	var res []*j.Statement
//	//res = append(res, pc.ServerIface.D(ctx)...)
//	//res = append(res, pc.RenderOpenFunc(ctx)...)
//	res = append(res, pc.renderNewFunc(ctx)...)
//	res = append(res, pc.Struct.D(ctx)...)
//	res = append(res, pc.RenderCommonMethods(ctx)...)
//	res = append(res, pc.renderProtoMethods(ctx)...)
//	if pc.Parent.Publisher {
//		res = append(res, pc.RenderCommonPublisherMethods(ctx)...)
//		res = append(res, pc.renderProtoPublisherMethods(ctx)...)
//	}
//	if pc.Parent.Subscriber {
//		res = append(res, pc.RenderCommonSubscriberMethods(ctx)...)
//	}
//	return res
//}

//func (pc ProtoChannel) U(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Channel", "", pc.Parent.Name, "usage", pc.Selectable(), "proto", pc.ProtoName)
//	defer ctx.LogFinishRender()
//	return pc.Struct.U(ctx)
//}

//func (pc ProtoChannel) ID() string {
//	return pc.Parent.Name
//}
//
//func (pc ProtoChannel) String() string {
//	return "AMQP ProtoChannel " + pc.Parent.Name
//}

//func (pc ProtoChannel) renderNewFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderNewFunc", "proto", pc.ProtoName)
//	return []*j.Statement{
//		// NewChannel1Proto(params Channel1Parameters, publisher proto.Publisher, subscriber proto.Subscriber) *Channel1Proto
//		j.Func().Id(pc.Struct.NewFuncName()).
//			ParamsFunc(func(g *j.Group) {
//				if pc.Parent.ParametersType != nil {
//					g.Id("params").Add(utils.ToCode(pc.Parent.ParametersType.U(ctx))...)
//				}
//				if pc.Parent.Publisher {
//					g.Id("publisher").Qual(ctx.RuntimeModule(pc.ProtoName), "Publisher")
//				}
//				if pc.Parent.Subscriber {
//					g.Id("subscriber").Qual(ctx.RuntimeModule(pc.ProtoName), "Subscriber")
//				}
//			}).
//			Op("*").Add(utils.ToCode(pc.Struct.U(ctx))...).
//			BlockFunc(func(bg *j.Group) {
//				bg.Op("res := ").Add(utils.ToCode(pc.Struct.U(ctx))...).Values(j.DictFunc(func(d j.Dict) {
//					d[j.Id("name")] = j.Id(pc.Parent.TypeNamePrefix + "Name").CallFunc(func(g *j.Group) {
//						if pc.Parent.ParametersType != nil {
//							g.Id("params")
//						}
//					})
//					if pc.Parent.Publisher {
//						d[j.Id("publisher")] = j.Id("publisher")
//					}
//					if pc.Parent.Subscriber {
//						d[j.Id("subscriber")] = j.Id("subscriber")
//					}
//				}))
//
//				if pc.Parent.BindingsType != nil {
//					bg.Id("bindings").Op(":=").Add(
//						utils.ToCode(pc.Parent.BindingsType.U(ctx))...,
//					).Values().Dot(pc.ProtoTitle).Call()
//					bg.Switch(j.Id("bindings.ChannelType")).BlockFunc(func(bg2 *j.Group) {
//						bg2.Case(j.Qual(ctx.RuntimeModule(pc.ProtoName), "ChannelTypeQueue")).Block(
//							j.Id("res.queue").Op("=").Op("res.name.String()"),
//						)
//						bg2.Default().Block(
//							j.Id("res.routingKey").Op("=").Op("res.name.String()"),
//						)
//					})
//					bg.Op(`
//						if bindings.ExchangeConfiguration.Name != nil {
//							res.exchange = *bindings.ExchangeConfiguration.Name
//						}
//						if bindings.QueueConfiguration.Name != "" {
//							res.queue = bindings.QueueConfiguration.Name
//						}`)
//				}
//				bg.Op(`return &res`)
//			}),
//	}
//}

//func (pc ProtoChannel) renderProtoMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderProtoMethods", "proto", pc.ProtoName)
//	rn := pc.Struct.ReceiverName()
//	receiver := j.Id(rn).Id(pc.Struct.Name)
//
//	return []*j.Statement{
//		// Method Exchange() string
//		j.Func().Params(receiver.Clone()).Id("Exchange").
//			Params().
//			String().
//			Block(
//				j.Return(j.Id(rn).Dot("exchange")),
//			),
//
//		// Method Queue() string
//		j.Func().Params(receiver.Clone()).Id("Queue").
//			Params().
//			String().
//			Block(
//				j.Return(j.Id(rn).Dot("queue")),
//			),
//
//		// Method RoutingKey() string
//		j.Func().Params(receiver.Clone()).Id("RoutingKey").
//			Params().
//			String().
//			Block(
//				j.Return(j.Id(rn).Dot("routingKey")),
//			),
//	}
//}

//func (pc ProtoChannel) renderProtoPublisherMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderProtoPublisherMethods", "proto", pc.ProtoName)
//	rn := pc.Struct.ReceiverName()
//	receiver := j.Id(rn).Id(pc.Struct.Name)
//
//	var msgTyp common.GolangType = render.GoPointer{Type: pc.FallbackMessageType, HasDefinition: true}
//	if pc.PublisherMessageTypePromise != nil {
//		msgTyp = render.GoPointer{Type: pc.PublisherMessageTypePromise.Target().OutType, HasDefinition: true}
//	}
//
//	return []*j.Statement{
//		// Method SealEnvelope(envelope proto.EnvelopeWriter, message *Message1Out) error
//		j.Func().Params(receiver.Clone()).Id("SealEnvelope").
//			ParamsFunc(func(g *j.Group) {
//				g.Id("envelope").Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeWriter")
//				g.Id("message").Add(utils.ToCode(msgTyp.U(ctx))...)
//			}).
//			Error().
//			BlockFunc(func(bg *j.Group) {
//				bg.Op("envelope.ResetPayload()")
//				if pc.PublisherMessageTypePromise == nil { // No Message set for Channel in spec
//					bg.Empty().Add(utils.QualSprintf(`
//						enc := %Q(encoding/json,NewEncoder)(envelope)
//						if err := enc.Encode(message); err != nil {
//							return err
//						}`))
//				} else { // Message is set for Channel in spec
//					bg.Op(`
//						if err := message.MarshalAMQPEnvelope(envelope); err != nil {
//							return err
//						}`)
//				}
//				bg.Id("envelope").Dot("SetRoutingKey").Call(j.Id(rn).Dot("routingKey"))
//				// Message SetBindings
//				if pc.PublisherMessageTypePromise != nil && pc.PublisherMessageTypePromise.Target().HasProtoBindings(pc.ProtoName) {
//					bg.Op("envelope.SetBindings").Call(
//						j.Add(utils.ToCode(pc.PublisherMessageTypePromise.Target().BindingsType.U(ctx))...).Values().Dot(pc.ProtoTitle).Call(),
//					)
//				}
//				bg.Return(j.Nil())
//			}),
//	}
//}
