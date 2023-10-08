package kafka

import (
	"encoding/json"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/samber/lo"
)

const protoName = "kafka"

func Register() {
	compile.ProtoServerBuilders[protoName] = BuildServer
	compile.ProtoChannelBuilders[protoName] = BuildChannel
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
	GroupID  any `json:"groupId" yaml:"groupId"`   // TODO: here is jsonschema, figure out how it's needed to be represented in bindings
	ClientID any `json:"clientId" yaml:"clientId"` //
}

type serverBindings struct {
	SchemaRegistryURL    *string `json:"schemaRegistryUrl" yaml:"schemaRegistryUrl"`
	SchemaRegistryVendor *string `json:"schemaRegistryVendor" yaml:"schemaRegistryVendor"`
}

func BuildChannel(ctx *common.CompileContext, channel *compile.Channel, channelKey string) (common.Assembler, error) {
	paramsLnk := assemble.NewListCbLink[*assemble.Parameter](func(item common.Assembler, path []string) bool {
		par, ok := item.(*assemble.Parameter)
		if !ok {
			return false
		}
		_, ok = channel.Parameters.Get(par.Name)
		return ok
	})
	ctx.Linker.AddMany(paramsLnk)

	chanResult := &ProtoChannel{
		Name: channelKey,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, ctx.CurrentObjName(), "Kafka"),
				Description: channel.Description,
				Render:      true,
				Package:     ctx.Stack.Top().PackageKind,
			},
			Fields: []assemble.StructField{
				{Name: "name", Type: &assemble.Simple{Type: "ParamString", Package: common.RuntimePackageKind}},
				{Name: "topic", Type: &assemble.Simple{Type: "string"}},
			},
		},
		FallbackMessageType: &assemble.Simple{Type: "any", IsIface: true},
	}

	// FIXME: remove in favor of the non-proto channel
	if channel.Parameters.Len() > 0 {
		chanResult.ParametersStructNoAssemble = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    compile.GenerateGolangTypeName(ctx, ctx.CurrentObjName(), "Parameters"),
				Render:  true,
				Package: ctx.Stack.Top().PackageKind,
			},
			Fields: nil,
		}
		for _, paramName := range channel.Parameters.Keys() {
			ref := path.Join(ctx.PathRef(), "parameters", paramName)
			lnk := assemble.NewRefLinkAsGolangType(ref)
			ctx.Linker.Add(lnk)
			chanResult.ParametersStructNoAssemble.Fields = append(chanResult.ParametersStructNoAssemble.Fields, assemble.StructField{
				Name: utils.ToGolangName(paramName, true),
				Type: lnk,
			})
		}
	}

	// Interface to match servers bound with a channel
	var ifaceFirstMethodParams []assemble.FuncParam
	if chanResult.ParametersStructNoAssemble != nil {
		ifaceFirstMethodParams = append(ifaceFirstMethodParams, assemble.FuncParam{
			Name: "params",
			Type: &assemble.Simple{Type: chanResult.ParametersStructNoAssemble.Name, Package: ctx.Stack.Top().PackageKind},
		})
	}
	chanResult.ServerIface = &assemble.Interface{
		BaseType: assemble.BaseType{
			Name:    utils.ToLowerFirstLetter(chanResult.Struct.Name + "Server"),
			Render:  true,
			Package: ctx.Stack.Top().PackageKind,
		},
		Methods: []assemble.FunctionSignature{
			{
				Name: "Open" + chanResult.Struct.Name,
				Args: ifaceFirstMethodParams,
				Return: []assemble.FuncParam{
					{Type: &assemble.Simple{Type: chanResult.Struct.Name, Package: ctx.Stack.Top().PackageKind}, Pointer: true},
					{Type: &assemble.Simple{Type: "error"}},
				},
			},
		},
	}

	// Publisher stuff
	if channel.Publish != nil {
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, assemble.StructField{
			Name:        "publisher",
			Description: channel.Publish.Description,
			Type: &assemble.Simple{
				Type:    "Publisher",
				Package: common.RuntimeKafkaPackageKind,
				IsIface: true,
			},
		})
		chanResult.Publisher = true
		if channel.Publish.Message != nil {
			ref := path.Join(ctx.PathRef(), "publish/message")
			chanResult.PubMessageLink = assemble.NewRefLink[*assemble.Message](ref)
			ctx.Linker.Add(chanResult.PubMessageLink)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, assemble.FunctionSignature{
			Name: "Producer",
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: &assemble.Simple{Type: "Producer", Package: common.RuntimeKafkaPackageKind, IsIface: true}},
			},
		})
	}

	// Subscriber stuff
	if channel.Subscribe != nil {
		chanResult.Struct.Fields = append(chanResult.Struct.Fields, assemble.StructField{
			Name:        "subscriber",
			Description: channel.Subscribe.Description,
			Type: &assemble.Simple{
				Type:    "Subscriber",
				Package: common.RuntimeKafkaPackageKind,
				IsIface: true,
			},
		})
		chanResult.Subscriber = true
		if channel.Subscribe.Message != nil {
			ref := path.Join(ctx.PathRef(), "subscribe/message")
			chanResult.SubMessageLink = assemble.NewRefLink[*assemble.Message](ref)
			ctx.Linker.Add(chanResult.SubMessageLink)
		}
		chanResult.ServerIface.Methods = append(chanResult.ServerIface.Methods, assemble.FunctionSignature{
			Name: "Consumer",
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: &assemble.Simple{Type: "Consumer", Package: common.RuntimeKafkaPackageKind, IsIface: true}},
			},
		})
	}

	if bindings, ok, err := buildChannelBindings(channel); err != nil {
		return nil, err
	} else if ok {
		chanResult.BindingsStructNoAssemble = &assemble.Struct{ // TODO: remove in favor of parent channel
			BaseType: assemble.BaseType{
				Name:    compile.GenerateGolangTypeName(ctx, ctx.CurrentObjName(), "Bindings"),
				Render:  true,
				Package: ctx.Stack.Top().PackageKind,
			},
		}
		chanResult.BindingsValues = &bindings
	}

	return chanResult, nil
}

