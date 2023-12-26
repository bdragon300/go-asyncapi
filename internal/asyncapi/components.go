package asyncapi

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
)

type ComponentsItem struct {
	Schemas  types.OrderedMap[string, Object]  `json:"schemas" yaml:"schemas" cgen:"directRender,packageDown=models"`
	Messages types.OrderedMap[string, Message] `json:"messages" yaml:"messages" cgen:"directRender,packageDown=messages"`
	// TODO: maybe it's needed to make a difference between channels/servers in components and root of schema?
	Channels types.OrderedMap[string, Channel] `json:"channels" yaml:"channels" cgen:"directRender,packageDown=channels"`
	// TODO: Channels are also known as "topics", "routing keys", "event types" or "paths".
	Servers        types.OrderedMap[string, Server]        `json:"servers" yaml:"servers" cgen:"directRender,packageDown=servers"`
	Parameters     types.OrderedMap[string, Parameter]     `json:"parameters" yaml:"parameters" cgen:"directRender,packageDown=parameters"`
	CorrelationIDs types.OrderedMap[string, CorrelationID] `json:"correlationIds" yaml:"correlationIds" cgen:"directRender,packageDown=correlationids"`
}
