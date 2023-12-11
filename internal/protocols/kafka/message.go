package kafka

import (
	"encoding/json"
	"errors"

	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type messageBindings struct {
	Key                     any    `json:"key" yaml:"key"` // jsonschema object
	SchemaIDLocation        string `json:"schemaIdLocation" yaml:"schemaIdLocation"`
	SchemaIDPayloadEncoding string `json:"schemaIdPayloadEncoding" yaml:"schemaIdPayloadEncoding"`
	SchemaLookupStrategy    string `json:"schemaLookupStrategy" yaml:"schemaLookupStrategy"`
}

func BuildMessageBindingsFunc(ctx *common.CompileContext, message *asyncapi.Message, bindingsStruct *render.Struct, _ string) (common.Renderer, error) {
	msgBindings, ok := message.Bindings.Get(ProtoName)
	if !ok {
		return nil, common.CompileError{Err: errors.New("expected message bindings for protocol"), Path: ctx.PathRef(), Proto: ProtoName}
	}
	var bindings messageBindings
	if err := utils.UnmarshalRawsUnion2(msgBindings, &bindings); err != nil {
		return nil, common.CompileError{Err: err, Path: ctx.PathRef()}
	}
	var values utils.OrderedMap[string, any]
	marshalFields := []string{"SchemaIDLocation", "SchemaIDPayloadEncoding", "SchemaLookupStrategy"}
	if err := utils.StructToOrderedMap(bindings, &values, marshalFields); err != nil {
		return nil, common.CompileError{Err: err, Path: ctx.PathRef()}
	}
	var jsonValues utils.OrderedMap[string, string]
	if bindings.Key != nil {
		v, err := json.Marshal(bindings.Key)
		if err != nil {
			return nil, common.CompileError{Err: err, Path: ctx.PathRef()}
		}
		jsonValues.Set("Key", string(v))
	}

	return &render.Func{
		FuncSignature: render.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []render.FuncParam{
				{Type: render.Simple{Name: "MessageBindings", Package: ctx.RuntimePackage(ProtoName)}},
			},
		},
		Receiver:     bindingsStruct,
		PackageName:  ctx.TopPackageName(),
		BodyRenderer: protocols.MessageBindingsBody(values, &jsonValues, ProtoName),
	}, nil
}

func RenderMessageMarshalEnvelopeMethod(ctx *common.RenderContext, message *render.Message) []*j.Statement {
	return protocols.RenderMessageMarshalEnvelopeMethod(ctx, message, ProtoName, protoAbbr)
}

func RenderMessageUnmarshalEnvelopeMethod(ctx *common.RenderContext, message *render.Message) []*j.Statement {
	return protocols.RenderMessageUnmarshalEnvelopeMethod(ctx, message, ProtoName, protoAbbr)
}
