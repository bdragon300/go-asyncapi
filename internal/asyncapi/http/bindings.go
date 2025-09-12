package http

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "http"
}

type operationBindings struct {
	Method string `json:"method" yaml:"method"`
	Query  any    `json:"query" yaml:"query"` // jsonschema object
}

func (pb ProtoBuilder) BuildChannelBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}

func (pb ProtoBuilder) BuildOperationBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Query"}, &lang.GoSimple{TypeName: "OperationBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.Query != nil {
		v, err2 := json.Marshal(bindings.Query)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("Query", string(v))
	}
	return
}

type messageBindings struct {
	Headers    any `json:"headers" yaml:"headers"` // jsonschema object
	StatusCode int `json:"statusCode" yaml:"statusCode"`
}

func (pb ProtoBuilder) BuildMessageBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings messageBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Headers"}, &lang.GoSimple{TypeName: "MessageBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.Headers != nil {
		v, err2 := json.Marshal(bindings.Headers)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("Headers", string(v))
	}

	return
}

func (pb ProtoBuilder) BuildServerBindings(_ *compile.Context, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
