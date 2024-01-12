package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type ComponentsItem struct {
	Schemas  types.OrderedMap[string, Object]  `json:"schemas" yaml:"schemas" cgen:"directRender,pkgScope=models"`
	Messages types.OrderedMap[string, Message] `json:"messages" yaml:"messages" cgen:"directRender,pkgScope=messages"`
	// TODO: maybe it's needed to make a difference between channels/servers in components and root of schema?
	Channels types.OrderedMap[string, Channel] `json:"channels" yaml:"channels" cgen:"directRender,pkgScope=channels"`
	// TODO: Channels are also known as "topics", "routing keys", "event types" or "paths".
	Servers    types.OrderedMap[string, Server]    `json:"servers" yaml:"servers" cgen:"directRender,pkgScope=servers"`
	Parameters types.OrderedMap[string, Parameter] `json:"parameters" yaml:"parameters" cgen:"directRender,pkgScope=parameters"`

	// CorrelationIDs don't get rendered directly, only as a part of other object. However, they have to be compiled as separate objects
	CorrelationIDs types.OrderedMap[string, CorrelationID] `json:"correlationIds" yaml:"correlationIds" cgen:"directRender"`

	// Bindings don't get rendered directly, only as a part of other object. However, they have to be compiled as separate objects
	ServerBindings    types.OrderedMap[string, ServerBindings]   `json:"serverBindings" yaml:"serverBindings"`
	ChannelBindings   types.OrderedMap[string, ChannelBindings]  `json:"channelBindings" yaml:"channelBindings"`
	OperationBindings types.OrderedMap[string, OperationBinding] `json:"operationBindings" yaml:"operationBindings"`
	MessageBindings   types.OrderedMap[string, MessageBindings]  `json:"messageBindings" yaml:"messageBindings"`
}
