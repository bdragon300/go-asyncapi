package kafka

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
)

const (
	ProtoName = "kafka"
	protoAbbr = "Kafka"
)

func Register() {
	compile.ProtoServerCompiler[ProtoName] = BuildServer
	compile.ProtoChannelCompiler[ProtoName] = BuildChannel
	compile.ProtoMessageBindingsBuilder[ProtoName] = BuildMessageBindingsFunc
	assemble.ProtoMessageMarshalEnvelopeMethodAssembler[ProtoName] = AssembleMessageMarshalEnvelopeMethod
	assemble.ProtoMessageUnmarshalEnvelopeMethodAssembler[ProtoName] = AssembleMessageUnmarshalEnvelopeMethod
}
