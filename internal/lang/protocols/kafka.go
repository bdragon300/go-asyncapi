package protocols

import (
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type KafkaChannel struct {
	Name                string
	Topic               string
	PubStruct           *lang.Struct
	PubMessage          render.LangRenderer
	PubMessageHasSchema bool
	SubStruct           *lang.Struct
	SubMessage          render.LangRenderer
	SubMessageHasSchema bool
}

func (k KafkaChannel) AllowRender() bool {
	return true
}

func (k KafkaChannel) RenderDefinition(ctx *render.Context) []*j.Statement {
	var res []*j.Statement
	if k.PubStruct != nil {
		res = append(res, k.PubStruct.RenderDefinition(ctx)...)
		res = append(res, k.producerMethods(ctx)...)
	}
	if k.SubStruct != nil {
		res = append(res, k.SubStruct.RenderDefinition(ctx)...)
		res = append(res, k.consumerMethods(ctx)...)
	}
	return res
}

func (k KafkaChannel) RenderUsage(_ *render.Context) []*j.Statement {
	panic("not implemented")
}

func (k KafkaChannel) commonMethods(ctx *render.Context, strct *lang.Struct, msg render.LangRenderer, clientField string) []*j.Statement {
	structName := strct.Name
	messageType := utils.CastSliceItems[*j.Statement, j.Code](msg.RenderUsage(ctx))
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	return []*j.Statement{
		// Method Name() -> string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().Block(
			j.Return(j.Lit(k.Name)),
		),
		// Method Topic() -> string
		j.Func().Params(receiver.Clone()).Id("Topic").
			Params().
			String().Block(
			j.Return(j.Lit(k.Topic)),
		),
		// Method Envelope(message MessageType, key []byte) -> *KafkaOutEnvelope
		j.Func().Params(receiver.Clone()).Id("Envelope").
			Params(j.Id("message").Add(messageType...), j.Id("key").Index().Byte()).
			Params(j.Op("*").Id("KafkaOutEnvelope"), j.Error()).Block(
			j.List(j.Id("payload"), j.Err()).Op(":=").Qual("encoding/json", "Marshal").Call(j.Id("message")),
			j.If(j.Err().Op("!=").Nil()).Block(
				j.Return(j.List(j.Nil(), j.Err())),
			),
			j.Return(j.List(j.Op("&").Id("KafkaOutEnvelope").Values(j.Dict{
				j.Id("KafkaMeta"): j.Id("KafkaMeta").Values(j.Dict{
					j.Id("Key"):       j.Id("key"),
					j.Id("Topic"):     j.Id(receiverName).Dot("Topic").Call(),
					j.Id("Partition"): j.Nil(),
				}),
				j.Id("Payload"): j.Id("payload"),
				j.Id("Headers"): j.Qual(ctx.RuntimePackage, "StructToMapByte").Call(j.Id("message").Dot("Headers")),
				j.Id("To"):      j.Id(receiverName).Dot(clientField),
			}), j.Nil())),
		),
	}
}

func (k KafkaChannel) producerMethods(ctx *render.Context) []*j.Statement {
	structName := k.PubStruct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	var res []*j.Statement
	res = append(res, k.commonMethods(ctx, k.PubStruct, k.PubMessage, "producers")...)
	publishMethod := j.Func().Params(receiver.Clone()).Id("Publish").
		Params(j.Id("ctx").Qual("context", "Context"), j.Id("envelope").Op("*").Id("KafkaOutEnvelope")).
		Error().Block(
		j.Id("p").Op(":=").Qual(ctx.RuntimePackage, "ErrorPool").Call(j.Len(j.Id("envelope").Dot("To"))),
		j.Op(`
for i := 0; i < len(envelope.To); i++ {
	i := i
	p.Go(func() error {
		return envelope.To[i].Produce(ctx, envelope)
	})
}
return p.Wait(ctx)
`),
	)

	return append(res, publishMethod)
}

func (k KafkaChannel) consumerMethods(ctx *render.Context) []*j.Statement {
	structName := k.SubStruct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := j.Id(receiverName).Op("*").Id(structName)

	var res []*j.Statement
	res = append(res, k.commonMethods(ctx, k.SubStruct, k.SubMessage, "consumer")...)
	publishMethod := j.Func().Params(receiver.Clone()).Id("Subscribe").
		Params(j.Id("ctx").Qual("context", "Context"), j.Id("cb").Id("KafkaConsumerCallback")).
		Error().Block(
		j.Return(j.Id(receiverName).Dot("consumer").Dot("Consume").Call(j.Id("ctx"), j.Id("cb"))),
	)

	return append(res, publishMethod)
}
