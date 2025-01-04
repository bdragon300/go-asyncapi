package ws

// Params below are passed to the New* implementation functions
type (
	ServerBindings struct {}
	ChannelBindings struct {
		Method string
		Query any // jsonschema
		Headers any // jsonschema
	}
	OperationBindings struct {}
	MessageBindings struct {}
)
