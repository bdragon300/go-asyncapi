package tcp

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

func (pb ProtoBuilder) BuildServer(ctx *common.CompileContext, server *asyncapi.Server, parent *render.Server) (*render.ProtoServer, error) {
	baseServer, err := asyncapi.BuildProtoServerStruct(ctx, server, parent, pb.ProtoName)
	if err != nil {
		return nil, err
	}
	return &render.ProtoServer{
		Server: parent,
		Type:   baseServer,
	}, nil
}

func (pb ProtoBuilder) BuildServerBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
