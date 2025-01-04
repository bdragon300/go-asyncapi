package mqtt

import (
	"encoding/json"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type messageBindings struct {
	PayloadFormatIndicator *int	`json:"payloadFormatIndicator" yaml:"payloadFormatIndicator"`
	CorrelationData any	`json:"correlationData" yaml:"correlationData"` // jsonschema object
	ContentType string	`json:"contentType" yaml:"contentType"`
	ResponseTopic string	`json:"responseTopic" yaml:"responseTopic"`
}

func (pb ProtoBuilder) BuildMessageBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings messageBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"CorrelationData", "PayloadFormatIndicator"}, &lang.GoSimple{TypeName: "MessageBindings", Import: ctx.RuntimeModule(pb.ProtoName)})
	if bindings.PayloadFormatIndicator != nil {
		switch *bindings.PayloadFormatIndicator {
		case 0:
			vals.StructValues.Set("PayloadFormatIndicator", &lang.GoSimple{TypeName: "PayloadFormatIndicatorUnspecified", Import: ctx.RuntimeModule(pb.ProtoName)})
		case 1:
			vals.StructValues.Set("PayloadFormatIndicator", &lang.GoSimple{TypeName: "PayloadFormatIndicatorUTF8", Import: ctx.RuntimeModule(pb.ProtoName)})
		}
	}
	if bindings.CorrelationData != nil {
		v, err2 := json.Marshal(bindings.CorrelationData)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
			return
		}
		jsonVals.Set("CorrelationData", string(v))
	}

	return
}
