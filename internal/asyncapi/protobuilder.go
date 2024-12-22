package asyncapi

import (
	"encoding/json"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type ProtocolBuilder interface {
	BuildChannel(ctx *common.CompileContext, channel *Channel, parent *render.Channel) (*render.ProtoChannel, error)
	BuildServer(ctx *common.CompileContext, server *Server, parent *render.Server) (*render.ProtoServer, error)

	BuildMessageBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildOperationBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildChannelBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)
	BuildServerBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error)

	ProtocolName() string
}

var ProtocolBuilders map[string]ProtocolBuilder // TODO: replace the global variable to smth better

func BuildProtoChannelStruct(
	ctx *common.CompileContext,
	source *Channel,
	target *render.Channel,
	protoName, golangName string,
) (*lang.GoStruct, error) {
	chanStruct := lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  golangName,
			Description:   source.Description,
			HasDefinition: true,
		},
		Fields: []lang.GoStructField{
			{Name: "name", Type: &lang.GoSimple{TypeName: "ParamString", Import: ctx.RuntimeModule("")}},
		},
	}

	// Publisher stuff
	if target.IsPublisher {
		ctx.Logger.Trace("Channel publish operation", "proto", protoName)
		chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{
			Name:        "publisher",
			Description: source.Publish.Description,
			Type: &lang.GoSimple{
				TypeName:    "Publisher",
				Import:      ctx.RuntimeModule(protoName),
				IsInterface: true,
			},
		})
	}

	// Subscriber stuff
	if target.IsSubscriber {
		ctx.Logger.Trace("Channel subscribe operation", "proto", protoName)
		chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{
			Name:        "subscriber",
			Description: source.Subscribe.Description,
			Type: &lang.GoSimple{
				TypeName:    "Subscriber",
				Import:      ctx.RuntimeModule(protoName),
				IsInterface: true,
			},
		})
	}

	return &chanStruct, nil
}

func BuildProtoServerStruct(
	ctx *common.CompileContext,
	source *Server,
	target *render.Server,
	protoName string,
) (*lang.GoStruct, error) {
	srvStruct := lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  target.OriginalName,
			Description:   source.Description,
			HasDefinition: true,
		},
	}
	// TODO: handle when protoName is empty (it appears when we build ProtoServer for unsupported protocol)
	// Producer/consumer
	ctx.Logger.Trace("Server producer", "proto", protoName)
	fld := lang.GoStructField{
		Name: "producer",
		Type: &lang.GoSimple{TypeName: "Producer", Import: ctx.RuntimeModule(protoName), IsInterface: true},
	}
	srvStruct.Fields = append(srvStruct.Fields, fld)

	ctx.Logger.Trace("Server consumer", "proto", protoName)
	fld = lang.GoStructField{
		Name: "consumer",
		Type: &lang.GoSimple{TypeName: "Consumer", Import: ctx.RuntimeModule(protoName), IsInterface: true},
	}
	srvStruct.Fields = append(srvStruct.Fields, fld)

	return &srvStruct, nil
}
