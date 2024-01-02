package asyncapi

import (
	"encoding/json"
	"path"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	renderProto "github.com/bdragon300/asyncapi-codegen-go/internal/render/proto"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type ProtocolBuilder interface {
	BuildChannel(ctx *common.CompileContext, channel *Channel, channelKey string, abstractChannel *render.Channel) (common.Renderer, error)
	BuildServer(ctx *common.CompileContext, server *Server, serverKey string) (common.Renderer, error)

	BuildMessageBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildOperationBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildChannelBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildServerBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error)

	ProtocolName() string
	ProtocolAbbreviation() string
}

var ProtocolBuilders map[string]ProtocolBuilder // TODO: replace the global variableon smth better

type BaseProtoBuilder struct {
	ProtoName, ProtoAbbr string
}

func (pb BaseProtoBuilder) BuildBaseProtoChannel(
	ctx *common.CompileContext,
	channel *Channel,
	channelKey string,
	abstractChannel *render.Channel,
) (*renderProto.BaseProtoChannel, error) {
	chName, _ := lo.Coalesce(channel.XGoName, channelKey)
	chanResult := &renderProto.BaseProtoChannel{
		Name: chName,
		Struct: &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(chName, pb.ProtoAbbr),
				Description:  channel.Description,
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
			Fields: []render.StructField{
				{Name: "name", Type: &render.Simple{Name: "ParamString", Package: ctx.RuntimePackage("")}},
			},
		},
		AbstractChannel:     abstractChannel,
		FallbackMessageType: &render.Simple{Name: "any", IsIface: true},
		ProtoName:           pb.ProtoName,
		ProtoAbbr:           pb.ProtoAbbr,
	}

	// Interface to match servers bound with a channel (type chan1KafkaServer interface)
	var openChannelServerIfaceArgs []render.FuncParam
	if chanResult.AbstractChannel.ParametersStruct != nil {
		openChannelServerIfaceArgs = append(openChannelServerIfaceArgs, render.FuncParam{
			Name: "params",
			Type: &render.Simple{Name: chanResult.AbstractChannel.ParametersStruct.Name, Package: ctx.TopPackageName()},
		})
	}
	chanResult.ServerIface = &render.Interface{
		BaseType: render.BaseType{
			Name:         utils.ToLowerFirstLetter(chanResult.Struct.Name + "Server"),
			DirectRender: true,
			PackageName:  ctx.TopPackageName(),
		},
		Methods: []render.FuncSignature{
			{
				Name: "Open" + chanResult.Struct.Name,
				Args: openChannelServerIfaceArgs,
				Return: []render.FuncParam{
					{Type: &render.Simple{Name: chanResult.Struct.Name, Package: ctx.TopPackageName()}, Pointer: true},
					{Type: &render.Simple{Name: "error"}},
				},
			},
		},
	}

	// Publisher stuff
	if channel.Publish != nil && !channel.Publish.XIgnore {
		ctx.Logger.Trace("Channel publish operation", "proto", pb.ProtoName)
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, render.StructField{
			Name:        "publisher",
			Description: channel.Publish.Description,
			Type: &render.Simple{
				Name:    "Publisher",
				Package: ctx.RuntimePackage(pb.ProtoName),
				IsIface: true,
			},
		})
		chanResult.Publisher = true
		if channel.Publish.Message != nil {
			ctx.Logger.Trace("Channel publish operation message", "proto", pb.ProtoName)
			ref := path.Join(ctx.PathRef(), "publish/message")
			chanResult.PubMessagePromise = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(chanResult.PubMessagePromise)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, render.FuncSignature{
			Name: "Producer",
			Args: nil,
			Return: []render.FuncParam{
				{Type: &render.Simple{Name: "Producer", Package: ctx.RuntimePackage(pb.ProtoName), IsIface: true}},
			},
		})
	}

	// Subscriber stuff
	if channel.Subscribe != nil && !channel.Subscribe.XIgnore {
		ctx.Logger.Trace("Channel subscribe operation", "proto", pb.ProtoName)
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, render.StructField{
			Name:        "subscriber",
			Description: channel.Subscribe.Description,
			Type: &render.Simple{
				Name:    "Subscriber",
				Package: ctx.RuntimePackage(pb.ProtoName),
				IsIface: true,
			},
		})
		chanResult.Subscriber = true
		if channel.Subscribe.Message != nil {
			ctx.Logger.Trace("Channel subscribe operation message", "proto", pb.ProtoName)
			ref := path.Join(ctx.PathRef(), "subscribe/message")
			chanResult.SubMessagePromise = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(chanResult.SubMessagePromise)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, render.FuncSignature{
			Name: "Consumer",
			Args: nil,
			Return: []render.FuncParam{
				{Type: &render.Simple{Name: "Consumer", Package: ctx.RuntimePackage(pb.ProtoName), IsIface: true}},
			},
		})
	}

	return chanResult, nil
}

func (pb BaseProtoBuilder) BuildBaseProtoServer(
	ctx *common.CompileContext,
	server *Server,
	serverKey string,
) (*renderProto.BaseProtoServer, error) {
	srvName, _ := lo.Coalesce(server.XGoName, serverKey)
	srvResult := &renderProto.BaseProtoServer{
		Name:            srvName,
		URL:             server.URL,
		ProtocolVersion: server.ProtocolVersion,
		Struct: &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(srvName, ""),
				Description:  server.Description,
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
		},
		ProtoName: pb.ProtoName,
		ProtoAbbr: pb.ProtoAbbr,
	}

	// Server variables
	for _, v := range server.Variables.Entries() {
		ctx.Logger.Trace("Server variable", "name", v.Key, "proto", pb.ProtoName)
		srvResult.Variables.Set(v.Key, renderProto.ServerVariable{
			ArgName:     utils.ToGolangName(v.Key, false),
			Enum:        v.Value.Enum,
			Default:     v.Value.Default,
			Description: v.Value.Description,
		})
	}

	// Channels which are connected to this server
	channelsPrm := render.NewListCbPromise[*render.Channel](func(item common.Renderer, path []string) bool {
		ch, ok := item.(*render.Channel)
		if !ok {
			return false
		}
		if len(ch.ExplicitServerNames) > 0 {
			return lo.Contains(ch.ExplicitServerNames, serverKey)
		}
		return len(ch.ExplicitServerNames) == 0 // Empty servers list means "all servers", see spec
	})
	srvResult.ChannelsPromise = channelsPrm
	ctx.PutListPromise(channelsPrm)

	// Producer/consumer
	ctx.Logger.Trace("Server producer", "proto", pb.ProtoName)
	fld := render.StructField{
		Name: "producer",
		Type: &render.Simple{Name: "Producer", Package: ctx.RuntimePackage(pb.ProtoName), IsIface: true},
	}
	srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)

	ctx.Logger.Trace("Server consumer", "proto", pb.ProtoName)
	fld = render.StructField{
		Name: "consumer",
		Type: &render.Simple{Name: "Consumer", Package: ctx.RuntimePackage(pb.ProtoName), IsIface: true},
	}
	srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)

	return srvResult, nil
}

func (pb BaseProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}

func (pb BaseProtoBuilder) ProtocolAbbreviation() string {
	return pb.ProtoAbbr
}

