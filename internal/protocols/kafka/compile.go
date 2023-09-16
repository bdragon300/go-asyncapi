package kafka

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/samber/lo"
)

const protoName = "kafka"

func Register() {
	compile.ProtoServerBuilders[protoName] = BuildServer
	compile.ProtoChannelBuilders[protoName] = BuildChannel
	compile.ProtoServerBuilders["kafka-secure"] = BuildServer // TODO: make a separate kafka-secure protocol
	compile.ProtoChannelBuilders["kafka-secure"] = BuildChannel
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
	}

	return chanResult, nil
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

	return srvResult, nil
}
