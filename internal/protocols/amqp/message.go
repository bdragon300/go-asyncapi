package amqp

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type messageBindings struct {
	ContentEncoding string `json:"contentEncoding" yaml:"contentEncoding"`
	MessageType     string `json:"messageType" yaml:"messageType"`
}

func BuildMessageBindingsFunc(ctx *common.CompileContext, message *compile.Message, bindingsStruct *assemble.Struct, _ string) (common.Assembler, error) {
	msgBindings, ok := message.Bindings.Get(ProtoName)
	if !ok {
		return nil, fmt.Errorf("no binding for protocol %s", ProtoName)
	}
	var bindings messageBindings
	if err := utils.UnmarshalRawsUnion2(msgBindings, &bindings); err != nil {
		return nil, err
	}
	var values utils.OrderedMap[string, any]
	marshalFields := []string{"ContentEncoding", "MessageType"}
	if err := utils.StructToOrderedMap(bindings, &values, marshalFields); err != nil {
		return nil, err
	}

	return &assemble.Func{
		FuncSignature: assemble.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: assemble.Simple{Type: "MessageBindings", Package: ctx.RuntimePackage(ProtoName)}},
			},
		},
		Receiver:      bindingsStruct,
		Package:       ctx.TopPackageName(),
		BodyAssembler: protocols.MessageBindingsBody(values, nil, ProtoName),
	}, nil
}

func AssembleMessageMarshalEnvelopeMethod(ctx *common.AssembleContext, message *assemble.Message) []*j.Statement {
	return protocols.AssembleMessageMarshalEnvelopeMethod(ctx, message, ProtoName, protoAbbr)
}

func AssembleMessageUnmarshalEnvelopeMethod(ctx *common.AssembleContext, message *assemble.Message) []*j.Statement {
	return protocols.AssembleMessageUnmarshalEnvelopeMethod(ctx, message, ProtoName, protoAbbr)
}
