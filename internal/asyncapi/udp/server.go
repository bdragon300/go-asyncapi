package udp

import (
	"encoding/json"

	renderUDP "github.com/bdragon300/go-asyncapi/internal/render/udp"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

func (pb ProtoBuilder) BuildServer(ctx *common.CompileContext, server *asyncapi.Server, parent *render.Server) (common.Renderer, error) {
	baseServer, err := pb.BuildBaseProtoServer(ctx, server, parent)
	if err != nil {
		return nil, err
	}
	return &renderUDP.ProtoServer{BaseProtoServer: *baseServer}, nil
}

func (pb ProtoBuilder) BuildServerBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
