package http

type (
	ServerBindings struct{}

	ChannelBindings struct {}

	OperationBindings struct {
		Method string
		Query  any // jsonschema contents
	}

	MessageBindings struct {
		Headers any // jsonschema contents
		StatusCode int
	}
)
