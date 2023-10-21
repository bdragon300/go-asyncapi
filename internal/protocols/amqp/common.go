package amqp

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
)

const (
	protoName = "amqp"
	protoAbbr = "AMQP"
)

func Register() {
	compile.ProtoChannelCompiler[protoName] = BuildChannel
	compile.ProtoMessageBindingsBuilder[protoName] = BuildMessageBindingsFunc
	assemble.ProtoMessageMarshalEnvelopeMethodAssembler[protoName] = AssembleMessageMarshalEnvelopeMethod
	assemble.ProtoMessageUnmarshalEnvelopeMethodAssembler[protoName] = AssembleMessageUnmarshalEnvelopeMethod
}
