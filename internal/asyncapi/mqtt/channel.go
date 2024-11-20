package mqtt

import (
	"encoding/json"
	renderMqtt "github.com/bdragon300/go-asyncapi/internal/render/mqtt"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type operationBindings struct {
	QoS    int  `json:"qos" yaml:"qos"`
	Retain bool `json:"retain" yaml:"retain"`
}

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (common.Renderer, error) {
	golangName := parent.GolangName + pb.ProtoTitle
	chanStruct, err := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.ProtoName, golangName)
	if err != nil {
		return nil, err
	}

	chanStruct.Fields = append(chanStruct.Fields, render.GoStructField{Name: "topic", Type: &render.GoSimple{Name: "string"}})

	return &renderMqtt.ProtoChannel{
		Channel: parent,
		GolangNameProto: golangName,
		Struct: chanStruct,
		ProtoName: pb.ProtoName,
		ProtoTitle: pb.ProtoTitle,
	}, nil
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
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}
	vals = render.ConstructGoValue(
		bindings, nil, &render.GoSimple{Name: "OperationBindings", Import: ctx.RuntimeModule(pb.ProtoName)},
	)
	return
}
