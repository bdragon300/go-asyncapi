package compiler

import "github.com/bdragon300/asyncapi-codegen/internal/utils"

type ComponentsItem struct {
	Schemas  utils.OrderedMap[string, Object]  `json:"schemas" yaml:"schemas" cgen:"noinline,packageDown=models"`
	Messages utils.OrderedMap[string, Message] `json:"messages" yaml:"messages" cgen:"noinline,packageDown=messages"`
	// TODO: maybe it's needed to make a difference between channels/servers in components and root of schema?
	Channels utils.OrderedMap[string, Channel] `json:"channels" yaml:"channels" cgen:"noinline,packageDown=channels"`
	// TODO: Channels are also known as "topics", "routing keys", "event types" or "paths".
	Servers utils.OrderedMap[string, Server] `json:"servers" yaml:"servers" cgen:"noinline,packageDown=servers"`
}
