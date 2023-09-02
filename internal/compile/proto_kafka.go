package compile

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/assemble/kafka"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/samber/lo"
)

func (c Channel) buildKafka(ctx *common.CompileContext, name string) (assemble.ChannelParts, error) {
	commonStruct := assemble.Struct{
		BaseType: assemble.BaseType{
			Name:        getTypeName(ctx, name, "KafkaChannel"),
			Description: c.Description,
			Render:      true,
			Package:     ctx.Top().PackageKind,
		},
	}
	res := assemble.ChannelParts{Common: &commonStruct}

	if c.Publish != nil {
		ch := kafka.ProtoChannel{
			Name:  name,
			Topic: name,
			Struct: &assemble.Struct{
				BaseType: assemble.BaseType{
					Name:        getTypeName(ctx, name, "KafkaPubChannel"),
					Description: utils.JoinNonemptyStrings("\n", c.Description, c.Publish.Description),
					Render:      true,
					Package:     ctx.Top().PackageKind,
				},
				Fields: []assemble.StructField{{
					Name: "Producer",
					Type: &assemble.Simple{Name: "KafkaProducer", Package: common.RuntimePackageKind},
				}},
			},
		}
		ch.Message, ch.MessageHasSchema = c.getMessageType(ctx, c.Publish, "publish")
		res.Publish = kafka.ProtoChannelPub{ProtoChannel: ch}
		commonStruct.Fields = append(commonStruct.Fields, assemble.StructField{
			Name: "",
			Type: assemble.Simple{Name: getTypeName(ctx, name, "KafkaPubChannel"), Package: ctx.Top().PackageKind},
		})
	}
	if c.Subscribe != nil {
		ch := kafka.ProtoChannel{
			Name:  name,
			Topic: name,
			Struct: &assemble.Struct{
				BaseType: assemble.BaseType{
					Name:        getTypeName(ctx, name, "KafkaSubChannel"),
					Description: utils.JoinNonemptyStrings("\n", c.Description, c.Subscribe.Description),
					Render:      true,
					Package:     ctx.Top().PackageKind,
				},
				Fields: []assemble.StructField{{
					Name: "Consumer",
					Type: &assemble.Simple{Name: "KafkaConsumer", Package: common.RuntimePackageKind},
				}},
			},
		}
		ch.Message, ch.MessageHasSchema = c.getMessageType(ctx, c.Subscribe, "subscribe")
		res.Subscribe = kafka.ProtoChannelSub{ProtoChannel: ch}
		commonStruct.Fields = append(commonStruct.Fields, assemble.StructField{
			Name: "",
			Type: assemble.Simple{Name: getTypeName(ctx, name, "KafkaSubChannel"), Package: ctx.Top().PackageKind},
		})
	}

	return res, nil
}

func (c Channel) getMessageType(ctx *common.CompileContext, operation *Operation, operationField string) (common.GolangType, bool) {
	if operation.Message != nil && operation.Message.Payload != nil {
		ref := fmt.Sprintf("#/%s/%s/message/payload", strings.Join(ctx.PathStack(), "/"), operationField)
		lnk := assemble.NewRefLinkAsGolangType(ctx.Top().PackageKind, ref)
		ctx.Linker.Add(lnk)
		return lnk, true
	}
	return &assemble.Simple{Name: "any"}, false
}

func (s Server) buildKafka(ctx *common.CompileContext, name string) (assemble.ServerParts, error) {
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

	srv := kafka.ProtoServer{
		Name: name,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        getTypeName(ctx, name, "PubServer"),
				Description: s.Description,
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
			Fields: []assemble.StructField{{
				Name: "Producer",
				Type: assemble.Simple{
					Name:    "KafkaProducer",
					Package: common.RuntimePackageKind,
				},
			}},
		},
		ChannelsLinks: channelsLnk,
	}
	res.Publish = kafka.ProtoServerPub{ProtoServer: srv}

	srv = kafka.ProtoServer{
		Name: name,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        getTypeName(ctx, name, "SubServer"),
				Description: s.Description,
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
			Fields: []assemble.StructField{{
				Name: "Consumer",
				Type: assemble.Simple{
					Name:    "KafkaConsumer",
					Package: common.RuntimePackageKind,
				},
			}},
		},
		ChannelsLinks: channelsLnk,
	}
	res.Subscribe = kafka.ProtoServerSub{ProtoServer: srv}

	srv = kafka.ProtoServer{
		Name: name,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        getTypeName(ctx, name, "Server"),
				Description: s.Description,
				Render:      true,
				Package:     ctx.Top().PackageKind,
			},
			Fields: []assemble.StructField{{
				Type: assemble.Simple{Name: getTypeName(ctx, name, "PubServer"), Package: ctx.Top().PackageKind},
			}, {
				Type: assemble.Simple{Name: getTypeName(ctx, name, "SubServer"), Package: ctx.Top().PackageKind},
			}},
		},
		ChannelsLinks: channelsLnk,
	}
	res.Common = kafka.ProtoServerCommon{ProtoServer: srv}

	return res, nil
}
