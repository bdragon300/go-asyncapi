package kafka

import (
	"encoding/json"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type serverBindings struct {
	SchemaRegistryURL    string `json:"schemaRegistryUrl" yaml:"schemaRegistryUrl"`
	SchemaRegistryVendor string `json:"schemaRegistryVendor" yaml:"schemaRegistryVendor"`
}

func (pb ProtoBuilder) BuildServerBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings serverBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		return vals, jsonVals, types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
	}
	vals = lang.ConstructGoValue(bindings, nil, &lang.GoSimple{TypeName: "ServerBindings", Import: ctx.RuntimeModule(pb.Protocol())})

	return
}
