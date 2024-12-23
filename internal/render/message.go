package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

type Message struct {
	OriginalName string
	OutType      *lang.GoStruct
	InType               *lang.GoStruct
	Dummy                bool
	IsComponent bool // true if message is defined in `components` section

	HeadersFallbackType  *lang.GoMap
	HeadersTypePromise   *lang.Promise[*lang.GoStruct]

	AllServersPromise    *lang.ListPromise[*Server]    // All servers we know about

	BindingsType         *lang.GoStruct                // nil if message bindings are not defined for message
	BindingsPromise      *lang.Promise[*Bindings]      // nil if message bindings are not defined for message as well

	ContentType          string                        // Message's content type
	CorrelationIDPromise *lang.Promise[*CorrelationID] // nil if correlationID is not defined for message
	PayloadType          common.GolangType // `any` or a particular type
	AsyncAPIPromise      *lang.Promise[*AsyncAPI]

	ProtoMessages []*ProtoMessage
}

func (m *Message) Kind() common.ObjectKind {
	return common.ObjectKindMessage
}

func (m *Message) Selectable() bool {
	return !m.Dummy && !m.IsComponent // Select only the messages defined in the `channels` section`
}

func (m *Message) Visible() bool {
	return !m.Dummy
}

func (m *Message) SelectProtoObject(protocol string) common.Renderable {
	objects := lo.Filter(m.ProtoMessages, func(p *ProtoMessage, _ int) bool {
		return p.Selectable() && p.Protocol == protocol
	})
	return lo.FirstOr(objects, nil)
}

func (m *Message) Name() string {
	return utils.CapitalizeUnchanged(m.OriginalName)
}

func (m *Message) EffectiveContentType() string {
	res, _ := lo.Coalesce(m.ContentType, m.AsyncAPIPromise.T().EffectiveDefaultContentType())
	return res
}

