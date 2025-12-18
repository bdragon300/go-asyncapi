package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type ComponentsItem struct {
	Schemas types.OrderedMap[string, Object] `json:"schemas,omitzero" yaml:"schemas" cgen:"data_model,selectable"`

	Servers    types.OrderedMap[string, Server]    `json:"servers,omitzero" yaml:"servers"`
	Channels   types.OrderedMap[string, Channel]   `json:"channels,omitzero" yaml:"channels"`
	Operations types.OrderedMap[string, Operation] `json:"operations,omitzero" yaml:"operations"`
	Messages   types.OrderedMap[string, Message]   `json:"messages,omitzero" yaml:"messages"`

	SecuritySchemes types.OrderedMap[string, SecurityScheme]        `json:"securitySchemes,omitzero" yaml:"securitySchemes"`
	ServerVariables types.OrderedMap[string, ServerVariable]        `json:"serverVariables,omitzero" yaml:"serverVariables"`
	Parameters      types.OrderedMap[string, Parameter]             `json:"parameters,omitzero" yaml:"parameters"`
	CorrelationIDs  types.OrderedMap[string, CorrelationID]         `json:"correlationIds,omitzero" yaml:"correlationIds"`
	Replies         types.OrderedMap[string, OperationReply]        `json:"replies,omitzero" yaml:"replies"`
	ReplyAddresses  types.OrderedMap[string, OperationReplyAddress] `json:"replyAddresses,omitzero" yaml:"replyAddresses"`
	ExternalDocs    types.OrderedMap[string, ExternalDocumentation] `json:"externalDocs,omitzero" yaml:"externalDocs"`
	Tags            types.OrderedMap[string, Tag]                   `json:"tags,omitzero" yaml:"tags"`

	OperationTraits types.OrderedMap[string, OperationTrait] `json:"operationTraits,omitzero" yaml:"operationTraits"`
	MessageTraits   types.OrderedMap[string, MessageTrait]   `json:"messageTraits,omitzero" yaml:"messageTraits"`

	ServerBindings    types.OrderedMap[string, ServerBindings]   `json:"serverBindings,omitzero" yaml:"serverBindings"`
	ChannelBindings   types.OrderedMap[string, ChannelBindings]  `json:"channelBindings,omitzero" yaml:"channelBindings"`
	OperationBindings types.OrderedMap[string, OperationBinding] `json:"operationBindings,omitzero" yaml:"operationBindings"`
	MessageBindings   types.OrderedMap[string, MessageBindings]  `json:"messageBindings,omitzero" yaml:"messageBindings"`
}
