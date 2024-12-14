package ip

//type ProtoServer struct {
//	Parent *render.Server
//	Type *render.GoStruct
//
//	ProtoName, ProtoTitle string
//}
//
//func (ps ProtoServer) Selectable() bool {
//	return true
//}
//
//func (ps ProtoServer) D(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Server", "", ps.Parent.GetOriginalName, "definition", ps.Selectable(), "proto", ps.ProtoName)
//	defer ctx.LogFinishRender()
//	var res []*j.Statement
//	res = append(res, ps.RenderNewFunc(ctx)...)
//	res = append(res, ps.Type.D(ctx)...)
//	res = append(res, ps.RenderCommonMethods(ctx)...)
//	res = append(res, ps.renderChannelMethods(ctx)...)
//	res = append(res, ps.RenderProducerMethods(ctx)...)
//	res = append(res, ps.RenderConsumerMethods(ctx)...)
//	return res
//}

//func (ps ProtoServer) U(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Server", "", ps.Parent.GetOriginalName, "usage", ps.Selectable(), "proto", ps.ProtoName)
//	defer ctx.LogFinishRender()
//	return ps.Type.U(ctx)
//}
//
//func (ps ProtoServer) ID() string {
//	return ps.Parent.GetOriginalName
//}
//
//func (ps ProtoServer) String() string {
//	return "IP ProtoServer " + ps.Parent.GetOriginalName
//}
//
//func (ps ProtoServer) renderChannelMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderChannelMethods", "proto", ps.ProtoName)
//
//	var res []*j.Statement
//
//	for _, ch := range ps.Parent.GetRelevantChannels() {
//		protoChan := ch.AllProtoChannels[ps.ProtoName].(*ProtoChannel)
//		res = append(res,
//			ps.RenderOpenChannelMethod(ctx, protoChan.Type, protoChan, protoChan.Parent.ParametersType)...,
//		)
//	}
//	return res
//}
