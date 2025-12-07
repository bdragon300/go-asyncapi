package nats

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "nats"
}

type operationBindings struct {
	Queue string `json:"queue,omitzero" yaml:"queue"`
}

func (pb ProtoBuilder) BuildChannelBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildOperationBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
		return
	}

	vals = lang.ConstructGoValue(bindings, nil, &lang.GoSimple{TypeName: "OperationBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	return
}

func (pb ProtoBuilder) BuildMessageBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildServerBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
