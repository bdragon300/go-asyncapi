package proto

//type BaseProtoServer struct {
//	Parent *render.Server
//	Type *render.GoStruct
//
//	Protocol, ProtoTitle string
//}

//func (ps BaseProtoServer) RenderNewFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderNewFunc", "proto", ps.Protocol)
//
//	return []*j.Statement{
//		// NewServer1(producer proto.Producer, consumer proto.Consumer) *Server1
//		j.Func().Id(ps.Type.NewFuncName()).
//			ParamsFunc(func(g *j.Group) {
//				g.Id("producer").Qual(ctx.RuntimeModule(ps.Protocol), "Producer")
//				g.Id("consumer").Qual(ctx.RuntimeModule(ps.Protocol), "Consumer")
//			}).
//			Op("*").Add(utils.ToCode(ps.Type.U(ctx))...).
//			Block(
//				j.Return(j.Op("&").Add(utils.ToCode(ps.Type.U(ctx))...).Values(j.DictFunc(func(d j.Dict) {
//					d[j.Id("producer")] = j.Id("producer")
//					d[j.Id("consumer")] = j.Id("consumer")
//				}))),
//			),
//	}
//}

//func (ps BaseProtoServer) RenderCommonMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderCommonMethods", "proto", ps.Protocol)
//
//	receiver := j.Id(ps.Type.ReceiverName()).Id(ps.Type.GetOriginalName)
//
//	return []*j.Statement{
//		// Method GetOriginalName() string
//		j.Func().Params(receiver.Clone()).Id("GetOriginalName").
//			Params().
//			String().
//			Block(
//				j.Return(j.Lit(ps.Parent.GetOriginalName)),
//			),
//	}
//}

//func (ps BaseProtoServer) RenderOpenChannelMethod(ctx *common.RenderContext, channelStruct *render.GoStruct, channel common.Renderer, channelParametersStructNoRender *render.GoStruct) []*j.Statement {
//	ctx.Logger.Trace("RenderOpenChannelMethod", "proto", ps.Protocol)
//
//	rn := ps.Type.ReceiverName()
//	receiver := j.Id(rn).Id(ps.Type.GetOriginalName)
//
//	return []*j.Statement{
//		// Method OpenChannel1Proto(ctx context.Context, params Channel1Parameters) (*Channel1Proto, error)
//		j.Func().Params(receiver.Clone()).Id("Open"+channelStruct.GetOriginalName).
//			ParamsFunc(func(g *j.Group) {
//				g.Id("ctx").Qual("context", "Context")
//				if channelParametersStructNoRender != nil {
//					g.Id("params").Add(utils.ToCode(channelParametersStructNoRender.U(ctx))...)
//				}
//			}).
//			Params(j.Op("*").Add(utils.ToCode(channel.U(ctx))...), j.Error()).
//			Block(
//				j.Return(j.Qual(ctx.GeneratedModule(channelStruct.Import), "Open"+channelStruct.GetOriginalName).CallFunc(func(g *j.Group) {
//					g.Id("ctx")
//					if channelParametersStructNoRender != nil {
//						g.Id("params")
//					}
//					g.Id(rn)
//				})),
//			),
//	}
//}

//func (ps BaseProtoServer) RenderProducerMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderProducerMethods", "proto", ps.Protocol)
//
//	rn := ps.Type.ReceiverName()
//	receiver := j.Id(rn).Id(ps.Type.GetOriginalName)
//
//	return []*j.Statement{
//		// Method Producer() proto.Producer
//		j.Func().Params(receiver.Clone()).Id("Producer").
//			Params().
//			Qual(ctx.RuntimeModule(ps.Protocol), "Producer").
//			Block(
//				j.Return(j.Id(rn).Dot("producer")),
//			),
//	}
//}

//func (ps BaseProtoServer) RenderConsumerMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderConsumerMethods", "proto", ps.Protocol)
//
//	rn := ps.Type.ReceiverName()
//	receiver := j.Id(rn).Id(ps.Type.GetOriginalName)
//
//	return []*j.Statement{
//		// Method Consumer() proto.Consumer
//		j.Func().Params(receiver.Clone()).Id("Consumer").
//			Params().
//			Qual(ctx.RuntimeModule(ps.Protocol), "Consumer").
//			Block(
//				j.Return(j.Id(rn).Dot("consumer")),
//			),
//	}
//}
