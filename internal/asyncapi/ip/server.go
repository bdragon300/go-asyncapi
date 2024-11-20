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

func (pb ProtoBuilder) BuildServer(ctx *common.CompileContext, server *asyncapi.Server, parent *render.Server) (common.Renderer, error) {
	baseServer, err := asyncapi.BuildProtoServerStruct(ctx, server, parent, pb.ProtoName, pb.ProtoTitle)
	if err != nil {
		return nil, err
	}
	return &renderIP.ProtoServer{BaseProtoServer: *baseServer}, nil
}

func (pb ProtoBuilder) BuildServerBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
