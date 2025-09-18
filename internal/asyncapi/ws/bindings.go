package ws

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "ws"
}

type channelBindings struct {
	Method  string `json:"method" yaml:"method"`
	Query   any    `json:"query" yaml:"query"`     // jsonschema object
	Headers any    `json:"headers" yaml:"headers"` // jsonschema object
}

func (pb ProtoBuilder) BuildChannelBindings(
	ctx *compile.Context,
	rawData types.Union2[json.RawMessage, yaml.Node],
) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Query", "Headers"}, &lang.GoSimple{TypeName: "ChannelBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.Query != nil {
		v, err2 := json.Marshal(bindings.Query)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("Query", string(v))
	}
	if bindings.Headers != nil {
		v, err2 := json.Marshal(bindings.Headers)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("Headers", string(v))
	}

	return
}

func (pb ProtoBuilder) BuildOperationBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildMessageBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildServerBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
