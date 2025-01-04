package http

import (
	"encoding/json"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type messageBindings struct {
	Headers any `json:"headers" yaml:"headers"` // jsonschema object
	StatusCode int `json:"statusCode" yaml:"statusCode"`
}

func (pb ProtoBuilder) BuildMessageBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings messageBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Headers"}, &lang.GoSimple{TypeName: "MessageBindings", Import: ctx.RuntimeModule(pb.ProtoName)})
	if bindings.Headers != nil {
		v, err2 := json.Marshal(bindings.Headers)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
			return
		}
		jsonVals.Set("Headers", string(v))
	}

	return
}
