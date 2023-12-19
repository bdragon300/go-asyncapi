package http

import (
	"encoding/json"
	"errors"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	j "github.com/dave/jennifer/jen"
)

type messageBindings struct {
	Headers any `json:"headers" yaml:"headers"` // jsonschema object
}

func BuildMessageBindingsFunc(ctx *common.CompileContext, message *asyncapi.Message, bindingsStruct *render.Struct, _ string) (common.Renderer, error) {
	msgBindings, ok := message.Bindings.Get(ProtoName)
	if !ok {
		return nil, types.CompileError{Err: errors.New("expected message bindings for protocol"), Path: ctx.PathRef(), Proto: ProtoName}
	}
	var bindings messageBindings
	if err := types.UnmarshalRawsUnion2(msgBindings, &bindings); err != nil {
		return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
	}
	var values types.OrderedMap[string, any]
	var jsonValues types.OrderedMap[string, string]
	if bindings.Headers != nil {
		v, err := json.Marshal(bindings.Headers)
		if err != nil {
			return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
		}
		jsonValues.Set("Headers", string(v))
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
