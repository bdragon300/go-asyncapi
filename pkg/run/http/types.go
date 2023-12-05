package http

type (
	ServerBindings struct{}

	ChannelBindings struct {
		PublisherBindings  OperationBindings
		SubscriberBindings OperationBindings
	}

	OperationBindings struct {
		Method string
		Query  any // jsonschema contents
	}

	MessageBindings struct {
		Headers any // jsonschema contents
	}
)
