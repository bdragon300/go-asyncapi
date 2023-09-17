package kafka

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type ProtoChannelBindings struct {
	StructValues             utils.OrderedMap[string, any]
	CleanupPolicyStructValue utils.OrderedMap[string, bool]
	// TODO: implement args validation by jsonschema
	GroupIDArgSchema  string
	ClientIDArgSchema string
}

type ProtoChannel struct {
	Name       string
	Topic      string
	Struct     *assemble.Struct
	Publisher  bool
	Subscriber bool

	PubMessageLink      *assemble.Link[*assemble.Message] // nil when message is not set
	PubChannelBindings  *ProtoChannelBindings
	SubMessageLink      *assemble.Link[*assemble.Message] // nil when message is not set
	SubChannelBindings  *ProtoChannelBindings
	FallbackMessageType common.Assembler
}

func (p ProtoChannel) AllowRender() bool {
	return true
}

func (p ProtoChannel) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	if p.PubChannelBindings != nil {
		res = append(res, p.assembleBindings(ctx, p.PubChannelBindings, "PubBindings")...)
	}
	if p.SubChannelBindings != nil {
		res = append(res, p.assembleBindings(ctx, p.SubChannelBindings, "SubBindings")...)
	}
	res = append(res, p.assembleNewFunc(ctx)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.assembleCommonMethods()...)
	if p.Publisher {
		res = append(res, p.assemblePublisherMethods(ctx)...)
	}
	if p.Subscriber {
		res = append(res, p.assembleSubscriberMethods(ctx)...)
	}
	return res
}

func (p ProtoChannel) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoChannel) assembleBindings(ctx *common.AssembleContext, bindings *ProtoChannelBindings, funcSuffix string) []*j.Statement {
	vals := lo.FromEntries(lo.Map(bindings.StructValues.Entries(), func(item lo.Entry[string, any], index int) lo.Entry[j.Code, j.Code] {
		return lo.Entry[j.Code, j.Code]{Key: j.Id(item.Key), Value: j.Lit(item.Value)}
	}))
	if bindings.CleanupPolicyStructValue.Len() > 0 {
		cleanupVals := lo.FromEntries(lo.Map(bindings.CleanupPolicyStructValue.Entries(), func(item lo.Entry[string, bool], index int) lo.Entry[j.Code, j.Code] {
			return lo.Entry[j.Code, j.Code]{Key: j.Id(item.Key), Value: j.Lit(item.Value)}
		}))
		vals[j.Id("CleanupPolicy")] = j.Qual(ctx.RuntimePackage("kafka"), "TopicCleanupPolicy").Values(j.Dict(cleanupVals))
	}

	var params []j.Code
	if bindings.GroupIDArgSchema != "" {
		params = append(params, j.Id("groupID").Id(bindings.GroupIDArgSchema))
		vals[j.Id("GroupID")] = j.Id("groupID")
	}
	if bindings.ClientIDArgSchema != "" {
		params = append(params, j.Id("clientID").Id(bindings.ClientIDArgSchema))
		vals[j.Id("ClientID")] = j.Id("clientID")
	}
	return []*j.Statement{
		j.Func().Id(p.Struct.Name+funcSuffix).
			Params(params...).
			Qual(ctx.RuntimePackage("kafka"), "ChannelBindings").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage("kafka"), "ChannelBindings").Values(j.Dict(vals))),
			),
	}
}

func (p ProtoChannel) assembleNewFunc(ctx *common.AssembleContext) []*j.Statement {
	var params []j.Code
	vals := j.Dict{}
	if p.Publisher {
		params = append(params,
			j.Id("publishers").Index().Qual(ctx.RuntimePackage(""), "Publisher").Types(
				j.Qual(ctx.RuntimePackage("kafka"), "OutEnvelope"),
			),
		)
		vals[j.Id("publishers")] = j.Id("publishers")
	}
	if p.Subscriber {
		params = append(params,
			j.Id("subscribers").Index().Qual(ctx.RuntimePackage(""), "Subscriber").Types(
				j.Qual(ctx.RuntimePackage("kafka"), "InEnvelope"),
			),
		)
		vals[j.Id("subscribers")] = j.Id("subscribers")
	}
	return []*j.Statement{
		// NewServer1Server(producer kafka.Producer, consumer kafka.Consumer) *Server1Server
		j.Func().Id(p.Struct.NewFuncName()).
			Params(params...).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(vals)),
			),
	}
}

func (p ProtoChannel) assembleCommonMethods() []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	var closeBody []j.Code
	if p.Publisher {
		closeBody = append(closeBody, utils.QualSprintf(`
			for _, pub := range %[1]s.publishers {
				err = %Q(errors,Join)(err, pub.Close())
			}`, rn),
		)
	}
	if p.Subscriber {
		closeBody = append(closeBody, utils.QualSprintf(`
			for _, sub := range %[1]s.subscribers {
				err = %Q(errors,Join)(err, sub.Close())
			}`, rn),
		)
	}
	closeBody = append(closeBody, j.Return())

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Name)),
			),

		// Method Topic() string
		j.Func().Params(receiver.Clone()).Id("Topic").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Topic)),
			),

		// Method Close() (err error)
		j.Func().Params(receiver.Clone()).Id("Close").
			Params().
			Params(j.Err().Error()).
			Block(closeBody...),
	}
}

