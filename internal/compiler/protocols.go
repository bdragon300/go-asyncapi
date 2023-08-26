package compiler

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/lang/protocols"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type protocolBuilderFunc func(ctx *scan.Context, channelName string) (render.LangRenderer, error)

func (c Channel) buildKafka(ctx *scan.Context, channelName string) (render.LangRenderer, error) {
	res := protocols.KafkaChannel{
		Name:  channelName,
		Topic: channelName,
	} // TODO: kafka bindings

	if c.Publish != nil {
		res.PubStruct = &lang.Struct{
			BaseType: lang.BaseType{
				Name:        getTypeName(ctx, channelName, "KafkaPubChannel"),
				Description: utils.JoinNonemptyStrings("\n", c.Description, c.Publish.Description),
				Render:      true,
			},
			Fields: []lang.StructField{{
				Name: "producer",
				Type: &lang.Simple{TypeName: "KafkaProducer", Package: common.RuntimePackageKind},
			}},
		}
		res.PubMessage, res.PubMessageHasSchema = c.getMessageType(ctx, c.Publish, "publish")
	}
	if c.Subscribe != nil {
		res.SubStruct = &lang.Struct{
			BaseType: lang.BaseType{
				Name:        getTypeName(ctx, channelName, "KafkaSubChannel"),
				Description: utils.JoinNonemptyStrings("\n", c.Description, c.Subscribe.Description),
				Render:      true,
			},
			Fields: []lang.StructField{{
				Name: "consumer",
				Type: &lang.Simple{TypeName: "KafkaConsumer", Package: common.RuntimePackageKind},
			}},
		}
		res.SubMessage, res.SubMessageHasSchema = c.getMessageType(ctx, c.Subscribe, "subscribe")
	}

	return res, nil
}

func (c Channel) getMessageType(ctx *scan.Context, operation *Operation, operationField string) (lang.LangType, bool) {
	if operation.Message != nil && operation.Message.Payload != nil {
		path := append(ctx.PathStack(), operationField, "message", "payload")
		q := lang.NewLinkerQueryTypePath(ctx.Top().PackageKind, path)
		ctx.Linker.Add(q)
		return q, true
	}
	return &lang.Simple{TypeName: "any"}, false
}
