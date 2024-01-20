package ws

// Params below are passed to the New* implementation functions
type (
	ServerBindings struct {}
	ChannelBindings struct {
		Method string
		Query any // jsonschema
		Headers any // jsonschema
		PublisherBindings OperationBindings
		SubscriberBindings OperationBindings
	}
	OperationBindings struct {}
	MessageBindings struct {}
)
