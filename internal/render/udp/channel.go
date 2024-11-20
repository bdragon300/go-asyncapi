package udp

//type ProtoChannel struct {
//	*render.Channel
//	GolangNameProto string // Channel GolangName name concatenated with protocol name, e.g. Channel1Kafka
//	Struct          *render.GoStruct
//
//	ProtoName, ProtoTitle string
//}
//
//func (pc ProtoChannel) DirectRendering() bool {
//	return true
//}
//
//func (pc ProtoChannel) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Channel", "", pc.Parent.Name, "definition", pc.DirectRendering(), "proto", pc.ProtoName)
//	defer ctx.LogFinishRender()
//	var res []*j.Statement
//	res = append(res, pc.ServerIface.RenderDefinition(ctx)...)
//	res = append(res, pc.RenderOpenFunc(ctx)...)
//	res = append(res, pc.renderNewFunc(ctx)...)
//	res = append(res, pc.Struct.RenderDefinition(ctx)...)
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
//
//func (pc ProtoChannel) RenderUsage(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Channel", "", pc.Parent.Name, "usage", pc.DirectRendering(), "proto", pc.ProtoName)
//	defer ctx.LogFinishRender()
//	return pc.Struct.RenderUsage(ctx)
//}
//
//func (pc ProtoChannel) ID() string {
//	return pc.Parent.Name
//}
//
//func (pc ProtoChannel) String() string {
//	return "UDP ProtoChannel " + pc.Parent.Name
//}
//
//func (pc ProtoChannel) renderNewFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderNewFunc", "proto", pc.ProtoName)
//
//	return []*j.Statement{
//		// NewChannel1Proto(params Channel1Parameters, publisher proto.Publisher, subscriber proto.Subscriber) *Channel1Proto
//		j.Func().Id(pc.Struct.NewFuncName()).
//			ParamsFunc(func(g *j.Group) {
//				if pc.Parent.ParametersStruct != nil {
//					g.Id("params").Add(utils.ToCode(pc.Parent.ParametersStruct.RenderUsage(ctx))...)
//				}
//				if pc.Parent.Publisher {
//					g.Id("publisher").Qual(ctx.RuntimeModule(pc.ProtoName), "Publisher")
//				}
//				if pc.Parent.Subscriber {
//					g.Id("subscriber").Qual(ctx.RuntimeModule(pc.ProtoName), "Subscriber")
//				}
//			}).
//			Op("*").Add(utils.ToCode(pc.Struct.RenderUsage(ctx))...).
//			BlockFunc(func(bg *j.Group) {
//				bg.Op("res := ").Add(utils.ToCode(pc.Struct.RenderUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
//					d[j.Id("name")] = j.Id(pc.Parent.GolangName + "Name").CallFunc(func(g *j.Group) {
//						if pc.Parent.ParametersStruct != nil {
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
//				bg.Op(`return &res`)
//			}),
//	}
//}
//
//func (pc ProtoChannel) renderProtoMethods(_ *common.RenderContext) []*j.Statement {
//	return []*j.Statement{}
//}
//
//func (pc ProtoChannel) renderProtoPublisherMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderProtoPublisherMethods", "proto", pc.ProtoName)
//
//	rn := pc.Struct.ReceiverName()
//	receiver := j.Id(rn).Id(pc.Struct.Name)
//
//	var msgTyp common.GolangType = render.GoPointer{Type: pc.Parent.FallbackMessageType, DirectRender: true}
//	if pc.Parent.PubMessagePromise != nil {
//		msgTyp = render.GoPointer{Type: pc.Parent.PubMessagePromise.Target().OutStruct, DirectRender: true}
//	}
//
//	return []*j.Statement{
//		// Method SealEnvelope(envelope proto.EnvelopeWriter, message *Message1Out) error
//		j.Func().Params(receiver.Clone()).Id("SealEnvelope").
//			ParamsFunc(func(g *j.Group) {
//				g.Id("envelope").Qual(ctx.RuntimeModule(pc.ProtoName), "EnvelopeWriter")
//				g.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...)
//			}).
//			Error().
//			BlockFunc(func(bg *j.Group) {
//				bg.Op("envelope.ResetPayload()")
//				if pc.Parent.PubMessagePromise == nil { // No Message set for Channel in spec
//					bg.Empty().Add(utils.QualSprintf(`
//						enc := %Q(encoding/json,NewEncoder)(envelope)
//						if err := enc.Encode(message); err != nil {
//							return err
//						}`))
//				} else { // Message is set for Channel in spec
//					bg.Op(`
//						if err := message.MarshalUDPEnvelope(envelope); err != nil {
//							return err
//						}`)
//				}
//				// Message SetBindings
//				if pc.Parent.PubMessagePromise != nil && pc.Parent.PubMessagePromise.Target().HasProtoBindings(pc.ProtoName) {
//					bg.Op("envelope.SetBindings").Call(
//						j.Add(utils.ToCode(pc.Parent.PubMessagePromise.Target().BindingsStruct.RenderUsage(ctx))...).Values().Dot(pc.ProtoTitle).Call(),
//					)
//				}
//				bg.Return(j.Nil())
//			}),
//	}
//}
