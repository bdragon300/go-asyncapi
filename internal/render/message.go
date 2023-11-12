package render

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type protoMessageRendererFunc func(ctx *common.RenderContext, message *Message) []*j.Statement

var (
	ProtoMessageUnmarshalEnvelopeMethodRenderer = map[string]protoMessageRendererFunc{}
	ProtoMessageMarshalEnvelopeMethodRenderer   = map[string]protoMessageRendererFunc{}
)

type Message struct {
	Name                       string
	OutStruct                  *Struct
	InStruct                   *Struct
	PayloadType                common.GolangType
	PayloadHasSchema           bool
	HeadersFallbackType        *Map
	HeadersTypeLink            *Link[*Struct]
	AllServers                 *LinkList[*Server] // For extracting all using protocols
	BindingsStruct             *Struct            // nil if message bindings are not defined
	BindingsStructProtoMethods utils.OrderedMap[string, common.Renderer]
	ContentType                string // Message's content type or default from schema or fallback
}

func (m Message) DirectRendering() bool {
	return true
}

func (m Message) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	ctx.LogRender("Message", "", m.Name, "definition", m.DirectRendering())
	defer ctx.LogReturn()

	if m.BindingsStruct != nil {
		res = append(res, m.BindingsStruct.RenderDefinition(ctx)...)
		for _, e := range m.BindingsStructProtoMethods.Entries() {
			res = append(res, e.Value.RenderDefinition(ctx)...)
		}
	}

	if m.PayloadHasSchema {
		res = append(res, m.PayloadType.RenderDefinition(ctx)...)
	}
	if m.HeadersTypeLink != nil {
		res = append(res, m.HeadersFallbackType.RenderDefinition(ctx)...)
	}

	res = append(res, m.renderPublishMessageStruct(ctx)...)
	res = append(res, m.renderSubscribeMessageStruct(ctx)...)

	return res
}

func (m Message) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (m Message) String() string {
	return "Message " + m.Name
}

func (m Message) renderPublishMessageStruct(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	res = append(res, j.Func().Id(m.OutStruct.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.OutStruct.RenderUsage(ctx))...).Block(
		j.Return(j.Op("&").Add(utils.ToCode(m.OutStruct.RenderUsage(ctx))...).Values()),
	))
	res = append(res, m.OutStruct.RenderDefinition(ctx)...)

	for _, srv := range m.AllServers.Targets() {
		res = append(res, ProtoMessageMarshalEnvelopeMethodRenderer[srv.Protocol](ctx, &m)...)
	}
	res = append(res, m.renderPublishCommonMethods(ctx)...)

	return res
}

func (m Message) renderPublishCommonMethods(ctx *common.RenderContext) []*j.Statement {
	structName := m.OutStruct.Name
	rn := m.OutStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(structName)
	payloadFieldType := utils.ToCode(m.PayloadType.RenderUsage(ctx))
	headersFieldType := utils.ToCode(m.HeadersFallbackType.RenderUsage(ctx))
	if m.HeadersTypeLink != nil {
		headersFieldType = utils.ToCode(m.HeadersTypeLink.Target().RenderUsage(ctx))
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

func (m Message) renderSubscribeMessageStruct(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	res = append(res, j.Func().Id(m.InStruct.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.InStruct.RenderUsage(ctx))...).Block(
		j.Return(j.Op("&").Add(utils.ToCode(m.InStruct.RenderUsage(ctx))...).Values()),
	))
	res = append(res, m.InStruct.RenderDefinition(ctx)...)

	for _, srv := range m.AllServers.Targets() {
		res = append(res, ProtoMessageUnmarshalEnvelopeMethodRenderer[srv.Protocol](ctx, &m)...)
	}
	res = append(res, m.renderSubscribeCommonMethods(ctx)...)

	return res
}

func (m Message) renderSubscribeCommonMethods(ctx *common.RenderContext) []*j.Statement {
	structName := m.InStruct.Name
	rn := m.InStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(structName)
	payloadFieldType := utils.ToCode(m.PayloadType.RenderUsage(ctx))
	headersFieldType := utils.ToCode(m.HeadersFallbackType.RenderUsage(ctx))
	if m.HeadersTypeLink != nil {
		headersFieldType = utils.ToCode(m.HeadersTypeLink.Target().RenderUsage(ctx))
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
