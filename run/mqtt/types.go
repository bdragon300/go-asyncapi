package mqtt

import (
	"time"
)

// Params below are passed to the New* implementation functions
type (
	ServerBindings struct {
		ClientID string
		CleanSession bool
		LastWill *LastWill
		KeepAlive time.Duration
	}

	LastWill struct {
		Topic string
		QoS   int
		Message string
		Retain bool
	}

	ChannelBindings struct {}

	OperationBindings struct {
		QoS int
		Retain bool
	}

	MessageBindings struct {}
)
