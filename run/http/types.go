package http

type (
	ServerBindings struct{}

	ChannelBindings struct {}

	OperationBindings struct {
		Type  string
		Method string
		Query  any // jsonschema contents
	}

	MessageBindings struct {
		Headers any // jsonschema contents
	}
)
