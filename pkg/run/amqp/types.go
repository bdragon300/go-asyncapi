package amqp

import (
	"bytes"
	"time"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
)

type ChannelType int

const (
	ChannelTypeRoutingKey ChannelType = iota // Default
	ChannelTypeQueue
)

type ExchangeType int

const (
	ExchangeTypeDefault ExchangeType = iota
	ExchangeTypeTopic
	ExchangeTypeDirect
	ExchangeTypeFanout
	ExchangeTypeHeaders
)

type DeliveryMode int

const (
	DeliveryModeTransient  DeliveryMode = 1
	DeliveryModePersistent DeliveryMode = 2
)

type (
	ServerBindings struct{}

	ChannelBindings struct {
		ChannelType           ChannelType
		ExchangeConfiguration ExchangeConfiguration
		QueueConfiguration    QueueConfiguration

		PublisherBindings  PublishOperationBindings
		SubscriberBindings SubscribeOperationBindings
	}

	PublishOperationBindings struct {
		Expiration   time.Duration
		UserID       string
		CC           []string
		Priority     int
		DeliveryMode DeliveryMode
		Mandatory    bool
		BCC          []string
		ReplyTo      string
		Timestamp    bool
	}

	SubscribeOperationBindings struct {
		Expiration   time.Duration
		UserID       string
		CC           []string
		Priority     int
		DeliveryMode DeliveryMode
		ReplyTo      string
		Timestamp    bool
		Ack          bool
	}

	MessageBindings struct {
		ContentEncoding string
		MessageType     string
	}

	ExchangeConfiguration struct {
		Name       string
		Type       ExchangeType
		Durable    bool
		AutoDelete bool
		VHost      string
	}

	QueueConfiguration struct {
		Name       string
		Durable    bool
		Exclusive  bool
		AutoDelete bool
		VHost      string
	}
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{Payload: bytes.NewBuffer(make([]byte, 0))}
}

// "Fallback" variant for envelope when no implementation has been selected
type EnvelopeOut struct {
	Payload         *bytes.Buffer
	MessageHeaders  run.Headers
	MessageBindings MessageBindings

	Exchange string
	Queue    string
}

func (o *EnvelopeOut) Write(p []byte) (n int, err error) {
	return o.Payload.Write(p)
}

func (o *EnvelopeOut) SetHeaders(headers run.Headers) {
	o.MessageHeaders = headers
}

func (o *EnvelopeOut) Protocol() run.Protocol {
	return run.ProtocolAMQP
}

func (o *EnvelopeOut) SetBindings(bindings MessageBindings) {
	o.MessageBindings = bindings
}

func (o *EnvelopeOut) SetExchange(exchange string) {
	o.Exchange = exchange
}

func (o *EnvelopeOut) SetQueue(queue string) {
	o.Queue = queue
}

func (o *EnvelopeOut) ResetPayload() {
	o.Payload.Reset()
}

func NewEnvelopeIn() *EnvelopeIn {
	return &EnvelopeIn{Payload: bytes.NewBuffer(make([]byte, 0))}
}

// "Fallback" variant for envelope when no implementation has been selected
type EnvelopeIn struct {
	Payload        *bytes.Buffer
	MessageHeaders run.Headers

	Exchange string
	Queue    string
}

func (i *EnvelopeIn) Read(p []byte) (n int, err error) {
	return i.Payload.Read(p)
}

func (i *EnvelopeIn) Headers() run.Headers {
	return i.MessageHeaders
}

func (i *EnvelopeIn) Protocol() run.Protocol {
	return run.ProtocolAMQP
}

func (i *EnvelopeIn) Ack() {
	panic("implement me")
}

func (i *EnvelopeIn) Nack() {
	panic("implement me")
}

