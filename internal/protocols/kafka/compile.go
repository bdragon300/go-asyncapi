package kafka

import (
	"fmt"
	"path"

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
	compile.ProtoServerBuilders["kafka-secure"] = BuildServer
	compile.ProtoChannelBuilders["kafka-secure"] = BuildChannel
}

func BuildChannel(ctx *common.CompileContext, channel *compile.Channel, name string) (assemble.ChannelParts, error) {
	commonStruct := ProtoChannelCommon{
		Struct: assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, name, "KafkaChannel"),
				Description: channel.Description,
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
		},
	}
	res := assemble.ChannelParts{Common: &commonStruct}

	if channel.Publish != nil {
		chGolangName := compile.GenerateGolangTypeName(ctx, name, "KafkaPubChannel")
		strct := assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        chGolangName,
				Description: utils.JoinNonemptyStrings("\n", channel.Description, channel.Publish.Description),
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
		}
		iface := &assemble.Interface{
			BaseType: assemble.BaseType{
				Name:    compile.GenerateGolangTypeName(ctx, name, "KafkaPubServer"),
				Render:  true,
				Package: ctx.Top().PackageKind,
			},
			Methods: []assemble.FunctionSignature{
				{
					Name:   compile.GenerateGolangTypeName(ctx, name, "PubChannel"),
					Return: []assemble.FuncParam{{Type: &strct, Pointer: true}},
				},
				{
					Name:   "Producer",
					Return: []assemble.FuncParam{{Type: &assemble.Simple{Name: "Producer", Package: common.RuntimeKafkaPackageKind}}},
				},
			},
		}
		strct.Fields = []assemble.StructField{{
			Name: "servers",
			Type: &assemble.Array{
				BaseType:  assemble.BaseType{Package: ctx.Top().PackageKind},
				ItemsType: iface,
			},
			RequiredValue: false,
		}}
		ch := ProtoChannel{
			Name:        name,
			Topic:       name,
			Struct:      &strct,
			Iface:       iface,
			MessageLink: getOperationMessageType(ctx, channel.Publish, "publish"),
		}
		ch.NewFunc = &assemble.NewFunc{
			BaseType:           assemble.BaseType{Name: "New" + strct.Name, Render: true, Package: ctx.Top().PackageKind},
			Struct:             &strct,
			NewFuncArgs:        []assemble.FuncParam{{Name: "servers", Type: ch.Iface, Variadic: true}},
			NewFuncAllocFields: []string{"servers"},
		}
		res.Publish = &ProtoChannelPub{ProtoChannel: ch}
		commonStruct.Fields = append(commonStruct.Fields, assemble.StructField{
			Type: assemble.Simple{Name: chGolangName, Package: ctx.Top().PackageKind},
		})
	}
	if channel.Subscribe != nil {
		chGolangName := compile.GenerateGolangTypeName(ctx, name, "KafkaSubChannel")
		strct := assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        chGolangName,
				Description: utils.JoinNonemptyStrings("\n", channel.Description, channel.Subscribe.Description),
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
		}
		iface := &assemble.Interface{
			BaseType: assemble.BaseType{
				Name:    compile.GenerateGolangTypeName(ctx, name, "KafkaSubServer"),
				Render:  true,
				Package: ctx.Top().PackageKind,
			},
			Methods: []assemble.FunctionSignature{
				{
					Name:   compile.GenerateGolangTypeName(ctx, name, "SubChannel"),
					Return: []assemble.FuncParam{{Type: &strct, Pointer: true}},
				},
				{
					Name:   "Consumer",
					Return: []assemble.FuncParam{{Type: &assemble.Simple{Name: "Consumer", Package: common.RuntimeKafkaPackageKind}}},
				},
			},
		}
		strct.Fields = []assemble.StructField{{
			Name: "servers",
			Type: &assemble.Array{
				BaseType:  assemble.BaseType{Package: ctx.Top().PackageKind},
				ItemsType: iface,
			},
			RequiredValue: false,
		}}
		ch := ProtoChannel{
			Name:        name,
			Topic:       name,
			Struct:      &strct,
			Iface:       iface,
			MessageLink: getOperationMessageType(ctx, channel.Subscribe, "subscribe"),
		}
		ch.NewFunc = &assemble.NewFunc{
			BaseType:           assemble.BaseType{Name: "New" + strct.Name, Render: true, Package: ctx.Top().PackageKind},
			Struct:             &strct,
			NewFuncArgs:        []assemble.FuncParam{{Name: "servers", Type: ch.Iface, Variadic: true}},
			NewFuncAllocFields: []string{"servers"},
		}
		res.Subscribe = &ProtoChannelSub{ProtoChannel: ch}
		commonStruct.Fields = append(commonStruct.Fields, assemble.StructField{
			Type: assemble.Simple{Name: chGolangName, Package: ctx.Top().PackageKind},
		})
	}

	return res, nil
}

