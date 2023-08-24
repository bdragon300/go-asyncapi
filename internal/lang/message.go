package lang

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

type Message struct {
	Name             string
	Struct           *Struct
	PayloadType      LangType
	PayloadHasSchema bool
	HeadersType      LangType
	HeadersHasSchema bool
}

func (m Message) AllowRender() bool {
	return true
}

func (m Message) RenderDefinition(ctx *render.Context) []*jen.Statement {
	var res []*jen.Statement

	if m.PayloadHasSchema {
		res = append(res, m.PayloadType.RenderDefinition(ctx)...)
	}
	if m.HeadersHasSchema {
		res = append(res, m.HeadersType.RenderDefinition(ctx)...)
	}

	// NewMessage constructor function
	res = append(res, jen.Func().Id(fmt.Sprintf("New%s", m.Struct.Name)).Params().Op("*").Id(m.Struct.GetName()).Block(
		jen.Return(jen.Op("&").Id(m.Struct.GetName()).Values()),
	))

	res = append(res, m.Struct.RenderDefinition(ctx)...)
	res = append(res, messageMethods(ctx, &m)...)
	return res
}

func (m Message) RenderUsage(ctx *render.Context) []*jen.Statement {
	return m.Struct.RenderUsage(ctx)
}

func (m Message) AdditionalImports() map[string]string {
	return nil
}

func messageMethods(ctx *render.Context, msg *Message) []*jen.Statement {
	structName := msg.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := jen.Id(receiverName).Op("*").Id(structName)
	payloadFieldType := utils.CastSliceItems[*jen.Statement, jen.Code](msg.Struct.MustFindField("Payload").Type.RenderUsage(ctx))
	headersFieldType := utils.CastSliceItems[*jen.Statement, jen.Code](msg.Struct.MustFindField("Headers").Type.RenderUsage(ctx))

	return []*jen.Statement{
		jen.Func().Params(receiver.Clone()).Id("MarshalBinary").Params().Params(jen.Index().Byte(), jen.Error()).Block(
			jen.Return(jen.Qual("encoding/json", "Marshal").Call(jen.Id(receiverName).Dot("Payload"))),
		),
		jen.Func().Params(receiver.Clone()).Id("UnmarshalBinary").Params(jen.Id("data").Index().Byte()).Params(jen.Error()).Block(
			jen.Return(jen.Qual("encoding/json", "Unmarshal").Call(jen.Id("data"), jen.Op("&").Id(receiverName).Dot("Payload"))),
		),
		jen.Func().Params(receiver.Clone()).Id("WithID").Params(jen.Id("ID").String()).Params(jen.Op("*").Id(structName)).Block(
			jen.Id(receiverName).Dot("ID").Op("=").Id("ID"),
			jen.Return(jen.Id(receiverName)),
		),
		jen.Func().Params(receiver.Clone()).Id("WithPayload").Params(jen.Id("payload").Add(payloadFieldType...)).Params(jen.Op("*").Id(structName)).Block(
			jen.Id(receiverName).Dot("Payload").Op("=").Id("payload"),
			jen.Return(jen.Id(receiverName)),
		),
		jen.Func().Params(receiver.Clone()).Id("WithHeaders").Params(jen.Id("headers").Add(headersFieldType...)).Params(jen.Op("*").Id(structName)).Block(
			jen.Id(receiverName).Dot("Headers").Op("=").Id("headers"),
			jen.Return(jen.Id(receiverName)),
		),
	}
}
