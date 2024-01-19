package mqtt

import (
	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
)

type ProtoBuilder struct {
	asyncapi.BaseProtoBuilder
}

var Builder = ProtoBuilder{
	BaseProtoBuilder: asyncapi.BaseProtoBuilder{
		ProtoName:  "mqtt",
		ProtoTitle: "MQTT",
	},
}
