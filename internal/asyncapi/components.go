package asyncapi

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
)

type ComponentsItem struct {
	Schemas  types.OrderedMap[string, Object]  `json:"schemas" yaml:"schemas" cgen:"noinline,packageDown=models"`
	Messages types.OrderedMap[string, Message] `json:"messages" yaml:"messages" cgen:"noinline,packageDown=messages"`
	// TODO: maybe it's needed to make a difference between channels/servers in components and root of schema?
	Channels types.OrderedMap[string, Channel] `json:"channels" yaml:"channels" cgen:"noinline,packageDown=channels"`
	// TODO: Channels are also known as "topics", "routing keys", "event types" or "paths".
	Servers    types.OrderedMap[string, Server]    `json:"servers" yaml:"servers" cgen:"noinline,packageDown=servers"`
	Parameters types.OrderedMap[string, Parameter] `json:"parameters" yaml:"parameters" cgen:"noinline,packageDown=parameters"`
}
