package assemble

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

type Message struct {
	Struct           *Struct
	PayloadType      common.GolangType
	PayloadHasSchema bool
	HeadersType      common.GolangType
	HeadersHasSchema bool
}

func (m Message) AllowRender() bool {
	return true
}

func (m Message) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement

	if m.PayloadHasSchema {
		res = append(res, m.PayloadType.AssembleDefinition(ctx)...)
	}
	if m.HeadersHasSchema {
		res = append(res, m.HeadersType.AssembleDefinition(ctx)...)
	}

	// NewMessage constructor function
	res = append(res, jen.Func().Id(fmt.Sprintf("New%s", m.Struct.Name)).Params().Op("*").Id(m.Struct.TypeName()).Block(
		jen.Return(jen.Op("&").Id(m.Struct.TypeName()).Values()),
	))

	res = append(res, m.Struct.AssembleDefinition(ctx)...)
	res = append(res, m.messageMethods(ctx)...)
	return res
}

func (m Message) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	return m.Struct.AssembleUsage(ctx)
}

func (m Message) messageMethods(ctx *common.AssembleContext) []*jen.Statement {
	structName := m.Struct.Name
	receiverName := strings.ToLower(string(structName[0]))
	receiver := jen.Id(receiverName).Op("*").Id(structName)
	payloadFieldType := utils.ToCode(m.PayloadType.AssembleUsage(ctx))
	headersFieldType := utils.ToCode(m.HeadersType.AssembleUsage(ctx))

	return []*jen.Statement{
		jen.Func().Params(receiver.Clone()).Id("MarshalBinary").
			Params().
			Params(jen.Index().Byte(), jen.Error()).
			Block(
				jen.Return(jen.Qual("encoding/json", "Marshal").Call(jen.Id(receiverName).Dot("Payload"))),
			),
		jen.Func().Params(receiver.Clone()).Id("UnmarshalBinary").
			Params(jen.Id("data").Index().Byte()).
			Params(jen.Error()).
			Block(
				jen.Return(jen.Qual("encoding/json", "Unmarshal").Call(jen.Id("data"), jen.Op("&").Id(receiverName).Dot("Payload"))),
			),
		jen.Func().Params(receiver.Clone()).Id("WithID").
			Params(jen.Id("id").String()).
			Params(jen.Op("*").Id(structName)).
			Block(
				jen.Id(receiverName).Dot("ID").Op("=").Id("id"),
				jen.Return(jen.Id(receiverName)),
			),
		jen.Func().Params(receiver.Clone()).Id("WithPayload").
			Params(jen.Id("payload").Add(payloadFieldType...)).
			Params(jen.Op("*").Id(structName)).
			Block(
				jen.Id(receiverName).Dot("Payload").Op("=").Id("payload"),
				jen.Return(jen.Id(receiverName)),
			),
		jen.Func().Params(receiver.Clone()).Id("WithHeaders").
			Params(jen.Id("headers").Add(headersFieldType...)).
			Params(jen.Op("*").Id(structName)).
			Block(
				jen.Id(receiverName).Dot("Headers").Op("=").Id("headers"),
				jen.Return(jen.Id(receiverName)),
			),
		jen.Func().Params(receiver.Clone()).Id("WithParameters").
			Params(jen.Id("parameters").Op("...").Qual(ctx.RuntimePackage(""), "Parameter")).
			Params(jen.Op("*").Id(structName)).
			Block(
				utils.QualSprintf(`
					if %[1]s.ChannelParameters == nil {
						%[1]s.ChannelParameters = make(map[string]%Q(%[2]s,Parameter))
					}
					for _, p := range parameters {
						%[1]s.ChannelParameters[p.Name()] = p
					}
					return %[1]s`, receiverName, ctx.RuntimePackage("")),
			),
	}
}
