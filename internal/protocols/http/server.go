package http

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	j "github.com/dave/jennifer/jen"
)

func BuildServer(ctx *common.CompileContext, server *asyncapi.Server, serverKey string) (common.Renderer, error) {
	baseServer, err := protocols.BuildServer(ctx, server, serverKey, ProtoName)
	if err != nil {
		return nil, err
	}
	srvResult := ProtoServer{BaseProtoServer: *baseServer}

	return srvResult, nil
}

type ProtoServer struct {
	protocols.BaseProtoServer
}

func (p ProtoServer) DirectRendering() bool {
	return true
}

func (p ProtoServer) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	if p.ProtocolVersion != "" {
		res = append(res, protocols.RenderServerProtocolVersionConst(p.Struct, p.ProtocolVersion)...)
	}
	res = append(res, protocols.RenderServerURLFunc(ctx, p.Struct, p.Variables, p.URL)...)
	res = append(res, protocols.RenderServerNewFunc(ctx, p.Struct, p.Producer, p.Consumer, ProtoName)...)
	res = append(res, p.Struct.RenderDefinition(ctx)...)
	res = append(res, protocols.RenderServerCommonMethods(ctx, p.Struct, p.Name)...)
	res = append(res, p.renderChannelMethods(ctx)...)
	if p.Producer {
		res = append(res, protocols.RenderServerProducerMethods(ctx, p.Struct, ProtoName)...)
	}
	if p.Consumer {
		res = append(res, protocols.RenderServerConsumerMethods(ctx, p.Struct, ProtoName)...)
	}
	return res
}

func (p ProtoServer) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	return p.Struct.RenderUsage(ctx)
}

func (p ProtoServer) String() string {
	return p.BaseProtoServer.Name
}

func (p ProtoServer) renderChannelMethods(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement

	for _, ch := range p.ChannelsPromise.Targets() {
		protoChan := ch.AllProtocols[ProtoName].(*ProtoChannel)
		res = append(res,
			protocols.RenderServerChannelMethod(ctx, p.Struct, protoChan.Struct, protoChan, protoChan.ParametersStructNoRender)...,
		)
	}
	return res
}
