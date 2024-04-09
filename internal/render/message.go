package render

import (
	"fmt"
	"sort"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Message struct {
	Name                 string
	Dummy                bool
	OutStruct            *GoStruct
	InStruct             *GoStruct
	PayloadType          common.GolangType // `any` or a particular type
	PayloadHasSchema     bool
	HeadersFallbackType  *GoMap
	HeadersTypePromise   *Promise[*GoStruct]
	AllServersPromises   []*Promise[*Server]      // For extracting all using protocols
	BindingsStruct       *GoStruct                // nil if message bindings are not defined for message
	BindingsPromise      *Promise[*Bindings]      // nil if message bindings are not defined for message as well
	ContentType          string                   // Message's content type or default from schema or fallback
	CorrelationIDPromise *Promise[*CorrelationID] // nil if correlationID is not defined for message
}

func (m Message) DirectRendering() bool {
	return !m.Dummy
}

func (m Message) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	ctx.LogStartRender("Message", "", m.Name, "definition", m.DirectRendering())
	defer ctx.LogFinishRender()

	// Bindings struct and its methods according to protocols of channels where the message is used
	if m.BindingsStruct != nil {
		res = append(res, m.BindingsStruct.RenderDefinition(ctx)...)

		if m.BindingsPromise != nil {
			tgt := m.BindingsPromise.Target()
			protocols := m.getServerProtocols(ctx)
			ctx.Logger.Debug("Message protocols", "protocols", protocols)
			for _, p := range protocols {
				protoTitle := ctx.ProtoRenderers[p].ProtocolTitle()
				res = append(res, tgt.RenderBindingsMethod(ctx, m.BindingsStruct, p, protoTitle)...)
			}
		}
	}

	if m.PayloadHasSchema {
		res = append(res, m.PayloadType.RenderDefinition(ctx)...)
	}

	res = append(res, m.renderPublishMessageStruct(ctx)...)
	res = append(res, m.renderSubscribeMessageStruct(ctx)...)

	return res
}

func (m Message) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (m Message) ID() string {
	return m.Name
}

func (m Message) String() string {
	return "Message " + m.Name
}

func (m Message) HasProtoBindings(protoName string) bool {
	if m.BindingsPromise == nil {
		return false
	}
	_, ok1 := m.BindingsPromise.Target().Values.Get(protoName)
	_, ok2 := m.BindingsPromise.Target().JSONValues.Get(protoName)
	return ok1 || ok2
}

func (m Message) renderPublishMessageStruct(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderPublishMessageStruct")

	var res []*j.Statement
	res = append(res, j.Func().Id(m.OutStruct.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.OutStruct.RenderUsage(ctx))...).Block(
		j.Return(j.Op("&").Add(utils.ToCode(m.OutStruct.RenderUsage(ctx))...).Values()),
	))
	res = append(res, m.OutStruct.RenderDefinition(ctx)...)

	for _, p := range m.getServerProtocols(ctx) {
		res = append(res, m.renderMarshalEnvelopeMethod(ctx, p, ctx.ProtoRenderers[p].ProtocolTitle())...)
	}
	res = append(res, m.renderPublishCommonMethods(ctx)...)

	return res
}

func (m Message) renderPublishCommonMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderPublishCommonMethods")

	structName := m.OutStruct.Name
	rn := m.OutStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(structName)
	payloadFieldType := utils.ToCode(m.PayloadType.RenderUsage(ctx))
	headersFieldType := utils.ToCode(m.HeadersFallbackType.RenderUsage(ctx))
	if m.HeadersTypePromise != nil {
		headersFieldType = utils.ToCode(m.HeadersTypePromise.Target().RenderUsage(ctx))
	}

	res := []*j.Statement{
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
	if m.CorrelationIDPromise != nil {
		// Method SetCorrelationID(value any)
		res = append(res, m.CorrelationIDPromise.Target().RenderSetterDefinition(ctx, &m)...)
	}
	return res
}

