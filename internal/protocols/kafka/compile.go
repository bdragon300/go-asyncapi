package kafka

import (
	"encoding/json"
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

const protoName = "kafka"

func Register() {
	compile.ProtoServerBuilders[protoName] = BuildServer
	compile.ProtoChannelBuilders[protoName] = BuildChannel
	compile.ProtoServerBuilders["kafka-secure"] = BuildServer // TODO: make a separate kafka-secure protocol
	compile.ProtoChannelBuilders["kafka-secure"] = BuildChannel
}

type channelBindings struct {
	Topic              *string             `json:"topic" yaml:"topic"`
	Partitions         *int                `json:"partitions" yaml:"partitions"`
	Replicas           *int                `json:"replicas" yaml:"replicas"`
	TopicConfiguration *topicConfiguration `json:"topicConfiguration" yaml:"topicConfiguration"`
}

type topicConfiguration struct {
	CleanupPolicy     []string `json:"cleanup.policy" yaml:"cleanup.policy"`
	RetentionMs       *int     `json:"retention.ms" yaml:"retention.ms"`
	RetentionBytes    *int     `json:"retention.bytes" yaml:"retention.bytes"`
	DeleteRetentionMs *int     `json:"delete.retention.ms" yaml:"delete.retention.ms"`
	MaxMessageBytes   *int     `json:"max.message.bytes" yaml:"max.message.bytes"`
}

type operationBindings struct {
	GroupID  *compile.Object `json:"groupId" yaml:"groupId"`
	ClientID *compile.Object `json:"clientId" yaml:"clientId"`
}

type serverBindings struct {
	SchemaRegistryURL    *string `json:"schemaRegistryUrl" yaml:"schemaRegistryUrl"`
	SchemaRegistryVendor *string `json:"schemaRegistryVendor" yaml:"schemaRegistryVendor"`
}

func BuildChannel(ctx *common.CompileContext, channel *compile.Channel, channelKey string) (common.Assembler, error) {
	chanResult := &ProtoChannel{
		Name:  channelKey,
		Topic: channelKey,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, ctx.CurrentObjName(), "KafkaChannel"),
				Description: channel.Description,
				Render:      true,
				Package:     ctx.Stack.Top().PackageKind,
			},
		},
		FallbackMessageType: &assemble.Simple{Type: "any", IsIface: true},
	}

	if channel.Publish != nil {
		fld := assemble.StructField{
			Name:        "publishers",
			Description: channel.Publish.Description,
			Type: &assemble.Array{
				BaseType: assemble.BaseType{Package: ctx.Stack.Top().PackageKind},
				ItemsType: &assemble.Simple{
					Type:    "Publisher",
					Package: common.RuntimePackageKind,
					TypeParamValues: []common.Assembler{
						&assemble.Simple{Type: "OutEnvelope", Package: common.RuntimeKafkaPackageKind},
					},
					IsIface: true,
				},
			},
		}
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, fld)
		chanResult.Publisher = true
		if channel.Publish.Message != nil {
			ref := path.Join(ctx.PathRef(), "publish/message")
			chanResult.PubMessageLink = assemble.NewRefLink[*assemble.Message](ref)
			ctx.Linker.Add(chanResult.PubMessageLink)
		}
		chBindings, chOk := channel.Bindings.Get(protoName)
		opBindings, opOk := channel.Publish.Bindings.Get(protoName)
		if chOk || opOk {
			var err error
			chanResult.PubChannelBindings, err = buildChannelBindings(&chBindings, &opBindings)
			if err != nil {
				return nil, err
			}
		}
	}
	if channel.Subscribe != nil {
		fld := assemble.StructField{
			Name:        "subscribers",
			Description: channel.Subscribe.Description,
			Type: &assemble.Array{
				BaseType: assemble.BaseType{Package: ctx.Stack.Top().PackageKind},
				ItemsType: &assemble.Simple{
					Type:    "Subscriber",
					Package: common.RuntimePackageKind,
					TypeParamValues: []common.Assembler{
						&assemble.Simple{Type: "InEnvelope", Package: common.RuntimeKafkaPackageKind},
					},
					IsIface: true,
				},
			},
		}
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, fld)
		chanResult.Subscriber = true
		if channel.Subscribe.Message != nil {
			ref := path.Join(ctx.PathRef(), "subscribe/message")
			chanResult.SubMessageLink = assemble.NewRefLink[*assemble.Message](ref)
			ctx.Linker.Add(chanResult.SubMessageLink)
		}
		chBindings, chOk := channel.Bindings.Get(protoName)
		opBindings, opOk := channel.Subscribe.Bindings.Get(protoName)
		if chOk || opOk {
			var err error
			chanResult.SubChannelBindings, err = buildChannelBindings(&chBindings, &opBindings)
			if err != nil {
				return nil, err
			}
		}
	}

	return chanResult, nil
}

