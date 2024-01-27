package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type ComponentsItem struct {
	Schemas types.OrderedMap[string, Object] `json:"schemas" yaml:"schemas" cgen:"directRender,components,pkgScope=models"`
	Servers types.OrderedMap[string, Server] `json:"servers" yaml:"servers" cgen:"directRender,components,pkgScope=servers"`
	// ServerVariables don't get rendered directly, only as a part of other object. However, they have to be compiled as separate objects
	ServerVariables types.OrderedMap[string, ServerVariable] `json:"serverVariables" yaml:"serverVariables" cgen:"components"`
	Channels        types.OrderedMap[string, Channel]        `json:"channels" yaml:"channels" cgen:"directRender,components,pkgScope=channels"`
	Messages        types.OrderedMap[string, Message]        `json:"messages" yaml:"messages" cgen:"directRender,components,pkgScope=messages"`
	Parameters      types.OrderedMap[string, Parameter]      `json:"parameters" yaml:"parameters" cgen:"directRender,components,pkgScope=parameters"`
	// CorrelationIDs don't get rendered directly, only as a part of other object. However, they have to be compiled as separate objects
	CorrelationIDs types.OrderedMap[string, CorrelationID] `json:"correlationIds" yaml:"correlationIds" cgen:"directRender,components"`

	// Bindings don't get rendered directly, only as a part of other object. However, they have to be compiled as separate objects
	ServerBindings    types.OrderedMap[string, ServerBindings]   `json:"serverBindings" yaml:"serverBindings" cgen:"components"`
	ChannelBindings   types.OrderedMap[string, ChannelBindings]  `json:"channelBindings" yaml:"channelBindings" cgen:"components"`
	OperationBindings types.OrderedMap[string, OperationBinding] `json:"operationBindings" yaml:"operationBindings" cgen:"components"`
	MessageBindings   types.OrderedMap[string, MessageBindings]  `json:"messageBindings" yaml:"messageBindings" cgen:"components"`
}
