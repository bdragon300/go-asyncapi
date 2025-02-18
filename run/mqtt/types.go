package mqtt

import (
	"time"
)

type PayloadFormatIndicator int

const (
	PayloadFormatIndicatorUnspecified PayloadFormatIndicator = 0
	PayloadFormatIndicatorUTF8        PayloadFormatIndicator = 1
)

// Params below are passed to the New* implementation functions
type (
	ServerBindings struct {
		ClientID              string
		CleanSession          bool
		LastWill              *LastWill
		KeepAlive             time.Duration
		SessionExpiryInterval time.Duration // MQTT >=5
		MaximumPacketSize     int           // MQTT >=5
	}

	LastWill struct {
		Topic   string
		QoS     int
		Message string
		Retain  bool
	}

	ChannelBindings struct{}

	OperationBindings struct {
		QoS                   int
		Retain                bool
		MessageExpiryInterval time.Duration // MQTT >=5
	}

	MessageBindings struct {
		PayloadFormatIndicator PayloadFormatIndicator // MQTT >=5
		CorrelationData        any                    // jsonschema contents, MQTT >=5
		ContentType            string                 // MQTT >=5
		ResponseTopic          string                 // MQTT >=5
	}
)
