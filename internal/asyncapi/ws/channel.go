package ws

import (
	"encoding/json"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

type channelBindings struct {
	Method  string `json:"method" yaml:"method"`
	Query   any    `json:"query" yaml:"query"`     // jsonschema object
	Headers any    `json:"headers" yaml:"headers"` // jsonschema object
}

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (*render.ProtoChannel, error) {
	golangName := utils.ToGolangName(parent.OriginalName + lo.Capitalize(pb.ProtoName), true)
	chanStruct, err := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.ProtoName, golangName)
	if err != nil {
		return nil, err
	}

	chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{Name: "topic", Type: &lang.GoSimple{TypeName: "string"}})

	return &render.ProtoChannel{
		Channel:         parent,
		Type:            chanStruct,
		ProtoName:       pb.ProtoName,
	}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(
	ctx *common.CompileContext,
	rawData types.Union2[json.RawMessage, yaml.Node],
) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Query", "Headers"}, &lang.GoSimple{TypeName: "ChannelBindings", Import: ctx.RuntimeModule(pb.ProtoName)})
	if bindings.Query != nil {
		v, err2 := json.Marshal(bindings.Query)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
			return
		}
		jsonVals.Set("Query", string(v))
	}
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

func (pb ProtoBuilder) BuildOperationBindings(_ *common.CompileContext, _ types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	return
}
