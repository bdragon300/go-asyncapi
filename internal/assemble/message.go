package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type protoMessageAssemblerFunc func(ctx *common.AssembleContext, message *Message) []*j.Statement

var (
	ProtoMessageUnmarshalEnvelopeMethodAssembler = map[string]protoMessageAssemblerFunc{}
	ProtoMessageMarshalEnvelopeMethodAssembler   = map[string]protoMessageAssemblerFunc{}
)

type Message struct {
	OutStruct                  *Struct
	InStruct                   *Struct
	PayloadType                common.GolangType
	PayloadHasSchema           bool
	HeadersFallbackType        *Map
	HeadersTypeLink            *Link[*Struct]
	AllServers                 *LinkList[*Server] // For extracting all using protocols
	BindingsStruct             *Struct            // nil if message bindings are not defined
	BindingsStructProtoMethods []common.Assembler
}

func (m Message) AllowRender() bool {
	return true
}

func (m Message) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement

	if m.BindingsStruct != nil {
		res = append(res, m.BindingsStruct.AssembleDefinition(ctx)...)
		for _, mtd := range m.BindingsStructProtoMethods {
			res = append(res, mtd.AssembleDefinition(ctx)...)
		}
	}

	if m.PayloadHasSchema {
		res = append(res, m.PayloadType.AssembleDefinition(ctx)...)
	}
	if m.HeadersTypeLink != nil {
		res = append(res, m.HeadersFallbackType.AssembleDefinition(ctx)...)
	}

	res = append(res, m.assemblePublishMessageStruct(ctx)...)
	res = append(res, m.assembleSubscribeMessageStruct(ctx)...)

	return res
}

func (m Message) AssembleUsage(_ *common.AssembleContext) []*j.Statement {
	panic("not implemented")
}

func (m Message) assemblePublishMessageStruct(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	res = append(res, j.Func().Id(m.OutStruct.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.OutStruct.AssembleUsage(ctx))...).Block(
		j.Return(j.Op("&").Add(utils.ToCode(m.OutStruct.AssembleUsage(ctx))...).Values()),
	))
	res = append(res, m.OutStruct.AssembleDefinition(ctx)...)

	for _, srv := range m.AllServers.Targets() {
		res = append(res, ProtoMessageMarshalEnvelopeMethodAssembler[srv.Protocol](ctx, &m)...)
	}
	res = append(res, m.assemblePublishCommonMethods(ctx)...)

	return res
}

func (m Message) assemblePublishCommonMethods(ctx *common.AssembleContext) []*j.Statement {
	structName := m.OutStruct.Name
	rn := m.OutStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(structName)
	payloadFieldType := utils.ToCode(m.PayloadType.AssembleUsage(ctx))
	headersFieldType := utils.ToCode(m.HeadersFallbackType.AssembleUsage(ctx))
	if m.HeadersTypeLink != nil {
		headersFieldType = utils.ToCode(m.HeadersTypeLink.Target().AssembleUsage(ctx))
	}

	return []*j.Statement{
		// Method WithID(ID string) *Message2Out
		j.Func().Params(receiver.Clone()).Id("WithID").
			Params(j.Id("id").String()).
			Params(j.Op("*").Id(structName)).
			Block(
				j.Op(fmt.Sprintf(`
					%[1]s.ID = id
					return %[1]s`, rn)),
			),
		// Method WithPayload(payload Model2) *Message2Out
		j.Func().Params(receiver.Clone()).Id("WithPayload").
			Params(j.Id("payload").Add(payloadFieldType...)).
			Params(j.Op("*").Id(structName)).
			Block(
				j.Op(fmt.Sprintf(`
					%[1]s.Payload = payload
					return %[1]s`, rn)),
			),
		// Method WithHeaders(headers Message2Headers) *Message2Out
		j.Func().Params(receiver.Clone()).Id("WithHeaders").
			Params(j.Id("headers").Add(headersFieldType...)).
			Params(j.Op("*").Id(structName)).
			Block(
				j.Op(fmt.Sprintf(`
					%[1]s.Headers = headers
					return %[1]s`, rn)),
			),
	}
}

func (m Message) assembleSubscribeMessageStruct(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	res = append(res, j.Func().Id(m.InStruct.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.InStruct.AssembleUsage(ctx))...).Block(
		j.Return(j.Op("&").Add(utils.ToCode(m.InStruct.AssembleUsage(ctx))...).Values()),
	))
	res = append(res, m.InStruct.AssembleDefinition(ctx)...)

	for _, srv := range m.AllServers.Targets() {
		res = append(res, ProtoMessageUnmarshalEnvelopeMethodAssembler[srv.Protocol](ctx, &m)...)
	}
	res = append(res, m.assembleSubscribeCommonMethods(ctx)...)

	return res
}

func (m Message) assembleSubscribeCommonMethods(ctx *common.AssembleContext) []*j.Statement {
	structName := m.InStruct.Name
	rn := m.InStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(structName)
	payloadFieldType := utils.ToCode(m.PayloadType.AssembleUsage(ctx))
	headersFieldType := utils.ToCode(m.HeadersFallbackType.AssembleUsage(ctx))
	if m.HeadersTypeLink != nil {
		headersFieldType = utils.ToCode(m.HeadersTypeLink.Target().AssembleUsage(ctx))
	}

	return []*j.Statement{
		// Method MessageID() string
		j.Func().Params(receiver.Clone()).Id("MessageID").
			Params().
			String().
			Block(
				j.Return(j.Id(rn).Dot("ID")),
			),
		// Method MessagePayload() Model2
		j.Func().Params(receiver.Clone()).Id("MessagePayload").
			Params().
			Add(payloadFieldType...).
			Block(
				j.Return(j.Id(rn).Dot("Payload")),
			),
		// Method MessageHeaders() Message2Headers
		j.Func().Params(receiver.Clone()).Id("MessageHeaders").
			Params().
			Add(headersFieldType...).
			Block(
				j.Return(j.Id(rn).Dot("Headers")),
			),
	}
}