func (m Message) renderMarshalEnvelopeMethod(ctx *common.RenderContext, protoName, protoTitle string) []*j.Statement {
	ctx.Logger.Trace("renderMarshalEnvelopeMethod")

	rn := m.OutStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(m.OutStruct.Name)

	return []*j.Statement{
		// Method MarshalProtoEnvelope(envelope proto.EnvelopeWriter) error
		j.Func().Params(receiver.Clone()).Id("Marshal" + protoTitle + "Envelope").
			Params(j.Id("envelope").Qual(ctx.RuntimeModule(protoName), "EnvelopeWriter")).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("enc := ").Qual(ctx.GeneratedModule(encodingPackageName), "NewEncoder").Call(
					j.Lit(m.ContentType),
					j.Id("envelope"),
				)
				bg.Op(fmt.Sprintf(`
					if err := enc.Encode(%[1]s.Payload); err != nil {
						return err
					}`, rn))
				bg.Op("envelope.SetContentType").Call(j.Lit(m.ContentType))
				if m.HeadersTypePromise != nil {
					bg.Id("envelope").Dot("SetHeaders").Call(
						j.Qual(ctx.RuntimeModule(""), "Headers").Values(j.DictFunc(func(d j.Dict) {
							for _, f := range m.HeadersTypePromise.Target().Fields {
								d[j.Lit(f.Name)] = j.Id(rn).Dot("Headers").Dot(f.Name)
							}
						})),
					)
				} else {
					bg.Id("envelope.SetHeaders").Call(
						j.Qual(ctx.RuntimeModule(""), "Headers").Call(j.Id(rn).Dot("Headers")),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}

func (m Message) renderSubscribeMessageStruct(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderSubscribeMessageStruct")

	var res []*j.Statement
	res = append(res, j.Func().Id(m.InStruct.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.InStruct.RenderUsage(ctx))...).Block(
		j.Return(j.Op("&").Add(utils.ToCode(m.InStruct.RenderUsage(ctx))...).Values()),
	))
	res = append(res, m.InStruct.RenderDefinition(ctx)...)

	for _, p := range m.getServerProtocols(ctx) {
		res = append(res, m.renderUnmarshalEnvelopeMethod(ctx, p, ctx.ProtoRenderers[p].ProtocolTitle())...)
	}
	res = append(res, m.renderSubscribeCommonMethods(ctx)...)

	return res
}

func (m Message) renderSubscribeCommonMethods(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("renderSubscribeCommonMethods")

	structName := m.InStruct.Name
	rn := m.InStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(structName)
	payloadFieldType := utils.ToCode(m.PayloadType.RenderUsage(ctx))
	headersFieldType := utils.ToCode(m.HeadersFallbackType.RenderUsage(ctx))
	if m.HeadersTypePromise != nil {
		headersFieldType = utils.ToCode(m.HeadersTypePromise.Target().RenderUsage(ctx))
	}

	res := []*j.Statement{
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
	if m.CorrelationIDPromise != nil {
		// Method CorrelationID(value any)
		res = append(res, m.CorrelationIDPromise.Target().RenderGetterDefinition(ctx, &m)...)
	}
	return res
}

func (m Message) renderUnmarshalEnvelopeMethod(ctx *common.RenderContext, protoName, protoTitle string) []*j.Statement {
	ctx.Logger.Trace("renderUnmarshalEnvelopeMethod")

	rn := m.InStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(m.InStruct.Name)

	return []*j.Statement{
		// Method UnmarshalProtoEnvelope(envelope proto.EnvelopeReader) error
		j.Func().Params(receiver.Clone()).Id("Unmarshal" + protoTitle + "Envelope").
			Params(j.Id("envelope").Qual(ctx.RuntimeModule(protoName), "EnvelopeReader")).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("dec := ").Qual(ctx.GeneratedModule(encodingPackageName), "NewDecoder").Call(
					j.Lit(m.ContentType),
					j.Id("envelope"),
				)
				bg.Op(fmt.Sprintf(`
					if err := dec.Decode(&%[1]s.Payload); err != nil {
						return err
					}`, rn))
				if m.HeadersTypePromise != nil {
					if len(m.HeadersTypePromise.Target().Fields) > 0 { // Object defined as empty should not provide code
						bg.Op("headers := envelope.Headers()")
						for _, f := range m.HeadersTypePromise.Target().Fields {
							fType := j.Add(utils.ToCode(f.Type.RenderUsage(ctx))...)
							bg.If(j.Op("v, ok := headers").Index(j.Lit(f.Name)), j.Id("ok")).
								Block(j.Id(rn).Dot("Headers").Dot(f.Name).Op("=").Id("v").Assert(fType))
						}
					}
				} else {
					bg.Id(rn).Dot("Headers").Op("=").Add(utils.ToCode(m.HeadersFallbackType.RenderUsage(ctx))...).Call(
						j.Op("envelope.Headers()"),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}

func (m Message) getServerProtocols(ctx *common.RenderContext) []string {
	res := lo.FilterMap(m.AllServersPromises, func(item *Promise[*Server], _ int) (string, bool) {
		_, ok := ctx.ProtoRenderers[item.Target().Protocol]
		if !ok {
			ctx.Logger.Warnf("Skip protocol %q since it is not supported", item.Target().Protocol)
		}
		return item.Target().Protocol, ok
	})
	sort.Strings(res)
	return res
}
