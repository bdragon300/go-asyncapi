package http

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
)

const (
	ProtoName = "http"
	protoAbbr = "HTTP"
)

func Register() {
	asyncapi.ProtoServerCompiler[ProtoName] = BuildServer
	asyncapi.ProtoChannelCompiler[ProtoName] = BuildChannel
	asyncapi.ProtoMessageBindingsBuilder[ProtoName] = BuildMessageBindingsFunc
	render.ProtoMessageMarshalEnvelopeMethodRenderer[ProtoName] = RenderMessageMarshalEnvelopeMethod
	render.ProtoMessageUnmarshalEnvelopeMethodRenderer[ProtoName] = RenderMessageUnmarshalEnvelopeMethod
}
