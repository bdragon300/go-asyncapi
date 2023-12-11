package amqp

import (
	"errors"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type messageBindings struct {
	ContentEncoding string `json:"contentEncoding" yaml:"contentEncoding"`
	MessageType     string `json:"messageType" yaml:"messageType"`
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
	marshalFields := []string{"ContentEncoding", "MessageType"}
	if err := utils.StructToOrderedMap(bindings, &values, marshalFields); err != nil {
		return nil, common.CompileError{Err: err, Path: ctx.PathRef()}
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
		BodyRenderer: protocols.MessageBindingsBody(values, nil, ProtoName),
	}, nil
}

func RenderMessageMarshalEnvelopeMethod(ctx *common.RenderContext, message *render.Message) []*j.Statement {
	return protocols.RenderMessageMarshalEnvelopeMethod(ctx, message, ProtoName, protoAbbr)
}

func RenderMessageUnmarshalEnvelopeMethod(ctx *common.RenderContext, message *render.Message) []*j.Statement {
	return protocols.RenderMessageUnmarshalEnvelopeMethod(ctx, message, ProtoName, protoAbbr)
}
