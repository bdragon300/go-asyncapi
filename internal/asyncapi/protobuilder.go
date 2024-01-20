package asyncapi

import (
	"encoding/json"
	"path"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	renderProto "github.com/bdragon300/go-asyncapi/internal/render/proto"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type ProtocolBuilder interface {
	BuildChannel(ctx *common.CompileContext, channel *Channel, channelKey string, abstractChannel *render.Channel) (common.Renderer, error)
	BuildServer(ctx *common.CompileContext, server *Server, serverKey string, abstractServer *render.Server) (common.Renderer, error)

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
	channelKey string,
	abstractChannel *render.Channel,
) (*renderProto.BaseProtoChannel, error) {
	chName, _ := lo.Coalesce(channel.XGoName, channelKey)
	chanResult := &renderProto.BaseProtoChannel{
		Name: chName,
		Struct: &render.GoStruct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(chName, pb.ProtoTitle),
				Description:  channel.Description,
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
			Fields: []render.GoStructField{
				{Name: "name", Type: &render.GoSimple{Name: "ParamString", Import: ctx.RuntimeModule("")}},
			},
		},
		AbstractChannel:     abstractChannel,
		FallbackMessageType: &render.GoSimple{Name: "any", IsIface: true},
		ProtoName:           pb.ProtoName,
		ProtoTitle:          pb.ProtoTitle,
	}

	// Interface to match servers bound with a channel (type chan1KafkaServer interface)
	var openChannelServerIfaceArgs []render.GoFuncParam
	if chanResult.AbstractChannel.ParametersStruct != nil {
		openChannelServerIfaceArgs = append(openChannelServerIfaceArgs, render.GoFuncParam{
			Name: "params",
			Type: &render.GoSimple{Name: chanResult.AbstractChannel.ParametersStruct.Name, Import: ctx.CurrentPackage()},
		})
	}
	chanResult.ServerIface = &render.GoInterface{
		BaseType: render.BaseType{
			Name:         utils.ToLowerFirstLetter(chanResult.Struct.Name + "Server"),
			DirectRender: true,
			Import:       ctx.CurrentPackage(),
		},
		Methods: []render.GoFuncSignature{
			{
				Name: "Open" + chanResult.Struct.Name,
				Args: openChannelServerIfaceArgs,
				Return: []render.GoFuncParam{
					{Type: &render.GoSimple{Name: chanResult.Struct.Name, Import: ctx.CurrentPackage()}, Pointer: true},
					{Type: &render.GoSimple{Name: "error", IsIface: true}},
				},
			},
		},
	}

	// Publisher stuff
	if channel.Publish != nil && !channel.Publish.XIgnore && ctx.CompileOpts.GeneratePublishers {
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
		chanResult.Publisher = true
		if channel.Publish.Message != nil {
			ctx.Logger.Trace("Channel publish operation message", "proto", pb.ProtoName)
			ref := path.Join(ctx.PathRef(), "publish/message")
			chanResult.PubMessagePromise = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(chanResult.PubMessagePromise)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, render.GoFuncSignature{
			Name: "Producer",
			Args: nil,
			Return: []render.GoFuncParam{
				{Type: &render.GoSimple{Name: "Producer", Import: ctx.RuntimeModule(pb.ProtoName), IsIface: true}},
			},
		})
	}

	// Subscriber stuff
	if channel.Subscribe != nil && !channel.Subscribe.XIgnore && ctx.CompileOpts.GenerateSubscribers {
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
		chanResult.Subscriber = true
		if channel.Subscribe.Message != nil {
			ctx.Logger.Trace("Channel subscribe operation message", "proto", pb.ProtoName)
			ref := path.Join(ctx.PathRef(), "subscribe/message")
			chanResult.SubMessagePromise = render.NewPromise[*render.Message](ref, common.PromiseOriginInternal)
			ctx.PutPromise(chanResult.SubMessagePromise)
		}
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
	serverKey string,
	abstractServer *render.Server,
) (*renderProto.BaseProtoServer, error) {
	srvName, _ := lo.Coalesce(server.XGoName, serverKey)
	srvResult := &renderProto.BaseProtoServer{
		Name:            srvName,
		URL:             server.URL,
		ProtocolVersion: server.ProtocolVersion,
		Struct: &render.GoStruct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(srvName, ""),
				Description:  server.Description,
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
		},
		AbstractServer: abstractServer,
		ProtoName:      pb.ProtoName,
		ProtoTitle:     pb.ProtoTitle,
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

