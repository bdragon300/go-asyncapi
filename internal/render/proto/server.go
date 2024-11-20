package proto

//type BaseProtoServer struct {
//	Parent *render.Server
//	Struct *render.GoStruct
//
//	ProtoName, ProtoTitle string
//}

//func (ps BaseProtoServer) RenderNewFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderNewFunc", "proto", ps.ProtoName)
//
//	return []*j.Statement{
//		// NewServer1(producer proto.Producer, consumer proto.Consumer) *Server1
//		j.Func().Id(ps.Struct.NewFuncName()).
//			ParamsFunc(func(g *j.Group) {
//				g.Id("producer").Qual(ctx.RuntimeModule(ps.ProtoName), "Producer")
//				g.Id("consumer").Qual(ctx.RuntimeModule(ps.ProtoName), "Consumer")
//			}).
//			Op("*").Add(utils.ToCode(ps.Struct.RenderUsage(ctx))...).
//			Block(
//				j.Return(j.Op("&").Add(utils.ToCode(ps.Struct.RenderUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
//					d[j.Id("producer")] = j.Id("producer")
//					d[j.Id("consumer")] = j.Id("consumer")
//				}))),
//			),
//	}
//}

//func (ps BaseProtoServer) RenderCommonMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderCommonMethods", "proto", ps.ProtoName)
//
//	receiver := j.Id(ps.Struct.ReceiverName()).Id(ps.Struct.Name)
//
//	return []*j.Statement{
//		// Method Name() string
//		j.Func().Params(receiver.Clone()).Id("Name").
//			Params().
//			String().
//			Block(
//				j.Return(j.Lit(ps.Parent.Name)),
//			),
//	}
//}

//func (ps BaseProtoServer) RenderOpenChannelMethod(ctx *common.RenderContext, channelStruct *render.GoStruct, channel common.Renderer, channelParametersStructNoRender *render.GoStruct) []*j.Statement {
//	ctx.Logger.Trace("RenderOpenChannelMethod", "proto", ps.ProtoName)
//
//	rn := ps.Struct.ReceiverName()
//	receiver := j.Id(rn).Id(ps.Struct.Name)
//
//	return []*j.Statement{
//		// Method OpenChannel1Proto(ctx context.Context, params Channel1Parameters) (*Channel1Proto, error)
//		j.Func().Params(receiver.Clone()).Id("Open"+channelStruct.Name).
//			ParamsFunc(func(g *j.Group) {
//				g.Id("ctx").Qual("context", "Context")
//				if channelParametersStructNoRender != nil {
//					g.Id("params").Add(utils.ToCode(channelParametersStructNoRender.RenderUsage(ctx))...)
//				}
//			}).
//			Params(j.Op("*").Add(utils.ToCode(channel.RenderUsage(ctx))...), j.Error()).
//			Block(
//				j.Return(j.Qual(ctx.GeneratedModule(channelStruct.Import), "Open"+channelStruct.Name).CallFunc(func(g *j.Group) {
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
//	ctx.Logger.Trace("RenderProducerMethods", "proto", ps.ProtoName)
//
//	rn := ps.Struct.ReceiverName()
//	receiver := j.Id(rn).Id(ps.Struct.Name)
//
//	return []*j.Statement{
//		// Method Producer() proto.Producer
//		j.Func().Params(receiver.Clone()).Id("Producer").
//			Params().
//			Qual(ctx.RuntimeModule(ps.ProtoName), "Producer").
//			Block(
//				j.Return(j.Id(rn).Dot("producer")),
//			),
//	}
//}

//func (ps BaseProtoServer) RenderConsumerMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderConsumerMethods", "proto", ps.ProtoName)
//
//	rn := ps.Struct.ReceiverName()
//	receiver := j.Id(rn).Id(ps.Struct.Name)
//
//	return []*j.Statement{
//		// Method Consumer() proto.Consumer
//		j.Func().Params(receiver.Clone()).Id("Consumer").
//			Params().
//			Qual(ctx.RuntimeModule(ps.ProtoName), "Consumer").
//			Block(
//				j.Return(j.Id(rn).Dot("consumer")),
//			),
//	}
//}
