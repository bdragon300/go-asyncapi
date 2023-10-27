package amqp

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/protocols"
	j "github.com/dave/jennifer/jen"
)

func BuildServer(ctx *common.CompileContext, server *compile.Server, serverKey string) (common.Assembler, error) {
	baseServer, err := protocols.BuildServer(ctx, server, serverKey, protoName)
	if err != nil {
		return nil, err
	}
	srvResult := ProtoServer{BaseProtoServer: *baseServer}

	// Server bindings (protocol has no server bindings)
	if server.Bindings.Len() > 0 {
		if _, ok := server.Bindings.Get(protoName); ok {
			vals := &assemble.StructInit{
				Type: &assemble.Simple{Type: "ServerBindings", Package: ctx.RuntimePackage(protoName)},
			}
			bindingsStruct := &assemble.Struct{
				BaseType: assemble.BaseType{
					Name:    srvResult.Struct.Name + "Bindings",
					Render:  true,
					Package: ctx.TopPackageName(),
				},
			}

			srvResult.BindingsMethod = &assemble.Func{
				FuncSignature: assemble.FuncSignature{
					Name: protoAbbr,
					Args: nil,
					Return: []assemble.FuncParam{
						{Type: assemble.Simple{Type: "ServerBindings", Package: ctx.RuntimePackage(protoName)}},
					},
				},
				Receiver:      bindingsStruct,
				Package:       ctx.TopPackageName(),
				BodyAssembler: protocols.ServerBindingsMethodBody(vals, nil),
			}
		}
	}

	return srvResult, nil
}

type ProtoServer struct {
	protocols.BaseProtoServer
	BindingsMethod *assemble.Func // nil if no bindings set in spec
}

func (p ProtoServer) AllowRender() bool {
	return true
}

func (p ProtoServer) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	if p.BindingsMethod != nil {
		res = append(res, p.BindingsMethod.AssembleDefinition(ctx)...)
	}
	if p.ProtocolVersion != "" {
		res = append(res, protocols.AssembleServerProtocolVersionConst(p.Struct, p.ProtocolVersion)...)
	}
	res = append(res, protocols.AssembleServerURLFunc(ctx, p.Struct, p.Variables, p.URL)...)
	res = append(res, protocols.AssembleServerNewFunc(ctx, p.Struct, p.Producer, p.Consumer, protoName)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, protocols.AssembleServerCommonMethods(ctx, p.Struct, p.Name, protoAbbr)...)
	res = append(res, p.assembleChannelMethods(ctx)...)
	if p.Producer {
		res = append(res, protocols.AssembleServerProducerMethods(ctx, p.Struct, protoName)...)
	}
	if p.Consumer {
		res = append(res, protocols.AssembleServerConsumerMethods(ctx, p.Struct, protoName)...)
	}
	return res
}

func (p ProtoServer) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoServer) assembleChannelMethods(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement

	for _, ch := range p.ChannelLinkList.Targets() {
		protoChan := ch.AllProtocols[protoName].(*ProtoChannel)
		res = append(res,
			protocols.AssembleServerChannelMethod(ctx, p.Struct, protoChan.Struct, protoChan, protoChan.ParametersStructNoAssemble)...,
		)
	}
	return res
}
