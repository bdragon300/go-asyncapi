package asyncapi

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
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

func BuildProtoChannelStruct(
	ctx *common.CompileContext,
	source *Channel,
	target *render.Channel,
	protoName, golangName string,
) (*render.GoStruct, error) {
	chanStruct := render.GoStruct{
		BaseType: render.BaseType{
			Name:         golangName,
			Description:  source.Description,
			DirectRender: true,
			Import:       ctx.CurrentPackage(),
		},
		Fields: []render.GoStructField{
			{Name: "name", Type: &render.GoSimple{Name: "ParamString", Import: ctx.RuntimeModule("")}},
		},
	}

	// Publisher stuff
	if target.Publisher {
		ctx.Logger.Trace("Channel publish operation", "proto", protoName)
		chanStruct.Fields = append(chanStruct.Fields, render.GoStructField{
			Name:        "publisher",
			Description: source.Publish.Description,
			Type: &render.GoSimple{
				Name:    "Publisher",
				Import:  ctx.RuntimeModule(protoName),
				IsIface: true,
			},
		})
	}

	// Subscriber stuff
	if target.Subscriber {
		ctx.Logger.Trace("Channel subscribe operation", "proto", protoName)
		chanStruct.Fields = append(chanStruct.Fields, render.GoStructField{
			Name:        "subscriber",
			Description: source.Subscribe.Description,
			Type: &render.GoSimple{
				Name:    "Subscriber",
				Import:  ctx.RuntimeModule(protoName),
				IsIface: true,
			},
		})
	}

	return &chanStruct, nil
}

func BuildProtoServerStruct(
	ctx *common.CompileContext,
	source *Server,
	target *render.Server,
	protoName, protoTitle string,
) (*render.GoStruct, error) {
	srvStruct := render.GoStruct{
		BaseType: render.BaseType{
			Name:         target.GolangName,
			Description:  source.Description,
			DirectRender: true,
			Import:       ctx.CurrentPackage(),
		},
	}

	// Producer/consumer
	ctx.Logger.Trace("Server producer", "proto", protoName)
	fld := render.GoStructField{
		Name: "producer",
		Type: &render.GoSimple{Name: "Producer", Import: ctx.RuntimeModule(protoName), IsIface: true},
	}
	srvStruct.Fields = append(srvStruct.Fields, fld)

	ctx.Logger.Trace("Server consumer", "proto", protoName)
	fld = render.GoStructField{
		Name: "consumer",
		Type: &render.GoSimple{Name: "Consumer", Import: ctx.RuntimeModule(protoName), IsIface: true},
	}
	srvStruct.Fields = append(srvStruct.Fields, fld)

	return &srvStruct, nil
}
