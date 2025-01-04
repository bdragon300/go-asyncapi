package amqp

import (
	"time"
)

type DeliveryMode int

const (
	DeliveryModeTransient  DeliveryMode = 1
	DeliveryModePersistent DeliveryMode = 2
)

type ExchangeType string

const (
	ExchangeTypeDefault ExchangeType = "default"
	ExchangeTypeTopic   ExchangeType = "topic"
	ExchangeTypeDirect  ExchangeType = "direct"
	ExchangeTypeFanout  ExchangeType = "fanout"
	ExchangeTypeHeaders ExchangeType = "headers"
)

type ChannelType string

const (
	ChannelTypeRoutingKey ChannelType = "routingKey"
	ChannelTypeQueue      ChannelType = "queue"
)

type (
	ServerBindings struct{}

	ChannelBindings struct {
		ChannelType           ChannelType
		ExchangeConfiguration ExchangeConfiguration
		QueueConfiguration    QueueConfiguration
	}

	OperationBindings struct {
		Expiration   time.Duration
		UserID       string
		CC           []string
		Priority     int
		DeliveryMode DeliveryMode
		Mandatory    bool
		BCC          []string
		ReplyTo      string
		Timestamp    bool
		Ack          bool
	}

	MessageBindings struct {
		ContentEncoding string
		MessageType     string
	}

	ExchangeConfiguration struct {
		Name       *string // Empty name points to default broker exchange
		Type       ExchangeType
		Durable    *bool
		AutoDelete *bool
		VHost      string
	}

	QueueConfiguration struct {
		Name       string
		Durable    *bool
		Exclusive  *bool
		AutoDelete *bool
		VHost      string
	}
)