func buildChannelBindings(channel *compile.Channel) (res ProtoChannelBindings, hasBindings bool, err error) {
	if chBindings, ok := channel.Bindings.Get(protoName); ok {
		var bindings channelBindings
		hasBindings = true
		if err = utils.UnmarhalRawsUnion2(chBindings, &bindings); err != nil {
			return
		}
		marshalFields := []string{"Topic", "Partitions", "Replicas"}
		if err = utils.StructToOrderedMap(bindings, &res.StructValues, marshalFields); err != nil {
			return
		}

		if bindings.TopicConfiguration != nil {
			tc := bindings.TopicConfiguration
			marshalFields = []string{"RetentionMs", "RetentionBytes", "DeleteRetentionMs", "MaxMessageBytes"}
			if err = utils.StructToOrderedMap(*bindings.TopicConfiguration, &res.StructValues, marshalFields); err != nil {
				return
			}
			if lo.Contains(tc.CleanupPolicy, "delete") {
				res.CleanupPolicyStructValue.Set("Delete", true)
			}
			if lo.Contains(tc.CleanupPolicy, "compact") {
				res.CleanupPolicyStructValue.Set("Compact", true)
			}
		}
	}

	if channel.Publish != nil {
		if opBindings, ok := channel.Publish.Bindings.Get(protoName); ok {
			hasBindings = true
			if res.PublisherValues, err = buildOperationBuilding(opBindings); err != nil {
				return
			}
		}
	}

	if channel.Subscribe != nil {
		if opBindings, ok := channel.Subscribe.Bindings.Get(protoName); ok {
			hasBindings = true
			if res.SubscriberValues, err = buildOperationBuilding(opBindings); err != nil {
				return
			}
		}
	}

	return
}

func buildOperationBuilding(opBindings utils.Union2[json.RawMessage, yaml.Node]) (res utils.OrderedMap[string, any], err error) {
	var bindings operationBindings
	if err = utils.UnmarhalRawsUnion2(opBindings, &bindings); err != nil {
		return
	}
	// TODO: represent jsonschemas in bindings somehow

	return
}

func BuildServer(ctx *common.CompileContext, server *compile.Server, serverKey string) (common.Assembler, error) {
	const buildProducer = true
	const buildConsumer = true

	srvResult := ProtoServer{
		Name:            serverKey,
		URL:             server.URL,
		ProtocolVersion: server.ProtocolVersion,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, ctx.CurrentObjName(), ""),
				Description: server.Description,
				Render:      true,
				Package:     ctx.Stack.Top().PackageKind,
			},
		},
	}

	// Server variables
	for _, v := range server.Variables.Entries() {
		srvResult.Variables.Set(v.Key, ProtoServerVariable{
			ArgName:     utils.ToGolangName(v.Key, false),
			Enum:        v.Value.Enum,
			Default:     v.Value.Default,
			Description: v.Value.Description,
		})
	}

	// Channels which are connected to this server
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

	// Producer/consumer
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

	// Server bindings
	if server.Bindings.Len() > 0 {
		srvResult.BindingsStructNoAssemble = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    srvResult.Struct.Name + "Bindings",
				Render:  true,
				Package: ctx.Stack.Top().PackageKind,
			},
		}
		if srvBindings, ok := server.Bindings.Get(protoName); ok {
			var bindings serverBindings
			if err := utils.UnmarhalRawsUnion2(srvBindings, &bindings); err != nil { // TODO: implement $ref
				return nil, err
			}
			marshalFields := []string{"SchemaRegistryURL", "SchemaRegistryVendor"}
			if err := utils.StructToOrderedMap(bindings, &srvResult.BindingsValues, marshalFields); err != nil {
				return nil, err
			}
		}
	}

	return srvResult, nil
}