func (p ProtoChannel) assemblePublisherMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	msgTyp := p.FallbackMessageType
	if p.PubMessageLink != nil {
		msgTyp = p.PubMessageLink.Target()
	}

	return []*j.Statement{
		// Method MakeEnvelope(envelope *kafka.OutEnvelope, message *Message1) error
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			Params(
				j.Id("envelope").Op("*").Qual(ctx.RuntimePackage("kafka"), "OutEnvelope"),
				j.Id("message").Op("*").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...),
			).
			Error().
			Block(
				utils.QualSprintf(`
					b, err := %Q(encoding/json,Marshal)(message.Payload)
					if err != nil {
						return err
					}
					envelope.Payload = b
					envelope.Headers = nil
					envelope.Metadata.Topic = %[1]s.Topic()
					envelope.Metadata.Partition = -1
					return nil
				`, rn),
			),

		// Method Publisher() runtime.Publisher[kafka.OutEnvelope]
		j.Func().Params(receiver.Clone()).Id("Publisher").
			Params().
			Qual(ctx.RuntimePackage(""), "Publisher").Types(j.Qual(ctx.RuntimePackage("kafka"), "OutEnvelope")).
			Block(
				j.Return(j.Qual(ctx.RuntimePackage(""), "NewPublisherFanOut").Call(j.Id(rn).Dot("publishers"))),
			),

		// Method Publish(ctx context.Context, messages ...*Message1) error
		j.Func().Params(receiver.Clone()).Id("Publish").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("messages").Op("...").Op("*").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...),
			).
			Error().
			Block(
				utils.QualSprintf(`
					pub := %[1]s.Publisher()
					defer pub.Close()
					envelopes := make([]*%Q(%[2]s,OutEnvelope), 0, len(messages))
					for ind := 0; ind < len(messages); ind++ {
						buf := new(%Q(%[2]s,OutEnvelope))
						if err := %[1]s.MakeEnvelope(buf, messages[ind]); err != nil {
							return %Q(fmt,Errorf)("envelope #%%d making error: %%w", ind, err)
						}
						envelopes = append(envelopes, buf)
					}
					return pub.Send(ctx, envelopes...)`, rn, ctx.RuntimePackage("kafka")),
			),
	}
}

func (p ProtoChannel) assembleSubscriberMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	msgTyp := p.FallbackMessageType
	if p.SubMessageLink != nil {
		msgTyp = p.SubMessageLink.Target()
	}

	return []*j.Statement{
		// Method ExtractEnvelope(envelope *kafka.InEnvelope, message *Message1) error
		j.Func().Params(receiver.Clone()).Id("ExtractEnvelope").
			Params(
				j.Id("envelope").Op("*").Qual(ctx.RuntimePackage("kafka"), "InEnvelope"),
				j.Id("message").Op("*").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...),
			).
			Error().
			Block(
				utils.QualSprintf(`
					if err := %Q(encoding/json,Unmarshal)(envelope.Payload, &message.Payload); err != nil {
						return err
					}
					message.Headers = nil // TODO
					message.ID = ""
					return nil
				`),
			),

		// Method Subscriber() runtime.Subscriber[kafka.InEnvelope]
		j.Func().Params(receiver.Clone()).Id("Subscriber").
			Params().
			Qual(ctx.RuntimePackage(""), "Subscriber").Types(j.Qual(ctx.RuntimePackage("kafka"), "InEnvelope")).
			Block(
				j.Return(j.Qual(ctx.RuntimePackage(""), "NewSubscriberFanIn").Call(
					j.Id(rn).Dot("subscribers"),
					j.Len(j.Id(rn).Dot("subscribers")),
					j.False(),
				)),
			),

		// Method Subscribe(ctx context.Context, cb func(message *Message2) error) (err error)
		j.Func().Params(receiver.Clone()).Id("Subscribe").
			Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("cb").Func().Params(j.Id("message").Op("*").Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...)).Error(),
			).
			Error().
			Block(
				j.Op(fmt.Sprintf(`
					sub := %[1]s.Subscriber()
					defer sub.Close()
					return sub.Receive(ctx, func(envelope *kafka.InEnvelope) error {`, rn),
				),
				j.Id("buf").Op(":=").New(j.Add(utils.ToCode(msgTyp.AssembleUsage(ctx))...)),
				j.Op(fmt.Sprintf(`
						if err := %[1]s.ExtractEnvelope(envelope, buf); err != nil {
							return fmt.Errorf("envelope extraction error: %%w", err)
						}
						return cb(buf)
					})`, rn),
				),
			),
	}
}
