package ws

//type ProtoServer struct {
//	Parent *render.Server
//	Struct *render.GoStruct
//
//	ProtoName, ProtoTitle string
//}
//
//func (ps ProtoServer) Selectable() bool {
//	return true
//}
//
//func (ps ProtoServer) D(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Server", "", ps.Parent.Name, "definition", ps.Selectable(), "proto", ps.ProtoName)
//	defer ctx.LogFinishRender()
//	var res []*j.Statement
//	res = append(res, ps.RenderNewFunc(ctx)...)
//	res = append(res, ps.Struct.D(ctx)...)
//	res = append(res, ps.RenderCommonMethods(ctx)...)
//	res = append(res, ps.renderChannelMethods(ctx)...)
//	res = append(res, ps.RenderProducerMethods(ctx)...)
//	res = append(res, ps.RenderConsumerMethods(ctx)...)
//	return res
//}
//
//func (ps ProtoServer) U(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("Server", "", ps.Parent.Name, "usage", ps.Selectable(), "proto", ps.ProtoName)
//	defer ctx.LogFinishRender()
//	return ps.Struct.U(ctx)
//}
//
//func (ps ProtoServer) ID() string {
//	return ps.Parent.Name
//}
//
//func (ps ProtoServer) String() string {
//	return "WS ProtoServer " + ps.Parent.Name
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
//			ps.RenderOpenChannelMethod(ctx, protoChan.Struct, protoChan, protoChan.Parent.ParametersStruct)...,
//		)
//	}
//	return res
//}
