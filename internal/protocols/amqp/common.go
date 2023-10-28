package amqp

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
)

const (
	ProtoName = "amqp"
	protoAbbr = "AMQP"
)

func Register() {
	compile.ProtoChannelCompiler[ProtoName] = BuildChannel
	compile.ProtoMessageBindingsBuilder[ProtoName] = BuildMessageBindingsFunc
	assemble.ProtoMessageMarshalEnvelopeMethodAssembler[ProtoName] = AssembleMessageMarshalEnvelopeMethod
	assemble.ProtoMessageUnmarshalEnvelopeMethodAssembler[ProtoName] = AssembleMessageUnmarshalEnvelopeMethod
}
