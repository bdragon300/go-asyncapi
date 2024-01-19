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

type serverBindings struct {
	ClientID     string    `json:"clientId" yaml:"clientId"`
	CleanSession bool      `json:"cleanSession" yaml:"cleanSession"`
	LastWill     *lastWill `json:"lastWill" yaml:"lastWill"`
	KeepAlive    int       `json:"keepAlive" yaml:"keepAlive"`
}

type lastWill struct {
	Topic   string `json:"topic" yaml:"topic"`
	QoS     int    `json:"qos" yaml:"qos"`
	Message string `json:"message" yaml:"message"`
	Retain  bool   `json:"retain" yaml:"retain"`
}

func (pb ProtoBuilder) BuildServer(ctx *common.CompileContext, server *asyncapi.Server, serverKey string, abstractServer *render.Server) (common.Renderer, error) {
	baseServer, err := pb.BuildBaseProtoServer(ctx, server, serverKey, abstractServer)
	if err != nil {
		return nil, err
	}
	return &renderMqtt.ProtoServer{BaseProtoServer: *baseServer}, nil
}

func (pb ProtoBuilder) BuildServerBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings serverBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		return vals, jsonVals, types.CompileError{Err: err, Path: ctx.PathRef(), Proto: pb.ProtoName}
	}
	vals = render.ConstructGoValue(
		bindings, []string{"LastWill"}, &render.GoSimple{Name: "ServerBindings", Import: ctx.RuntimeModule(pb.ProtoName)},
	)
	if bindings.LastWill != nil {
		vals.StructVals.Set("LastWill", render.ConstructGoValue(
			*bindings.LastWill, []string{}, &render.GoSimple{Name: "LastWill", Import: ctx.RuntimeModule(pb.ProtoName)},
		))
	}

	return
}
