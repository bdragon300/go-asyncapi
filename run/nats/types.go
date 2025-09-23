package nats

type (
	ServerBindings    struct{}
	ChannelBindings   struct{}
	OperationBindings struct {
		Queue string
	}
	MessageBindings struct{}
)
