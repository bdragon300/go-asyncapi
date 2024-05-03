package ip

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	renderIP "github.com/bdragon300/go-asyncapi/internal/render/ip"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (common.Renderer, error) {
	baseChan, err := pb.BuildBaseProtoChannel(ctx, channel, parent)
	if err != nil {
		return nil, err
	}

	return &renderIP.ProtoChannel{BaseProtoChannel: *baseChan}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildOperationBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
