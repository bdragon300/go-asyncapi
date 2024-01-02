package amqp

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render/proto"
	j "github.com/dave/jennifer/jen"
)

type ProtoServer struct {
	proto.BaseProtoServer
}

func (ps ProtoServer) DirectRendering() bool {
	return true
}

func (ps ProtoServer) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	if ps.ProtocolVersion != "" {
		res = append(res, ps.RenderProtocolVersionConst()...)
	}
	res = append(res, ps.RenderURLFunc(ctx)...)
	res = append(res, ps.RenderNewFunc(ctx)...)
	res = append(res, ps.Struct.RenderDefinition(ctx)...)
	res = append(res, ps.RenderCommonMethods(ctx)...)
	res = append(res, ps.renderChannelMethods(ctx)...)
	res = append(res, ps.RenderProducerMethods(ctx)...)
	res = append(res, ps.RenderConsumerMethods(ctx)...)
	return res
}

func (ps ProtoServer) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	return ps.Struct.RenderUsage(ctx)
}

func (ps ProtoServer) String() string {
	return ps.BaseProtoServer.Name
}

func (ps ProtoServer) renderChannelMethods(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement

	for _, ch := range ps.ChannelsPromise.Targets() {
		protoChan := ch.AllProtoChannels[ps.ProtoName].(*ProtoChannel)
		res = append(res,
			ps.RenderOpenChannelMethod(ctx, protoChan.Struct, protoChan, protoChan.AbstractChannel.ParametersStruct)...,
		)
	}
	return res
}