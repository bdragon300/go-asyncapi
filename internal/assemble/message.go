package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type Message struct {
	OutStruct           *Struct
	InStruct            *Struct
	PayloadType         common.GolangType
	PayloadHasSchema    bool
	HeadersFallbackType *Map
	HeadersTypeLink     *Link[*Struct]
	AllServers          *LinkList[*Server] // For extracting all using protocols
}

func (m Message) AllowRender() bool {
	return true
}

func (m Message) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement

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
		res = append(res, m.assemblePublishProtoMethod(ctx, utils.ToGolangName(srv.Protocol, false))...)
	}
	res = append(res, m.assemblePublishCommonMethods(ctx)...)

	return res
}

func (m Message) assemblePublishProtoMethod(ctx *common.AssembleContext, protoPackage string) []*j.Statement {
	rn := m.OutStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(m.OutStruct.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id(fmt.Sprintf("Marshal%sEnvelope", utils.ToGolangName(protoPackage, true))).
			Params(j.Id("envelope").Qual(ctx.RuntimePackage(protoPackage), "EnvelopeWriter")).
			Error().
			BlockFunc(func(blockGroup *j.Group) {
				blockGroup.Add(utils.QualSprintf(`
					enc := %Q(encoding/json,NewEncoder)(envelope)
					if err := enc.Encode(%[1]s.Payload); err != nil {
						return err
					}`, rn))
				if m.HeadersTypeLink != nil {
					blockGroup.Id("envelope").Dot("SetHeaders").Call(
						j.Qual(ctx.RuntimePackage(""), "Header").Values(j.DictFunc(func(d j.Dict) {
							for _, f := range m.HeadersTypeLink.Target().Fields {
								d[j.Lit(f.Name)] = j.Id(rn).Dot("Headers").Dot(f.Name)
							}
						})),
					)
				} else {
					blockGroup.Id("envelope.SetHeaders").Call(
						j.Qual(ctx.RuntimePackage(""), "Header").Call(j.Id(rn).Dot("Headers")),
					)
				}
				blockGroup.Return(j.Nil())
			}),
	}
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
		j.Func().Params(receiver.Clone()).Id("WithID").
			Params(j.Id("ID").String()).
			Params(j.Op("*").Id(structName)).
			Block(
				j.Op(fmt.Sprintf(`
					%[1]s.ID = ID
					return %[1]s`, rn)),
			),
		j.Func().Params(receiver.Clone()).Id("WithPayload").
			Params(j.Id("payload").Add(payloadFieldType...)).
			Params(j.Op("*").Id(structName)).
			Block(
				j.Op(fmt.Sprintf(`
					%[1]s.Payload = payload
					return %[1]s`, rn)),
			),
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
		res = append(res, m.assembleSubscribeProtoMethod(ctx, utils.ToGolangName(srv.Protocol, false))...)
	}
	res = append(res, m.assembleSubscribeCommonMethods(ctx)...)

	return res
}

func (m Message) assembleSubscribeProtoMethod(ctx *common.AssembleContext, protoPackage string) []*j.Statement {
	rn := m.InStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(m.InStruct.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id(fmt.Sprintf("Unmarshal%sEnvelope", utils.ToGolangName(protoPackage, true))).
			Params(j.Id("envelope").Qual(ctx.RuntimePackage(protoPackage), "EnvelopeReader")).
			Error().
			BlockFunc(func(blockGroup *j.Group) {
				blockGroup.Add(utils.QualSprintf(`
					dec := %Q(encoding/json,NewDecoder)(envelope)
					if err := dec.Decode(&%[1]s.Payload); err != nil {
						return err
					}`, rn))
				if m.HeadersTypeLink != nil {
					for _, f := range m.HeadersTypeLink.Target().Fields {
						fType := j.Add(utils.ToCode(f.Type.AssembleUsage(ctx))...)
						blockGroup.If(j.Op("v, ok := headers").Index(j.Lit(f.Name)), j.Id("ok")).
							Block(j.Id(rn).Dot("Headers").Id(f.Name).Op("=").Id("v").Assert(fType))
					}
				} else {
					blockGroup.Id(rn).Dot("Headers").Op("=").Add(utils.ToCode(m.HeadersFallbackType.AssembleUsage(ctx))...).Call(j.Op("envelope.Headers()"))
				}
				blockGroup.Return(j.Nil())
			}),
	}
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
		j.Func().Params(receiver.Clone()).Id("MessageID").
			Params().
			String().
			Block(
				j.Return(j.Id(rn).Dot("ID")),
			),
		j.Func().Params(receiver.Clone()).Id("MessagePayload").
			Params().
			Add(payloadFieldType...).
			Block(
				j.Return(j.Id(rn).Dot("Payload")),
			),
		j.Func().Params(receiver.Clone()).Id("MessageHeaders").
			Params().
			Add(headersFieldType...).
			Block(
				j.Return(j.Id(rn).Dot("Headers")),
			),
	}
}
