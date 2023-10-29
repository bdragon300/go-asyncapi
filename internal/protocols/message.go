package protocols

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

func AssembleMessageUnmarshalEnvelopeMethod(ctx *common.AssembleContext, message *assemble.Message, protoName, protoAbbr string) []*j.Statement {
	rn := message.InStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(message.InStruct.Name)

	return []*j.Statement{
		// Method UnmarshalProtoEnvelope(envelope proto.EnvelopeReader) error
		j.Func().Params(receiver.Clone()).Id("Unmarshal" + protoAbbr + "Envelope").
			Params(j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeReader")).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("dec := ").Qual(ctx.GeneratedPackage("utils"), "NewDecoder").Call(j.Lit(message.ContentType), j.Id("envelope"))
				bg.Op(fmt.Sprintf(`
					if err := dec.Decode(&%[1]s.Payload); err != nil {
						return err
					}`, rn))
				if message.HeadersTypeLink != nil {
					for _, f := range message.HeadersTypeLink.Target().Fields {
						fType := j.Add(utils.ToCode(f.Type.AssembleUsage(ctx))...)
						bg.If(j.Op("v, ok := headers").Index(j.Lit(f.Name)), j.Id("ok")).
							Block(j.Id(rn).Dot("Headers").Id(f.Name).Op("=").Id("v").Assert(fType))
					}
				} else {
					bg.Id(rn).Dot("Headers").Op("=").Add(utils.ToCode(message.HeadersFallbackType.AssembleUsage(ctx))...).Call(j.Op("envelope.Headers()"))
				}
				bg.Return(j.Nil())
			}),
	}
}

func AssembleMessageMarshalEnvelopeMethod(ctx *common.AssembleContext, message *assemble.Message, protoName, protoAbbr string) []*j.Statement {
	rn := message.OutStruct.ReceiverName()
	receiver := j.Id(rn).Op("*").Id(message.OutStruct.Name)

	return []*j.Statement{
		// Method MarshalProtoEnvelope(envelope proto.EnvelopeWriter) error
		j.Func().Params(receiver.Clone()).Id("Marshal" + protoAbbr + "Envelope").
			Params(j.Id("envelope").Qual(ctx.RuntimePackage(protoName), "EnvelopeWriter")).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("enc := ").Qual(ctx.GeneratedPackage("utils"), "NewEncoder").Call(j.Lit(message.ContentType), j.Id("envelope"))
				bg.Op(fmt.Sprintf(`
					if err := enc.Encode(%[1]s.Payload); err != nil {
						return err
					}`, rn))
				if message.HeadersTypeLink != nil {
					bg.Id("envelope").Dot("SetHeaders").Call(
						j.Qual(ctx.RuntimePackage(""), "Header").Values(j.DictFunc(func(d j.Dict) {
							for _, f := range message.HeadersTypeLink.Target().Fields {
								d[j.Lit(f.Name)] = j.Id(rn).Dot("Headers").Dot(f.Name)
							}
						})),
					)
				} else {
					bg.Id("envelope.SetHeaders").Call(
						j.Qual(ctx.RuntimePackage(""), "Headers").Call(j.Id(rn).Dot("Headers")),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}

func MessageBindingsBody(values utils.OrderedMap[string, any], jsonValues *utils.OrderedMap[string, string], protoName string) func(ctx *common.AssembleContext, p *assemble.Func) []*j.Statement {
	return func(ctx *common.AssembleContext, p *assemble.Func) []*j.Statement {
		var res []*j.Statement
		res = append(res,
			j.Id("b").Op(":=").Qual(ctx.RuntimePackage(protoName), "MessageBindings").Values(j.DictFunc(func(d j.Dict) {
				for _, e := range values.Entries() {
					d[j.Id(e.Key)] = j.Lit(e.Value)
				}
			})),
		)
		if jsonValues != nil {
			for _, e := range jsonValues.Entries() {
				n := utils.ToLowerFirstLetter(e.Key)
				res = append(res,
					j.Id(n).Op(":=").Lit(e.Value),
					j.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.%[2]s)", n, e.Key)),
				)
			}
		}
		res = append(res, j.Return(j.Id("b")))
		return res
	}
}