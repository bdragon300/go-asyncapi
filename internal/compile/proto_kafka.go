package compile

import (
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
				},
				Fields: []assemble.StructField{{
					Name: "producer",
					Type: &assemble.Simple{Name: "KafkaProducer", Package: common.RuntimePackageKind},
				}},
			},
		}
		ch.Message, ch.MessageHasSchema = c.getMessageType(ctx, c.Publish, "publish")
		res.Publish = kafka.ProtoChannelPub{ProtoChannel: ch}
		commonStruct.Fields = append(commonStruct.Fields, assemble.StructField{
			Name: "",
			Type: assemble.Simple{Name: getTypeName(ctx, name, "KafkaPubChannel")},
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
				},
				Fields: []assemble.StructField{{
					Name: "consumer",
					Type: &assemble.Simple{Name: "KafkaConsumer", Package: common.RuntimePackageKind},
				}},
			},
		}
		ch.Message, ch.MessageHasSchema = c.getMessageType(ctx, c.Subscribe, "subscribe")
		res.Subscribe = kafka.ProtoChannelSub{ProtoChannel: ch}
		commonStruct.Fields = append(commonStruct.Fields, assemble.StructField{
			Name: "",
			Type: assemble.Simple{Name: getTypeName(ctx, name, "KafkaSubChannel")},
		})
	}

	return res, nil
}

func (c Channel) getMessageType(ctx *common.CompileContext, operation *Operation, operationField string) (common.GolangType, bool) {
	if operation.Message != nil && operation.Message.Payload != nil {
		path := append(ctx.PathStack(), operationField, "message", "payload")
		q := assemble.NewLinkQueryTypePath(ctx.Top().PackageKind, path)
		ctx.Linker.Add(q)
		return q, true
	}
	return &assemble.Simple{Name: "any"}, false
}

func (s Server) buildKafka(ctx *common.CompileContext, name string) (assemble.ServerParts, error) {
	res := assemble.ServerParts{}

	channelsLnk := assemble.NewLinkCbList[*assemble.Channel](common.ChannelsPackageKind, func(item any, path []string) bool {
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
			},
			Fields: []assemble.StructField{{
				Type: assemble.Simple{Name: getTypeName(ctx, name, "PubServer")},
			}, {
				Type: assemble.Simple{Name: getTypeName(ctx, name, "SubServer")},
			}},
		},
		ChannelsLinks: channelsLnk,
	}
	res.Common = kafka.ProtoServerCommon{ProtoServer: srv}

	return res, nil
}
