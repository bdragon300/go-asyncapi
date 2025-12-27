package kafka

import (
	"encoding/json"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type BindingsBuilder struct{}

func (pb BindingsBuilder) Protocol() string {
	return "kafka"
}

type channelBindings struct {
	Topic              string              `json:"topic,omitzero" yaml:"topic"`
	Partitions         *int                `json:"partitions,omitzero" yaml:"partitions"`
	Replicas           *int                `json:"replicas,omitzero" yaml:"replicas"`
	TopicConfiguration *topicConfiguration `json:"topicConfiguration,omitzero" yaml:"topicConfiguration"`
}

type topicConfiguration struct {
	CleanupPolicy     []string `json:"cleanup.policy,omitzero" yaml:"cleanup.policy"`
	RetentionMs       int      `json:"retention.ms,omitzero" yaml:"retention.ms"`
	RetentionBytes    int      `json:"retention.bytes,omitzero" yaml:"retention.bytes"`
	DeleteRetentionMs int      `json:"delete.retention.ms,omitzero" yaml:"delete.retention.ms"`
	MaxMessageBytes   int      `json:"max.message.bytes,omitzero" yaml:"max.message.bytes"`
}

type operationBindings struct {
	GroupID  any `json:"groupId,omitzero" yaml:"groupId"`   // jsonschema object
	ClientID any `json:"clientId,omitzero" yaml:"clientId"` // jsonschema object
}

func (pb BindingsBuilder) BuildChannelBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
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

func (pb BindingsBuilder) BuildOperationBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
		return
	}

	if bindings.GroupID != nil {
		v, err2 := json.Marshal(bindings.GroupID)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("GroupID", string(v))
	}
	if bindings.ClientID != nil {
		v, err2 := json.Marshal(bindings.ClientID)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("ClientID", string(v))
	}
	return
}

type messageBindings struct {
	Key                     any    `json:"key,omitzero" yaml:"key"` // jsonschema object
	SchemaIDLocation        string `json:"schemaIdLocation,omitzero" yaml:"schemaIdLocation"`
	SchemaIDPayloadEncoding string `json:"schemaIdPayloadEncoding,omitzero" yaml:"schemaIdPayloadEncoding"`
	SchemaLookupStrategy    string `json:"schemaLookupStrategy,omitzero" yaml:"schemaLookupStrategy"`
}

func (pb BindingsBuilder) BuildMessageBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings messageBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Key"}, &lang.GoSimple{TypeName: "MessageBindings", Import: ctx.RuntimeModule(pb.Protocol())})
	if bindings.Key != nil {
		v, err2 := json.Marshal(bindings.Key)
		if err2 != nil {
			err = types.CompileError{Err: err2, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
			return
		}
		jsonVals.Set("Key", string(v))
	}

	return
}

type serverBindings struct {
	SchemaRegistryURL    string `json:"schemaRegistryUrl,omitzero" yaml:"schemaRegistryUrl"`
	SchemaRegistryVendor string `json:"schemaRegistryVendor,omitzero" yaml:"schemaRegistryVendor"`
}

func (pb BindingsBuilder) BuildServerBindings(ctx *compile.Context, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings serverBindings
	if err = types.UnmarshalRawMessageUnion2(rawData, &bindings); err != nil {
		return vals, jsonVals, types.CompileError{Err: err, Path: ctx.CurrentRefPointer(), Proto: pb.Protocol()}
	}
	vals = lang.ConstructGoValue(bindings, nil, &lang.GoSimple{TypeName: "ServerBindings", Import: ctx.RuntimeModule(pb.Protocol())})

	return
}
