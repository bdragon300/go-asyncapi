package amqp

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/compile"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
)

const (
	ProtoName = "amqp"
	protoAbbr = "AMQP"
)

func Register() {
	compile.ProtoServerCompiler[ProtoName] = BuildServer
	compile.ProtoChannelCompiler[ProtoName] = BuildChannel
	compile.ProtoMessageBindingsBuilder[ProtoName] = BuildMessageBindingsFunc
	render.ProtoMessageMarshalEnvelopeMethodRenderer[ProtoName] = RenderMessageMarshalEnvelopeMethod
	render.ProtoMessageUnmarshalEnvelopeMethodRenderer[ProtoName] = RenderMessageUnmarshalEnvelopeMethod
}
