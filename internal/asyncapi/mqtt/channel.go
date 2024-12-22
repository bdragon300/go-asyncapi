package mqtt

import (
	"encoding/json"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"

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

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (*render.ProtoChannel, error) {
	golangName := utils.ToGolangName(parent.OriginalName + lo.Capitalize(pb.ProtoName), true)
	chanStruct, err := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.ProtoName, golangName)
	if err != nil {
		return nil, err
	}

	chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{Name: "topic", Type: &lang.GoSimple{TypeName: "string"}})

	return &render.ProtoChannel{
		Channel:  parent,
		Type:     chanStruct,
		Protocol: pb.ProtoName,
	}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildOperationBindings(
	ctx *common.CompileContext,
	rawData types.Union2[json.RawMessage, yaml.Node],
) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}
	vals = lang.ConstructGoValue(bindings, nil, &lang.GoSimple{TypeName: "OperationBindings", Import: ctx.RuntimeModule(pb.ProtoName)})
	return
}
