package redis

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

func (pb ProtoBuilder) BuildChannel(ctx *compile.Context, channel *asyncapi.Channel, parent *render.Channel) (*render.ProtoChannel, error) {
	golangName := utils.ToGolangName(parent.OriginalName+lo.Capitalize(pb.Protocol()), true)
	chanStruct := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.Protocol(), golangName)

	return &render.ProtoChannel{
		Channel:  parent,
		Type:     chanStruct,
		Protocol: pb.Protocol(),
	}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildOperationBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
