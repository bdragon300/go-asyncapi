package amqp

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
)

type ProtoBuilder struct {
	asyncapi.BaseProtoBuilder
}

var Builder = ProtoBuilder{
	BaseProtoBuilder: asyncapi.BaseProtoBuilder{
		ProtoName:  "amqp",
		ProtoTitle: "AMQP",
	},
}