func getOperationMessageType(ctx *common.CompileContext, operation *compile.Operation, operationField string) *assemble.Link[*assemble.Message] {
	if operation.Message == nil {
		return nil
	}

	ref := fmt.Sprintf("#/%s/%s/message", path.Join(ctx.PathStack()...), operationField)
	lnk := assemble.NewRefLink[*assemble.Message](ctx.Top().PackageKind, ref)
	if operation.Message.Ref != "" {
		lnk = assemble.NewRefLink[*assemble.Message](common.MessagesPackageKind, operation.Message.Ref)
	}
	ctx.Linker.Add(lnk)
	return lnk
}

func BuildServer(ctx *common.CompileContext, server *compile.Server, name string) (assemble.ServerParts, error) {
	res := assemble.ServerParts{}

	channelsLnk := assemble.NewListCbLink[*assemble.Channel](common.ChannelsPackageKind, func(item any, path []string) bool {
		ch, ok := item.(*assemble.Channel)
		if !ok {
			return false
		}
		if len(ch.AppliedServers) > 0 {
			return lo.Contains(ch.AppliedServers, name)
		}
		return ch.AppliedToAllServersLinks != nil
	})
	ctx.Linker.AddMany(channelsLnk)

	pub := ProtoServer{
		Name: name,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, name, "PubServer"),
				Description: server.Description,
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
			Fields: []assemble.StructField{{
				Name: "producer",
				Type: assemble.Simple{Name: "Producer", Package: common.RuntimeKafkaPackageKind},
			}},
		},
		ChannelsLinks: channelsLnk,
	}
	pub.NewFunc = &assemble.NewFunc{
		BaseType: assemble.BaseType{Name: "New" + pub.Struct.Name, Render: true, Package: ctx.Top().PackageKind},
		Struct:   pub.Struct,
		NewFuncArgs: []assemble.FuncParam{{
			Name: "producer",
			Type: &assemble.Simple{Name: "Producer", Package: common.RuntimeKafkaPackageKind},
		}},
		NewFuncAllocFields: []string{"producer"},
	}
	res.Publish = ProtoServerPub{ProtoServer: pub}

	sub := ProtoServer{
		Name: name,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, name, "SubServer"),
				Description: server.Description,
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
			Fields: []assemble.StructField{{
				Name: "consumer",
				Type: assemble.Simple{Name: "Consumer", Package: common.RuntimeKafkaPackageKind},
			}},
		},
		ChannelsLinks: channelsLnk,
	}
	sub.NewFunc = &assemble.NewFunc{
		BaseType: assemble.BaseType{Name: "New" + sub.Struct.Name, Render: true, Package: ctx.Top().PackageKind},
		Struct:   sub.Struct,
		NewFuncArgs: []assemble.FuncParam{{
			Name: "consumer",
			Type: &assemble.Simple{Name: "Consumer", Package: common.RuntimeKafkaPackageKind},
		}},
		NewFuncAllocFields: []string{"consumer"},
	}
	res.Subscribe = ProtoServerSub{ProtoServer: sub}

	srv := ProtoServer{
		Name: name,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        compile.GenerateGolangTypeName(ctx, name, "Server"),
				Description: server.Description,
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
			Fields: []assemble.StructField{{Type: sub.Struct}, {Type: pub.Struct}},
		},
		ChannelsLinks: channelsLnk,
	}
	srv.NewFunc = &assemble.NewFunc{
		BaseType: assemble.BaseType{Name: "New" + srv.Struct.Name, Render: true, Package: ctx.Top().PackageKind},
		Struct:   srv.Struct,
		NewFuncArgs: []assemble.FuncParam{{
			Name: "producer",
			Type: &assemble.Simple{Name: "Producer", Package: common.RuntimeKafkaPackageKind},
		}, {
			Name: "consumer",
			Type: &assemble.Simple{Name: "Consumer", Package: common.RuntimeKafkaPackageKind},
		}},
		NewFuncAllocFields: []string{pub.Struct.Name, sub.Struct.Name},
	}
	res.Common = ProtoServerCommon{ProtoServer: srv, PubStruct: pub.Struct, SubStruct: sub.Struct}

	return res, nil
}
