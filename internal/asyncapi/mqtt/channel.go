package mqtt

import (
	"encoding/json"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type operationBindings struct {
	QoS                   int  `json:"qos" yaml:"qos"`
	Retain                bool `json:"retain" yaml:"retain"`
	MessageExpiryInterval int  `json:"messageExpiryInterval" yaml:"messageExpiryInterval"` // Seconds
}

func (pb ProtoBuilder) BuildChannel(ctx *compile.Context, channel *asyncapi.Channel, parent *render.Channel) (*render.ProtoChannel, error) {
	golangName := utils.ToGolangName(parent.OriginalName+lo.Capitalize(pb.Protocol()), true)
	chanStruct := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.Protocol(), golangName)

	chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{Name: "topic", Type: &lang.GoSimple{TypeName: "string"}})

	return &render.ProtoChannel{
		Channel:  parent,
		Type:     chanStruct,
		Protocol: pb.Protocol(),
	}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildOperationBindings(
	ctx *compile.Context,
	rawData types.Union2[json.RawMessage, yaml.Node],
) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
		return
	}
	vals = lang.ConstructGoValue(bindings, []string{"MessageExpiryInterval"}, &lang.GoSimple{TypeName: "OperationBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.MessageExpiryInterval > 0 {
		v := lang.ConstructGoValue(bindings.MessageExpiryInterval*int(time.Second), nil, &lang.GoSimple{TypeName: "Duration", Import: "time"})
		vals.StructValues.Set("MessageExpiryInterval", v)
	}
	return
}
