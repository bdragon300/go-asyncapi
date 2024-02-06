package udp

type (
	ServerBindings struct {}
	ChannelBindings struct {
		LocalAddress string  // TODO: move to message bindings?
		LocalPort int
	}
	OperationBindings struct {}
	MessageBindings struct {
		MTU int // TODO: needed here?
		//TODO: flags
	}
)