func (m *Message) BindingsProtocols() (res []string) {
	if m.BindingsType == nil {
		return nil
	}
	if m.BindingsPromise != nil {
		res = append(res, m.BindingsPromise.T().Values.Keys()...)
		res = append(res, m.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

func (m *Message) HeadersType() *lang.GoStruct {
	if m.HeadersTypePromise != nil {
		return m.HeadersTypePromise.T()
	}
	return nil
}

func (m *Message) AllServers() []*Server {
	return m.AllServersPromise.T()
}

func (m *Message) Bindings() *Bindings {
	if m.BindingsPromise != nil {
		return m.BindingsPromise.T()
	}
	return nil
}

func (m *Message) CorrelationID() *CorrelationID {
	if m.CorrelationIDPromise != nil {
		return m.CorrelationIDPromise.T()
	}
	return nil
}

func (m *Message) AsyncAPI() *AsyncAPI {
	return m.AsyncAPIPromise.T()
}

//func (m Message) Selectable() bool {
//	return !m.Dummy
//}

//func (m Message) D(ctx *common.RenderContext) []*j.Statement {
	//var res []*j.Statement
	//ctx.LogStartRender("Message", "", m.GetOriginalName, "definition", m.Selectable())
	//defer ctx.LogFinishRender()
	//
	//// Bindings struct and its methods according to protocols of channels where the message is used
	//if m.BindingsType != nil {
	//	res = append(res, m.BindingsType.D(ctx)...)
	//
	//	if m.BindingsPromise != nil {
	//		tgt := m.BindingsPromise.Target()
	//		protocols := m.ServersProtocols(ctx)
	//		ctx.Logger.Debug("Message protocols", "protocols", protocols)
	//		for _, p := range protocols {
	//			protoTitle := ctx.ProtoRenderers[p].ProtocolTitle()
	//			res = append(res, tgt.RenderBindingsMethod(ctx, m.BindingsType, p, protoTitle)...)
	//		}
	//	}
	//}
	//
	//res = append(res, m.renderPublishMessageStruct(ctx)...)
	//res = append(res, m.renderSubscribeMessageStruct(ctx)...)
	//
	//return res
//}

//func (m Message) U(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}
//
//func (m Message) ID() string {
//	return m.GetOriginalName
//}
//
func (m *Message) String() string {
	return "Message " + m.OriginalName
}

func (m *Message) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ServerBindings", Import: common.GetContext().RuntimeModule(protoName)},
		EmptyCurlyBrackets: true,
	}
	if m.BindingsPromise != nil {
		if b, ok := m.BindingsPromise.T().Values.Get(protoName); ok {
			//ctx.Logger.Debug("Server bindings", "proto", protoName)
			res = b
		}
	}
	return res
}

//func (m Message) renderPublishMessageStruct(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderPublishMessageStruct")
//
//	var res []*j.Statement
//	res = append(res, j.Func().Id(m.OutType.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.OutType.U(ctx))...).Block(
//		j.Return(j.Op("&").Add(utils.ToCode(m.OutType.U(ctx))...).Values()),
//	))
//	res = append(res, m.OutType.D(ctx)...)
//
//	for _, p := range m.ServersProtocols(ctx) {
//		res = append(res, m.renderMarshalEnvelopeMethod(ctx, p, ctx.ProtoRenderers[p].ProtocolTitle())...)
//	}
//	res = append(res, m.renderPublishCommonMethods(ctx)...)
//
//	return res
//}

//func (m Message) renderPublishCommonMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderPublishCommonMethods")
//
//	structName := m.OutType.GetOriginalName
//	rn := m.OutType.ReceiverName()
//	receiver := j.Id(rn).Op("*").Id(structName)
//	payloadFieldType := utils.ToCode(m.PayloadType.U(ctx))
//	headersFieldType := utils.ToCode(m.HeadersFallbackType.U(ctx))
//	if m.HeadersTypePromise != nil {
//		headersFieldType = utils.ToCode(m.HeadersTypePromise.Target().U(ctx))
//	}
//
//	res := []*j.Statement{
//		// Method WithPayload(payload Model2) *Message2Out
//		j.Func().Params(receiver.Clone()).Id("WithPayload").
//			Params(j.Id("payload").Add(payloadFieldType...)).
//			Params(j.Op("*").Id(structName)).
//			Block(
//				j.Op(fmt.Sprintf(`
//					%[1]s.Payload = payload
//					return %[1]s`, rn)),
//			),
//		// Method WithHeaders(headers Message2Headers) *Message2Out
//		j.Func().Params(receiver.Clone()).Id("WithHeaders").
//			Params(j.Id("headers").Add(headersFieldType...)).
//			Params(j.Op("*").Id(structName)).
//			Block(
//				j.Op(fmt.Sprintf(`
//					%[1]s.Headers = headers
//					return %[1]s`, rn)),
//			),
//	}
//	if m.CorrelationIDPromise != nil {
//		// Method SetCorrelationID(value any)
//		res = append(res, m.CorrelationIDPromise.Target().RenderSetterDefinition(ctx, &m)...)
//	}
//	return res
//}

//func (m Message) renderMarshalEnvelopeMethod(ctx *common.RenderContext, protoName, protoTitle string) []*j.Statement {
//	ctx.Logger.Trace("renderMarshalEnvelopeMethod")
//
//	rn := m.OutType.ReceiverName()
//	receiver := j.Id(rn).Op("*").Id(m.OutType.GetOriginalName)
//
//	return []*j.Statement{
//		// Method MarshalEnvelopeProto(envelope proto.EnvelopeWriter) error
//		j.Func().Params(receiver.Clone()).Id("Marshal" + protoTitle + "Envelope").
//			Params(j.Id("envelope").Qual(ctx.RuntimeModule(protoName), "EnvelopeWriter")).
//			Error().
//			BlockFunc(func(bg *j.Group) {
//				bg.Op("enc := ").Qual(ctx.GeneratedModule(encodingPackageName), "NewEncoder").Call(
//					j.Lit(m.ContentType),
//					j.Id("envelope"),
//				)
//				bg.Op(fmt.Sprintf(`
//					if err := enc.Encode(%[1]s.Payload); err != nil {
//						return err
//					}`, rn))
//				bg.Op("envelope.SetContentType").Call(j.Lit(m.ContentType))
//				if m.HeadersTypePromise != nil {
//					bg.Id("envelope").Dot("SetHeaders").Call(
//						j.Qual(ctx.RuntimeModule(""), "Headers").Values(j.DictFunc(func(d j.Dict) {
//							for _, f := range m.HeadersTypePromise.Target().Fields {
//								d[j.Lit(f.GetOriginalName)] = j.Id(rn).Dot("Headers").Dot(f.GetOriginalName)
//							}
//						})),
//					)
//				} else {
//					bg.Id("envelope.SetHeaders").Call(
//						j.Qual(ctx.RuntimeModule(""), "Headers").Call(j.Id(rn).Dot("Headers")),
//					)
//				}
//				bg.Return(j.Nil())
//			}),
//	}
//}

//func (m Message) renderSubscribeMessageStruct(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderSubscribeMessageStruct")
//
//	var res []*j.Statement
//	res = append(res, j.Func().Id(m.InType.NewFuncName()).Params().Op("*").Add(utils.ToCode(m.InType.U(ctx))...).Block(
//		j.Return(j.Op("&").Add(utils.ToCode(m.InType.U(ctx))...).Values()),
//	))
//	res = append(res, m.InType.D(ctx)...)
//
//	for _, p := range m.ServersProtocols(ctx) {
//		res = append(res, m.renderUnmarshalEnvelopeMethod(ctx, p, ctx.ProtoRenderers[p].ProtocolTitle())...)
//	}
//	res = append(res, m.renderSubscribeCommonMethods(ctx)...)
//
//	return res
//}

//func (m Message) renderSubscribeCommonMethods(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("renderSubscribeCommonMethods")
//
//	structName := m.InType.GetOriginalName
//	rn := m.InType.ReceiverName()
//	receiver := j.Id(rn).Op("*").Id(structName)
//	payloadFieldType := utils.ToCode(m.PayloadType.U(ctx))
//	headersFieldType := utils.ToCode(m.HeadersFallbackType.U(ctx))
//	if m.HeadersTypePromise != nil {
//		headersFieldType = utils.ToCode(m.HeadersTypePromise.Target().U(ctx))
//	}
//
//	res := []*j.Statement{
//		// Method MessagePayload() Model2
//		j.Func().Params(receiver.Clone()).Id("MessagePayload").
//			Params().
//			Add(payloadFieldType...).
//			Block(
//				j.Return(j.Id(rn).Dot("Payload")),
//			),
//		// Method MessageHeaders() Message2Headers
//		j.Func().Params(receiver.Clone()).Id("MessageHeaders").
//			Params().
//			Add(headersFieldType...).
//			Block(
//				j.Return(j.Id(rn).Dot("Headers")),
//			),
//	}
//	if m.CorrelationIDPromise != nil {
//		// Method CorrelationID(value any)
//		res = append(res, m.CorrelationIDPromise.Target().RenderGetterDefinition(ctx, &m)...)
//	}
//	return res
//}

//func (m Message) renderUnmarshalEnvelopeMethod(ctx *common.RenderContext, protoName, protoTitle string) []*j.Statement {
//	ctx.Logger.Trace("renderUnmarshalEnvelopeMethod")
//
//	rn := m.InType.ReceiverName()
//	receiver := j.Id(rn).Op("*").Id(m.InType.GetOriginalName)
//
//	return []*j.Statement{
//		// Method UnmarshalEnvelopeProto(envelope proto.EnvelopeReader) error
//		j.Func().Params(receiver.Clone()).Id("Unmarshal" + protoTitle + "Envelope").
//			Params(j.Id("envelope").Qual(ctx.RuntimeModule(protoName), "EnvelopeReader")).
//			Error().
//			BlockFunc(func(bg *j.Group) {
//				bg.Op("dec := ").Qual(ctx.GeneratedModule(encodingPackageName), "NewDecoder").Call(
//					j.Lit(m.ContentType),
//					j.Id("envelope"),
//				)
//				bg.Op(fmt.Sprintf(`
//					if err := dec.Decode(&%[1]s.Payload); err != nil {
//						return err
//					}`, rn))
//				if m.HeadersTypePromise != nil {
//					if len(m.HeadersTypePromise.Target().Fields) > 0 { // Object defined as empty should not provide code
//						bg.Op("headers := envelope.Headers()")
//						for _, f := range m.HeadersTypePromise.Target().Fields {
//							fType := j.Add(utils.ToCode(f.Type.U(ctx))...)
//							bg.If(j.Op("v, ok := headers").Index(j.Lit(f.GetOriginalName)), j.Id("ok")).
//								Block(j.Id(rn).Dot("Headers").Dot(f.GetOriginalName).Op("=").Id("v").Assert(fType))
//						}
//					}
//				} else {
//					bg.Id(rn).Dot("Headers").Op("=").Add(utils.ToCode(m.HeadersFallbackType.U(ctx))...).Call(
//						j.Op("envelope.Headers()"),
//					)
//				}
//				bg.Return(j.Nil())
//			}),
//	}
//}

// ServerProtocols returns supported protocol list for the given servers, throwing out unsupported ones
//func (m Message) ServerProtocols() []string {
//	res := lo.Uniq(lo.FilterMap(m.AllServersPromise.T(), func(item *Server, _ int) (string, bool) {
//		_, ok := ctx.ProtoRenderers[item.Protocol]
//		if !ok {
//			ctx.Logger.Warnf("Skip protocol %q since it is not supported", item.Protocol)
//		}
//		return item.Protocol, ok && !item.Dummy
//	}))
//	sort.Strings(res)
//	return res
//}

type ProtoMessage struct {
	*Message
	Protocol string
}

func (p *ProtoMessage) Selectable() bool {
	return !p.Dummy && p.isBound()
}

func (p *ProtoMessage) String() string {
	return fmt.Sprintf("ProtoMessage[%s] %s", p.Protocol, p.OriginalName)
}

// isBound returns true if the message is bound to the protocol
func (p *ProtoMessage) isBound() bool {
	return lo.Contains(
		lo.Map(p.AllServersPromise.T(), func(s *Server, _ int) string { return s.Protocol }),
		p.Protocol,
	)
}