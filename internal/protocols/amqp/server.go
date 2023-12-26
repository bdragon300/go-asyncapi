package amqp

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	j "github.com/dave/jennifer/jen"
)

func BuildServer(ctx *common.CompileContext, server *asyncapi.Server, serverKey string) (common.Renderer, error) {
	baseServer, err := protocols.BuildServer(ctx, server, serverKey, ProtoName)
	if err != nil {
		return nil, err
	}
	srvResult := ProtoServer{BaseProtoServer: *baseServer}

	// Server bindings (protocol has no server bindings)
	if server.Bindings.Len() > 0 {
		if _, ok := server.Bindings.Get(ProtoName); ok {
			ctx.Logger.Trace("Server bindings", "proto", ProtoName)
			vals := &render.StructInit{
				Type: &render.Simple{Name: "ServerBindings", Package: ctx.RuntimePackage(ProtoName)},
			}
			bindingsStruct := &render.Struct{
				BaseType: render.BaseType{
					Name:         srvResult.Struct.Name + "Bindings",
					DirectRender: true,
					PackageName:  ctx.TopPackageName(),
				},
			}

			srvResult.BindingsMethod = &render.Func{
				FuncSignature: render.FuncSignature{
					Name: protoAbbr,
					Args: nil,
					Return: []render.FuncParam{
						{Type: render.Simple{Name: "ServerBindings", Package: ctx.RuntimePackage(ProtoName)}},
					},
				},
				Receiver:     bindingsStruct,
				PackageName:  ctx.TopPackageName(),
				BodyRenderer: protocols.ServerBindingsMethodBody(vals, nil),
			}
		}
	}

	return srvResult, nil
}

type ProtoServer struct {
	protocols.BaseProtoServer
	BindingsMethod *render.Func // nil if no bindings set in spec
}

func (p ProtoServer) DirectRendering() bool {
	return true
}

func (p ProtoServer) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	if p.BindingsMethod != nil {
		res = append(res, p.BindingsMethod.RenderDefinition(ctx)...)
	}
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
