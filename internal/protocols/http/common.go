package http

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/compile"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
)

const (
	ProtoName = "http"
	protoAbbr = "HTTP"
)

func Register() {
	compile.ProtoServerCompiler[ProtoName] = BuildServer
	compile.ProtoChannelCompiler[ProtoName] = BuildChannel
	compile.ProtoMessageBindingsBuilder[ProtoName] = BuildMessageBindingsFunc
	render.ProtoMessageMarshalEnvelopeMethodRenderer[ProtoName] = RenderMessageMarshalEnvelopeMethod
	render.ProtoMessageUnmarshalEnvelopeMethodRenderer[ProtoName] = RenderMessageUnmarshalEnvelopeMethod
}
