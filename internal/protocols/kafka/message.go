package kafka

import (
	"encoding/json"
	"fmt"

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
			Name: "Kafka",
			Args: nil,
			Return: []assemble.FuncParam{
				{Type: assemble.Simple{Type: "MessageBindings", Package: ctx.RuntimePackage(protoName)}},
			},
		},
		Receiver:      bindingsStruct,
		Package:       ctx.TopPackageName(),
		BodyAssembler: messageBindingsBody(values, jsonValues),
	}, nil
}

func messageBindingsBody(values utils.OrderedMap[string, any], jsonValues utils.OrderedMap[string, string]) func(ctx *common.AssembleContext, p *assemble.Func) []*j.Statement {
	return func(ctx *common.AssembleContext, p *assemble.Func) []*j.Statement {
		var res []*j.Statement
		res = append(res,
			j.Id("b").Op(":=").Qual(ctx.RuntimePackage(protoName), "MessageBindings").Values(j.DictFunc(func(d j.Dict) {
				for _, e := range values.Entries() {
					d[j.Id(e.Key)] = j.Lit(e.Value)
				}
			})),
		)
		for _, e := range jsonValues.Entries() {
			n := utils.ToLowerFirstLetter(e.Key)
			res = append(res,
				j.Id(n).Op(":=").Lit(e.Value),
				j.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.%[2]s)", n, e.Key)),
			)
		}
		res = append(res, j.Return(j.Id("b")))
		return res
	}
}
