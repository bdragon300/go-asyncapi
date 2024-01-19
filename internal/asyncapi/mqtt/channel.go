package mqtt

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	renderMqtt "github.com/bdragon300/go-asyncapi/internal/render/mqtt"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type operationBindings struct {
	QoS    int  `json:"qos" yaml:"qos"`
	Retain bool `json:"retain" yaml:"retain"`
}

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, channelKey string, abstractChannel *render.Channel) (common.Renderer, error) {
	baseChan, err := pb.BuildBaseProtoChannel(ctx, channel, channelKey, abstractChannel)
	if err != nil {
		return nil, err
	}

	baseChan.Struct.Fields = append(baseChan.Struct.Fields, render.GoStructField{Name: "topic", Type: &render.GoSimple{Name: "string"}})

	return &renderMqtt.ProtoChannel{BaseProtoChannel: *baseChan}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildOperationBindings(
	ctx *common.CompileContext,
	rawData types.Union2[json.RawMessage, yaml.Node],
) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathRef(), Proto: pb.ProtoName}
		return
	}
	vals = render.ConstructGoValue(
		bindings, nil, &render.GoSimple{Name: "OperationBindings", Import: ctx.RuntimeModule(pb.ProtoName)},
	)
	return
}
