package kafka

import (
	"encoding/json"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type channelBindings struct {
	Topic              string              `json:"topic" yaml:"topic"`
	Partitions         *int                `json:"partitions" yaml:"partitions"`
	Replicas           *int                `json:"replicas" yaml:"replicas"`
	TopicConfiguration *topicConfiguration `json:"topicConfiguration" yaml:"topicConfiguration"`
}

type topicConfiguration struct {
	CleanupPolicy     []string `json:"cleanup.policy" yaml:"cleanup.policy"`
	RetentionMs       int      `json:"retention.ms" yaml:"retention.ms"`
	RetentionBytes    int      `json:"retention.bytes" yaml:"retention.bytes"`
	DeleteRetentionMs int      `json:"delete.retention.ms" yaml:"delete.retention.ms"`
	MaxMessageBytes   int      `json:"max.message.bytes" yaml:"max.message.bytes"`
}

type operationBindings struct {
	GroupID  any `json:"groupId" yaml:"groupId"`   // jsonschema object
	ClientID any `json:"clientId" yaml:"clientId"` // jsonschema object
}

func (pb ProtoBuilder) BuildChannel(ctx *compile.Context, channel *asyncapi.Channel, parent *render.Channel) (*render.ProtoChannel, error) {
	golangName := utils.ToGolangName(parent.OriginalName+lo.Capitalize(pb.Protocol()), true)
	chanStruct := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.Protocol(), golangName)

	chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{Name: "topic", Type: &lang.GoSimple{TypeName: "string"}})

	return &render.ProtoChannel{
		Channel:  parent,
		Type:     chanStruct,
		Protocol: pb.Protocol(),
	}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Partitions", "Replicas", "TopicConfiguration"}, &lang.GoSimple{TypeName: "ChannelBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.Partitions != nil {
		vals.StructValues.Set("Partitions", *bindings.Partitions)
	}
	if bindings.Replicas != nil {
		vals.StructValues.Set("Replicas", *bindings.Replicas)
	}
	if bindings.TopicConfiguration != nil {
		tcVals := lang.ConstructGoValue(*bindings.TopicConfiguration, []string{"CleanupPolicy", "RetentionMs", "DeleteRetentionMs"}, &lang.GoSimple{TypeName: "TopicConfiguration", Import: ctx.RuntimeModule(pb.Protocol())})

		// TopicConfiguration->CleanupPolicy
		if len(bindings.TopicConfiguration.CleanupPolicy) > 0 {
			cpVal := &lang.GoValue{Type: &lang.GoSimple{TypeName: "TopicCleanupPolicy", Import: ctx.RuntimeModule(pb.Protocol())}, EmptyCurlyBrackets: true}
			switch {
			case lo.Contains(bindings.TopicConfiguration.CleanupPolicy, "delete"):
				cpVal.StructValues.Set("Delete", true)
			case lo.Contains(bindings.TopicConfiguration.CleanupPolicy, "compact"):
				cpVal.StructValues.Set("Compact", true)
			}
			if cpVal.StructValues.Len() > 0 {
				tcVals.StructValues.Set("CleanupPolicy", cpVal)
			}
		}
		// TopicConfiguration->RetentionTime
		if bindings.TopicConfiguration.RetentionMs > 0 {
			v := lang.ConstructGoValue(bindings.TopicConfiguration.RetentionMs*int(time.Millisecond), nil, &lang.GoSimple{TypeName: "Duration", Import: "time"})
			tcVals.StructValues.Set("RetentionTime", v)
		}
		// TopicConfiguration->DeleteRetentionTime
		if bindings.TopicConfiguration.DeleteRetentionMs > 0 {
			v := lang.ConstructGoValue(bindings.TopicConfiguration.DeleteRetentionMs*int(time.Millisecond), nil, &lang.GoSimple{TypeName: "Duration", Import: "time"})
			tcVals.StructValues.Set("DeleteRetentionTime", v)
		}
		vals.StructValues.Set("TopicConfiguration", tcVals)
	}

	return
}

func (pb ProtoBuilder) BuildOperationBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
		return
	}

	if bindings.GroupID != nil {
		v, err2 := json.Marshal(bindings.GroupID)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("GroupID", string(v))
	}
	if bindings.ClientID != nil {
		v, err2 := json.Marshal(bindings.ClientID)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentPositionRef(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("ClientID", string(v))
	}
	return
}