func buildChannelBindings(chBindings, opBindings *utils.Union2[json.RawMessage, yaml.Node]) (*ProtoChannelBindings, error) {
	res := ProtoChannelBindings{}
	if chBindings != nil {
		var bindings channelBindings
		if err := utils.UnmarhalRawsUnion2(*chBindings, &bindings); err != nil {
			return nil, err
		}
		if bindings.Topic != nil {
			res.StructValues.Set("Topic", *bindings.Topic)
		}
		if bindings.Partitions != nil {
			res.StructValues.Set("Partitions", *bindings.Partitions)
		}
		if bindings.Replicas != nil {
			res.StructValues.Set("Replicas", *bindings.Replicas)
		}
		if bindings.TopicConfiguration != nil {
			tc := bindings.TopicConfiguration
			if lo.Contains(tc.CleanupPolicy, "delete") {
				res.CleanupPolicyStructValue.Set("Delete", true)
			}
			if lo.Contains(tc.CleanupPolicy, "compact") {
				res.CleanupPolicyStructValue.Set("Compact", true)
			}
			if tc.RetentionMs != nil {
				res.StructValues.Set("RetentionMs", *tc.RetentionMs)
			}
			if tc.RetentionBytes != nil {
				res.StructValues.Set("RetentionBytes", *tc.RetentionBytes)
			}
			if tc.DeleteRetentionMs != nil {
				res.StructValues.Set("DeleteRetentionMs", *tc.DeleteRetentionMs)
			}
			if tc.MaxMessageBytes != nil {
				res.StructValues.Set("MaxMessageBytes", *tc.MaxMessageBytes)
			}
		}
	}
	if opBindings != nil {
		var bindings operationBindings
		if err := utils.UnmarhalRawsUnion2(*opBindings, &bindings); err != nil {
			return nil, err
		}
		if bindings.GroupID != nil {
			res.GroupIDArgSchema = "string"
		}
		if bindings.ClientID != nil {
			res.ClientIDArgSchema = "string"
		}
	}

	return &res, nil
}

func BuildServer(ctx *common.CompileContext, server *compile.Server, serverKey string) (common.Assembler, error) {
	const buildProducer = true
	const buildConsumer = true

	srvResult := ProtoServer{
		Name: serverKey,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, ctx.CurrentObjName(), "Server"),
				Description: server.Description,
				Render:      true,
				Package:     ctx.Stack.Top().PackageKind,
			},
		},
	}

	channelsLnks := assemble.NewListCbLink[*assemble.Channel](func(item common.Assembler, path []string) bool {
		ch, ok := item.(*assemble.Channel)
		if !ok {
			return false
		}
		if len(ch.AppliedServers) > 0 {
			return lo.Contains(ch.AppliedServers, serverKey)
		}
		return ch.AppliedToAllServersLinks != nil
	})
	srvResult.ChannelLinkList = channelsLnks
	ctx.Linker.AddMany(channelsLnks)

	if buildProducer {
		fld := assemble.StructField{
			Name: "producer",
			Type: &assemble.Simple{Type: "Producer", Package: common.RuntimeKafkaPackageKind, IsIface: true},
		}
		srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)
		srvResult.Producer = true
	}
	if buildConsumer {
		fld := assemble.StructField{
			Name: "consumer",
			Type: &assemble.Simple{Type: "Consumer", Package: common.RuntimeKafkaPackageKind, IsIface: true},
		}
		srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)
		srvResult.Consumer = true
	}
	if server.Bindings.Len() > 0 {
		if srvBindings, ok := server.Bindings.Get(protoName); ok {
			var bindings serverBindings
			if err := utils.UnmarhalRawsUnion2(srvBindings, &bindings); err != nil {
				return nil, err
			}
			srvResult.Bindings = &ProtoServerBindings{}
			if bindings.SchemaRegistryURL != nil {
				srvResult.Bindings.StructValues.Set("SchemaRegistryURL", *bindings.SchemaRegistryURL)
			}
			if bindings.SchemaRegistryVendor != nil {
				srvResult.Bindings.StructValues.Set("SchemaRegistryVendor", *bindings.SchemaRegistryVendor)
			}
		}
	}

	return srvResult, nil
}
