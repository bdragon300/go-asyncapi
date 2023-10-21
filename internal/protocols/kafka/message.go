package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/protocols"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type messageBindings struct {
	Key                     any    `json:"key" yaml:"key"` // jsonschema object
	SchemaIDLocation        string `json:"schemaIdLocation" yaml:"schemaIdLocation"`
	SchemaIDPayloadEncoding string `json:"schemaIdPayloadEncoding" yaml:"schemaIdPayloadEncoding"`
	SchemaLookupStrategy    string `json:"schemaLookupStrategy" yaml:"schemaLookupStrategy"`
}

func BuildMessageBindingsFunc(ctx *common.CompileContext, message *compile.Message, bindingsStruct *assemble.Struct, _ string) (common.Assembler, error) {
	msgBindings, ok := message.Bindings.Get(protoName)
	if !ok {
		return nil, fmt.Errorf("no binding for protocol %s", protoName)
	}
	var bindings messageBindings
	if err := utils.UnmarshalRawsUnion2(msgBindings, &bindings); err != nil {
		return nil, err
	}
	var values utils.OrderedMap[string, any]
	marshalFields := []string{"SchemaIDLocation", "SchemaIDPayloadEncoding", "SchemaLookupStrategy"}
	if err := utils.StructToOrderedMap(bindings, &values, marshalFields); err != nil {
		return nil, err
	}
	var jsonValues utils.OrderedMap[string, string]
	if bindings.Key != nil {
		v, err := json.Marshal(bindings.Key)
		if err != nil {
			return nil, err
		}
		jsonValues.Set("Key", string(v))
	}

	return &assemble.Func{
		FuncSignature: assemble.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: assemble.Simple{Type: "MessageBindings", Package: ctx.RuntimePackage(protoName)}},
			},
		},
		Receiver:      bindingsStruct,
		Package:       ctx.TopPackageName(),
		BodyAssembler: protocols.MessageBindingsBody(values, jsonValues, protoName),
	}, nil
}

func AssembleMessageMarshalEnvelopeMethod(ctx *common.AssembleContext, message *assemble.Message) []*j.Statement {
	return protocols.AssembleMessageMarshalEnvelopeMethod(ctx, message, protoName, protoAbbr)
}

func AssembleMessageUnmarshalEnvelopeMethod(ctx *common.AssembleContext, message *assemble.Message) []*j.Statement {
	return protocols.AssembleMessageUnmarshalEnvelopeMethod(ctx, message, protoName, protoAbbr)
}
