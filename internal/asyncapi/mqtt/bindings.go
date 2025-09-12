package mqtt

import (
	"encoding/json"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "mqtt"
}

type operationBindings struct {
	QoS                   int  `json:"qos" yaml:"qos"`
	Retain                bool `json:"retain" yaml:"retain"`
	MessageExpiryInterval int  `json:"messageExpiryInterval" yaml:"messageExpiryInterval"` // Seconds
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

type messageBindings struct {
	PayloadFormatIndicator *int   `json:"payloadFormatIndicator" yaml:"payloadFormatIndicator"`
	CorrelationData        any    `json:"correlationData" yaml:"correlationData"` // jsonschema object
	ContentType            string `json:"contentType" yaml:"contentType"`
	ResponseTopic          string `json:"responseTopic" yaml:"responseTopic"`
}

func (pb ProtoBuilder) BuildMessageBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings messageBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"CorrelationData", "PayloadFormatIndicator"}, &lang.GoSimple{TypeName: "MessageBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.PayloadFormatIndicator != nil {
		switch *bindings.PayloadFormatIndicator {
		case 0:
			vals.StructValues.Set("PayloadFormatIndicator", &lang.GoSimple{TypeName: "PayloadFormatIndicatorUnspecified", Import: ctx.RuntimeModule(pb.Protocol())})
		case 1:
			vals.StructValues.Set("PayloadFormatIndicator", &lang.GoSimple{TypeName: "PayloadFormatIndicatorUTF8", Import: ctx.RuntimeModule(pb.Protocol())})
		}
	}
	if bindings.CorrelationData != nil {
		v, err2 := json.Marshal(bindings.CorrelationData)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("CorrelationData", string(v))
	}

	return
}

type serverBindings struct {
	ClientID              string    `json:"clientId" yaml:"clientId"`
	CleanSession          bool      `json:"cleanSession" yaml:"cleanSession"`
	LastWill              *lastWill `json:"lastWill" yaml:"lastWill"`
	KeepAlive             int       `json:"keepAlive" yaml:"keepAlive"`
	SessionExpiryInterval int       `json:"sessionExpiryInterval" yaml:"sessionExpiryInterval"`
	MaximumPacketSize     int       `json:"maximumPacketSize" yaml:"maximumPacketSize"`
}

type lastWill struct {
	Topic   string `json:"topic" yaml:"topic"`
	QoS     int    `json:"qos" yaml:"qos"`
	Message string `json:"message" yaml:"message"`
	Retain  bool   `json:"retain" yaml:"retain"`
}

func (pb ProtoBuilder) BuildServerBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings serverBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		return vals, jsonVals, types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
	}
	vals = lang.ConstructGoValue(bindings, []string{"LastWill"}, &lang.GoSimple{TypeName: "ServerBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.LastWill != nil {
		vals.StructValues.Set("LastWill", lang.ConstructGoValue(*bindings.LastWill, []string{}, &lang.GoSimple{TypeName: "LastWill", Import: ctx.RuntimeModule(pb.Protocol())}))
	}

	return
}
