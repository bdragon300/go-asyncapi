package kafka

import (
	"encoding/json"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type serverBindings struct {
	SchemaRegistryURL    string `json:"schemaRegistryUrl" yaml:"schemaRegistryUrl"`
	SchemaRegistryVendor string `json:"schemaRegistryVendor" yaml:"schemaRegistryVendor"`
}

func (pb ProtoBuilder) BuildServer(ctx *common.CompileContext, server *asyncapi.Server, parent *render.Server) (*render.ProtoServer, error) {
	baseServer, err := asyncapi.BuildProtoServerStruct(ctx, server, parent, pb.ProtoName)
	if err != nil {
		return nil, err
	}
	return &render.ProtoServer{
		Server:    parent,
		Type:      baseServer,
		ProtoName: pb.ProtoName,
	}, nil
}

func (pb ProtoBuilder) BuildServerBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings serverBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		return vals, jsonVals, types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
	}
	vals = lang.ConstructGoValue(bindings, nil, &lang.GoSimple{Name: "ServerBindings", Import: ctx.RuntimeModule(pb.ProtoName)})

	return
}
