package rawsocket

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	renderRawSocket "github.com/bdragon300/go-asyncapi/internal/render/rawsocket"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type channelBindings struct {
	LocalAddress string `json:"localAddress" yaml:"localAddress"`
}

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (common.Renderer, error) {
	baseChan, err := pb.BuildBaseProtoChannel(ctx, channel, parent)
	if err != nil {
		return nil, err
	}

	return &renderRawSocket.ProtoChannel{BaseProtoChannel: *baseChan}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathRef(), Proto: pb.ProtoName}
		return
	}

	vals = render.ConstructGoValue(
		bindings, nil, &render.GoSimple{Name: "ChannelBindings", Import: ctx.RuntimeModule(pb.ProtoName)},
	)
	return
}

func (pb ProtoBuilder) BuildOperationBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
