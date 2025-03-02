package ws

type (
	ServerBindings  struct{}
	ChannelBindings struct {
		Method  string
		Query   any // jsonschema
		Headers any // jsonschema
	}
	OperationBindings struct{}
	MessageBindings   struct{}
)
