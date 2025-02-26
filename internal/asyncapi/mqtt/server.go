package mqtt

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

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

func (pb ProtoBuilder) BuildServerBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings serverBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		return vals, jsonVals, types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.ProtoName}
	}
	vals = lang.ConstructGoValue(bindings, []string{"LastWill"}, &lang.GoSimple{TypeName: "ServerBindings", Import: ctx.RuntimeModule(pb.ProtoName)})
	if bindings.LastWill != nil {
		vals.StructValues.Set("LastWill", lang.ConstructGoValue(*bindings.LastWill, []string{}, &lang.GoSimple{TypeName: "LastWill", Import: ctx.RuntimeModule(pb.ProtoName)}))
	}

	return
}
