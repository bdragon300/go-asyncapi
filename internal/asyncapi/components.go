package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type ComponentsItem struct {
	Schemas types.OrderedMap[string, Object] `json:"schemas" yaml:"schemas" cgen:"components,data_model,definition,selectable"`

	Servers    types.OrderedMap[string, Server]    `json:"servers" yaml:"servers" cgen:"components"`
	Channels   types.OrderedMap[string, Channel]   `json:"channels" yaml:"channels" cgen:"components"`
	Operations types.OrderedMap[string, Operation] `json:"operations" yaml:"operations" cgen:"components"`
	Messages   types.OrderedMap[string, Message]   `json:"messages" yaml:"messages" cgen:"components"`

	// SecuritySchemes types.OrderedMap[string, SecurityScheme] `json:"securitySchemes" yaml:"securitySchemes" cgen:"components"`
	ServerVariables types.OrderedMap[string, ServerVariable]        `json:"serverVariables" yaml:"serverVariables" cgen:"components"`
	Parameters      types.OrderedMap[string, Parameter]             `json:"parameters" yaml:"parameters" cgen:"components"`
	CorrelationIDs  types.OrderedMap[string, CorrelationID]         `json:"correlationIds" yaml:"correlationIds" cgen:"components"`
	Replies         types.OrderedMap[string, OperationReply]        `json:"replies" yaml:"replies" cgen:"components"`
	ReplyAddresses  types.OrderedMap[string, OperationReplyAddress] `json:"replyAddresses" yaml:"replyAddresses" cgen:"components"`
	ExternalDocs    types.OrderedMap[string, ExternalDocumentation] `json:"externalDocs" yaml:"externalDocs" cgen:"components"`
	Tags            types.OrderedMap[string, Tag]                   `json:"tags" yaml:"tags" cgen:"components"`

	OperationTraits types.OrderedMap[string, OperationTrait] `json:"operationTraits" yaml:"operationTraits" cgen:"components"`
	MessageTraits   types.OrderedMap[string, MessageTrait]   `json:"messageTraits" yaml:"messageTraits" cgen:"components"`

	ServerBindings    types.OrderedMap[string, ServerBindings]   `json:"serverBindings" yaml:"serverBindings" cgen:"components"`
	ChannelBindings   types.OrderedMap[string, ChannelBindings]  `json:"channelBindings" yaml:"channelBindings" cgen:"components"`
	OperationBindings types.OrderedMap[string, OperationBinding] `json:"operationBindings" yaml:"operationBindings" cgen:"components"`
	MessageBindings   types.OrderedMap[string, MessageBindings]  `json:"messageBindings" yaml:"messageBindings" cgen:"components"`
}
