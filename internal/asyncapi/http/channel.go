package http

import (
	"encoding/json"
	renderHttp "github.com/bdragon300/go-asyncapi/internal/render/http"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type operationBindings struct {
	Type   string `json:"type" yaml:"type"`
	Method string `json:"method" yaml:"method"`
	Query  any    `json:"query" yaml:"query"` // jsonschema object
}

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (common.Renderer, error) {
	golangName := parent.GolangName + pb.ProtoTitle
	chanStruct, err := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.ProtoName, golangName)
	if err != nil {
		return nil, err
	}

	return &renderHttp.ProtoChannel{
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

func (pb ProtoBuilder) BuildOperationBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	vals = render.ConstructGoValue(
		bindings, []string{"Query"}, &render.GoSimple{Name: "OperationBindings", Import: ctx.RuntimeModule(pb.ProtoName)},
	)
	if bindings.Query != nil {
		v, err2 := json.Marshal(bindings.Query)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
			return
		}
		jsonVals.Set("Query", string(v))
	}
	return
}
