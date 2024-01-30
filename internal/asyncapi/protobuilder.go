package asyncapi

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	renderProto "github.com/bdragon300/go-asyncapi/internal/render/proto"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"gopkg.in/yaml.v3"
)

type ProtocolBuilder interface {
	BuildChannel(ctx *common.CompileContext, channel *Channel, parent *render.Channel) (common.Renderer, error)
	BuildServer(ctx *common.CompileContext, server *Server, parent *render.Server) (common.Renderer, error)

	BuildMessageBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildOperationBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildChannelBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildServerBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)

	ProtocolName() string
	ProtocolTitle() string
}

var ProtocolBuilders map[string]ProtocolBuilder // TODO: replace the global variableon smth better

type BaseProtoBuilder struct {
	ProtoName, ProtoTitle string
}

func (pb BaseProtoBuilder) BuildBaseProtoChannel(
	ctx *common.CompileContext,
	channel *Channel,
	parentChannel *render.Channel,
) (*renderProto.BaseProtoChannel, error) {
	golangName := parentChannel.GolangName + pb.ProtoTitle
	chanResult := &renderProto.BaseProtoChannel{
		Parent:          parentChannel,
		GolangNameProto: golangName,
		Struct: &render.GoStruct{
			BaseType: render.BaseType{
				Name:         golangName,
				Description:  channel.Description,
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
			Fields: []render.GoStructField{
				{Name: "name", Type: &render.GoSimple{Name: "ParamString", Import: ctx.RuntimeModule("")}},
			},
		},
		ProtoName:  pb.ProtoName,
		ProtoTitle: pb.ProtoTitle,
	}

	// Interface to match servers bound with a channel (type chan1KafkaServer interface)
	openChannelServerIfaceArgs := []render.GoFuncParam{{
		Name: "ctx",
		Type: &render.GoSimple{Name: "Context", Import: "context", IsIface: true},
	}}
	if chanResult.Parent.ParametersStruct != nil {
		openChannelServerIfaceArgs = append(openChannelServerIfaceArgs, render.GoFuncParam{
			Name: "params",
			Type: &render.GoSimple{Name: chanResult.Parent.ParametersStruct.Name, Import: ctx.CurrentPackage()},
		})
	}
	chanResult.ServerIface = &render.GoInterface{
		BaseType: render.BaseType{
			Name:         utils.ToLowerFirstLetter(chanResult.GolangNameProto + "Server"),
			DirectRender: true,
			Import:       ctx.CurrentPackage(),
		},
		Methods: []render.GoFuncSignature{
			{
				Name: "Open" + chanResult.GolangNameProto,
				Args: openChannelServerIfaceArgs,
				Return: []render.GoFuncParam{
					{Type: &render.GoSimple{Name: chanResult.Struct.Name, Import: ctx.CurrentPackage()}, Pointer: true},
					{Type: &render.GoSimple{Name: "error", IsIface: true}},
				},
			},
		},
	}

	// Publisher stuff
	if parentChannel.Publisher {
		ctx.Logger.Trace("Channel publish operation", "proto", pb.ProtoName)
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, render.GoStructField{
			Name:        "publisher",
			Description: channel.Publish.Description,
			Type: &render.GoSimple{
				Name:    "Publisher",
				Import:  ctx.RuntimeModule(pb.ProtoName),
				IsIface: true,
			},
		})

		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, render.GoFuncSignature{
			Name: "Producer",
			Args: nil,
			Return: []render.GoFuncParam{
				{Type: &render.GoSimple{Name: "Producer", Import: ctx.RuntimeModule(pb.ProtoName), IsIface: true}},
			},
		})
	}

	// Subscriber stuff
	if parentChannel.Subscriber {
		ctx.Logger.Trace("Channel subscribe operation", "proto", pb.ProtoName)
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, render.GoStructField{
			Name:        "subscriber",
			Description: channel.Subscribe.Description,
			Type: &render.GoSimple{
				Name:    "Subscriber",
				Import:  ctx.RuntimeModule(pb.ProtoName),
				IsIface: true,
			},
		})

		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, render.GoFuncSignature{
			Name: "Consumer",
			Args: nil,
			Return: []render.GoFuncParam{
				{Type: &render.GoSimple{Name: "Consumer", Import: ctx.RuntimeModule(pb.ProtoName), IsIface: true}},
			},
		})
	}

	return chanResult, nil
}

func (pb BaseProtoBuilder) BuildBaseProtoServer(
	ctx *common.CompileContext,
	server *Server,
	abstractServer *render.Server,
) (*renderProto.BaseProtoServer, error) {
	srvResult := &renderProto.BaseProtoServer{
		Struct: &render.GoStruct{
			BaseType: render.BaseType{
				Name:         abstractServer.GolangName,
				Description:  server.Description,
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
		},
		Parent:     abstractServer,
		ProtoName:  pb.ProtoName,
		ProtoTitle: pb.ProtoTitle,
	}

	// Producer/consumer
	ctx.Logger.Trace("Server producer", "proto", pb.ProtoName)
	fld := render.GoStructField{
		Name: "producer",
		Type: &render.GoSimple{Name: "Producer", Import: ctx.RuntimeModule(pb.ProtoName), IsIface: true},
	}
	srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)

	ctx.Logger.Trace("Server consumer", "proto", pb.ProtoName)
	fld = render.GoStructField{
		Name: "consumer",
		Type: &render.GoSimple{Name: "Consumer", Import: ctx.RuntimeModule(pb.ProtoName), IsIface: true},
	}
	srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)

	return srvResult, nil
}

func (pb BaseProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}

func (pb BaseProtoBuilder) ProtocolTitle() string {
	return pb.ProtoTitle
}

