package kafka

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
)

const (
	protoName = "kafka"
	protoAbbr = "Kafka"
)

func Register() {
	compile.ProtoServerCompiler[protoName] = BuildServer
	compile.ProtoChannelCompiler[protoName] = BuildChannel
	compile.ProtoMessageBindingsBuilder[protoName] = BuildMessageBindingsFunc
	assemble.ProtoMessageMarshalEnvelopeMethodAssembler[protoName] = AssembleMessageMarshalEnvelopeMethod
	assemble.ProtoMessageUnmarshalEnvelopeMethodAssembler[protoName] = AssembleMessageUnmarshalEnvelopeMethod
}
