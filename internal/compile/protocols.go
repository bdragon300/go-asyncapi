package compile

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/assemble/protocols"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type protocolBuilderFunc func(ctx *common.Context, channelName string) (common.Assembled, error)

func (c Channel) buildKafka(ctx *common.Context, channelName string) (common.Assembled, error) {
	res := protocols.KafkaChannel{
		Name:  channelName,
		Topic: channelName,
	} // TODO: kafka bindings

	if c.Publish != nil {
		res.PubStruct = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        getTypeName(ctx, channelName, "KafkaPubChannel"),
				Description: utils.JoinNonemptyStrings("\n", c.Description, c.Publish.Description),
				Render:      true,
			},
			Fields: []assemble.StructField{{
				Name: "producer",
				Type: &assemble.Simple{Name: "KafkaProducer", Package: common.RuntimePackageKind},
			}},
		}
		res.PubMessage, res.PubMessageHasSchema = c.getMessageType(ctx, c.Publish, "publish")
	}
	if c.Subscribe != nil {
		res.SubStruct = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        getTypeName(ctx, channelName, "KafkaSubChannel"),
				Description: utils.JoinNonemptyStrings("\n", c.Description, c.Subscribe.Description),
				Render:      true,
			},
			Fields: []assemble.StructField{{
				Name: "consumer",
				Type: &assemble.Simple{Name: "KafkaConsumer", Package: common.RuntimePackageKind},
			}},
		}
		res.SubMessage, res.SubMessageHasSchema = c.getMessageType(ctx, c.Subscribe, "subscribe")
	}

	return res, nil
}

func (c Channel) getMessageType(ctx *common.Context, operation *Operation, operationField string) (common.GolangType, bool) {
	if operation.Message != nil && operation.Message.Payload != nil {
		path := append(ctx.PathStack(), operationField, "message", "payload")
		q := assemble.NewLinkQueryTypePath(ctx.Top().PackageKind, path)
		ctx.Linker.Add(q)
		return q, true
	}
	return &assemble.Simple{Name: "any"}, false
}
