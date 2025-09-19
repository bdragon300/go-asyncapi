package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type ComponentsItem struct {
	Schemas types.OrderedMap[string, Object] `json:"schemas" yaml:"schemas" cgen:"data_model,selectable"`

	Servers    types.OrderedMap[string, Server]    `json:"servers" yaml:"servers"`
	Channels   types.OrderedMap[string, Channel]   `json:"channels" yaml:"channels"`
	Operations types.OrderedMap[string, Operation] `json:"operations" yaml:"operations"`
	Messages   types.OrderedMap[string, Message]   `json:"messages" yaml:"messages"`

	// SecuritySchemes types.OrderedMap[string, SecurityScheme] `json:"securitySchemes" yaml:"securitySchemes"`
	ServerVariables types.OrderedMap[string, ServerVariable]        `json:"serverVariables" yaml:"serverVariables"`
	Parameters      types.OrderedMap[string, Parameter]             `json:"parameters" yaml:"parameters"`
	CorrelationIDs  types.OrderedMap[string, CorrelationID]         `json:"correlationIds" yaml:"correlationIds"`
	Replies         types.OrderedMap[string, OperationReply]        `json:"replies" yaml:"replies"`
	ReplyAddresses  types.OrderedMap[string, OperationReplyAddress] `json:"replyAddresses" yaml:"replyAddresses"`
	ExternalDocs    types.OrderedMap[string, ExternalDocumentation] `json:"externalDocs" yaml:"externalDocs"`
	Tags            types.OrderedMap[string, Tag]                   `json:"tags" yaml:"tags"`

	OperationTraits types.OrderedMap[string, OperationTrait] `json:"operationTraits" yaml:"operationTraits"`
	MessageTraits   types.OrderedMap[string, MessageTrait]   `json:"messageTraits" yaml:"messageTraits"`

	ServerBindings    types.OrderedMap[string, ServerBindings]   `json:"serverBindings" yaml:"serverBindings"`
	ChannelBindings   types.OrderedMap[string, ChannelBindings]  `json:"channelBindings" yaml:"channelBindings"`
	OperationBindings types.OrderedMap[string, OperationBinding] `json:"operationBindings" yaml:"operationBindings"`
	MessageBindings   types.OrderedMap[string, MessageBindings]  `json:"messageBindings" yaml:"messageBindings"`
}
