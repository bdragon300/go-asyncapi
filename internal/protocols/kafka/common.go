package kafka

import (
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
)

const protoName = "kafka"

func Register() {
	compile.ProtoServerCompiler[protoName] = BuildServer
	compile.ProtoChannelCompiler[protoName] = BuildChannel
	compile.ProtoMessageBindingsBuilder[protoName] = BuildMessageBindingsFunc
}
