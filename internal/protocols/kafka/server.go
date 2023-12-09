package kafka

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/compile"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type serverBindings struct {
	SchemaRegistryURL    *string `json:"schemaRegistryUrl" yaml:"schemaRegistryUrl"`
	SchemaRegistryVendor *string `json:"schemaRegistryVendor" yaml:"schemaRegistryVendor"`
}

func BuildServer(ctx *common.CompileContext, server *compile.Server, serverKey string) (common.Renderer, error) {
	baseServer, err := protocols.BuildServer(ctx, server, serverKey, ProtoName)
	if err != nil {
		return nil, err
	}
	srvResult := ProtoServer{BaseProtoServer: *baseServer}

	// Server bindings
	if server.Bindings.Len() > 0 {
		ctx.Logger.Trace("Server bindings", "proto", ProtoName)
		if b, ok := server.Bindings.Get(ProtoName); ok {
			vals := &render.StructInit{
				Type: &render.Simple{Name: "ServerBindings", Package: ctx.RuntimePackage(ProtoName)},
			}
			bindingsStruct := &render.Struct{ // TODO: remove in favor of parent server
				BaseType: render.BaseType{
					Name:         ctx.GenerateObjName(serverKey, "Bindings"),
					DirectRender: true,
					PackageName:  ctx.TopPackageName(),
				},
			}

			var bindings serverBindings
			if err := utils.UnmarshalRawsUnion2(b, &bindings); err != nil { // TODO: implement $ref
				return nil, common.CompileError{Err: err, Path: ctx.PathRef()}
			}
			marshalFields := []string{"SchemaRegistryURL", "SchemaRegistryVendor"}
			if err := utils.StructToOrderedMap(bindings, &vals.Values, marshalFields); err != nil {
				return nil, common.CompileError{Err: err, Path: ctx.PathRef()}
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

	for _, ch := range p.ChannelLinkList.Targets() {
		protoChan := ch.AllProtocols[ProtoName].(*ProtoChannel)
		res = append(res,
			protocols.RenderServerChannelMethod(ctx, p.Struct, protoChan.Struct, protoChan, protoChan.ParametersStructNoRender)...,
		)
	}
	return res
}
