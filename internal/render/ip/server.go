package ip

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/proto"
	j "github.com/dave/jennifer/jen"
)

type ProtoServer struct {
	proto.BaseProtoServer
}

func (ps ProtoServer) DirectRendering() bool {
	return true
}

func (ps ProtoServer) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	ctx.LogStartRender("Server", "", ps.Parent.Name, "definition", ps.DirectRendering(), "proto", ps.ProtoName)
	defer ctx.LogFinishRender()
	var res []*j.Statement
	res = append(res, ps.RenderNewFunc(ctx)...)
	res = append(res, ps.Struct.RenderDefinition(ctx)...)
	res = append(res, ps.RenderCommonMethods(ctx)...)
	res = append(res, ps.renderChannelMethods(ctx)...)
	res = append(res, ps.RenderProducerMethods(ctx)...)
	res = append(res, ps.RenderConsumerMethods(ctx)...)
	return res
}

func (ps ProtoServer) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	ctx.LogStartRender("Server", "", ps.Parent.Name, "usage", ps.DirectRendering(), "proto", ps.ProtoName)
	defer ctx.LogFinishRender()
	return ps.Struct.RenderUsage(ctx)
}

func (ps ProtoServer) ID() string {
	return ps.Parent.Name
}

func (ps ProtoServer) String() string {
	return "IP ProtoServer " + ps.Parent.Name
}

func (ps ProtoServer) renderChannelMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderChannelMethods", "proto", ps.ProtoName)

	var res []*j.Statement

	for _, ch := range ps.Parent.GetRelevantChannels() {
		protoChan := ch.AllProtoChannels[ps.ProtoName].(*ProtoChannel)
		res = append(res,
			ps.RenderOpenChannelMethod(ctx, protoChan.Struct, protoChan, protoChan.Parent.ParametersStruct)...,
		)
	}
	return res
}
