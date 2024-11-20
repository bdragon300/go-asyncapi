package ip

import (
	"encoding/json"
	renderIP "github.com/bdragon300/go-asyncapi/internal/render/ip"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (common.Renderer, error) {
	golangName := parent.GolangName + pb.ProtoTitle
	chanStruct, err := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.ProtoName, golangName)
	if err != nil {
		return nil, err
	}

	return &renderIP.ProtoChannel{
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

func (pb ProtoBuilder) BuildOperationBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
