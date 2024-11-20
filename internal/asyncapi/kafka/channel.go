package kafka

import (
	"encoding/json"
	renderKafka "github.com/bdragon300/go-asyncapi/internal/render/kafka"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
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

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (common.Renderer, error) {
	golangName := parent.GolangName + pb.ProtoTitle
	chanStruct, err := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.ProtoName, golangName)
	if err != nil {
		return nil, err
	}

	chanStruct.Fields = append(chanStruct.Fields, render.GoStructField{Name: "topic", Type: &render.GoSimple{Name: "string"}})

	return &renderKafka.ProtoChannel{
		Channel: parent,
		GolangNameProto: golangName,
		Struct: chanStruct,
		ProtoName: pb.ProtoName,
		ProtoTitle: pb.ProtoTitle,
	}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	vals = render.ConstructGoValue(
		bindings, []string{"Partitions", "Replicas", "TopicConfiguration"}, &render.GoSimple{Name: "ChannelBindings", Import: ctx.RuntimeModule(pb.ProtoName)},
	)
	if bindings.Partitions != nil {
		vals.StructVals.Set("Partitions", *bindings.Partitions)
	}
	if bindings.Replicas != nil {
		vals.StructVals.Set("Replicas", *bindings.Replicas)
	}
	if bindings.TopicConfiguration != nil {
		tcVals := render.ConstructGoValue(
			*bindings.TopicConfiguration,
			[]string{"CleanupPolicy", "RetentionMs", "DeleteRetentionMs"},
			&render.GoSimple{Name: "TopicConfiguration", Import: ctx.RuntimeModule(pb.ProtoName)},
		)

		// TopicConfiguration->CleanupPolicy
		if len(bindings.TopicConfiguration.CleanupPolicy) > 0 {
			cpVal := &render.GoValue{Type: &render.GoSimple{Name: "TopicCleanupPolicy", Import: ctx.RuntimeModule(pb.ProtoName)}, NilCurlyBrakets: true}
			switch {
			case lo.Contains(bindings.TopicConfiguration.CleanupPolicy, "delete"):
				cpVal.StructVals.Set("Delete", true)
			case lo.Contains(bindings.TopicConfiguration.CleanupPolicy, "compact"):
				cpVal.StructVals.Set("Compact", true)
			}
			if cpVal.StructVals.Len() > 0 {
				tcVals.StructVals.Set("CleanupPolicy", cpVal)
			}
		}
		// TopicConfiguration->RetentionMs
		if bindings.TopicConfiguration.RetentionMs > 0 {
			v := render.ConstructGoValue(
				bindings.TopicConfiguration.RetentionMs*int(time.Millisecond),
				nil,
				&render.GoSimple{Name: "Duration", Import: "time"},
			)
			tcVals.StructVals.Set("RetentionMs", v)
		}
		// TopicConfiguration->DeleteRetentionMs
		if bindings.TopicConfiguration.DeleteRetentionMs > 0 {
			v := render.ConstructGoValue(
				bindings.TopicConfiguration.DeleteRetentionMs*int(time.Millisecond),
				nil,
				&render.GoSimple{Name: "Duration", Import: "time"},
			)
			tcVals.StructVals.Set("DeleteRetentionMs", v)
		}
		vals.StructVals.Set("TopicConfiguration", tcVals)
	}

	return
}

func (pb ProtoBuilder) BuildOperationBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	if bindings.GroupID != nil {
		v, err2 := json.Marshal(bindings.GroupID)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
			return
		}
		jsonVals.Set("GroupID", string(v))
	}
	if bindings.ClientID != nil {
		v, err2 := json.Marshal(bindings.ClientID)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
			return
		}
		jsonVals.Set("ClientID", string(v))
	}
	return
}
