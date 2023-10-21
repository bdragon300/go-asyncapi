package amqp

// TODO: fix local import
import (
	"bytes"
	"time"

	"github.com/bdragon300/asyncapi-codegen/pkg/run"
)

type ServerBindings struct{}

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

type ChannelBindings struct {
	ChannelType           ChannelType
	ExchangeConfiguration ExchangeConfiguration
	QueueConfiguration    QueueConfiguration

	PublisherBindings  PublishOperationBindings
	SubscriberBindings SubscribeOperationBindings
}

type ExchangeConfiguration struct {
	Name       string
	Type       ExchangeType
	Durable    bool
	AutoDelete bool
	VHost      string
}

type QueueConfiguration struct {
	Name       string
	Durable    bool
	Exclusive  bool
	AutoDelete bool
	VHost      string
}

type PublishOperationBindings struct {
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

type SubscribeOperationBindings struct {
	Expiration   time.Duration
	UserID       string
	CC           []string
	Priority     int
	DeliveryMode DeliveryMode
	ReplyTo      string
	Timestamp    bool
	Ack          bool
}

type MessageBindings struct {
	ContentEncoding string
	MessageType     string
}

// "Fallback" variant for envelope when no implementation has been selected
type EnvelopeOut struct {
	Payload         bytes.Buffer
	MessageHeaders  run.Header
	MessageBindings MessageBindings

	Exchange string
	Queue    string
}

func (o *EnvelopeOut) Write(p []byte) (n int, err error) {
	return o.Payload.Write(p)
}

func (o *EnvelopeOut) SetHeaders(headers run.Header) {
	o.MessageHeaders = headers
}

func (o *EnvelopeOut) Protocol() run.Protocol {
	return run.ProtocolAMQP
}

func (o *EnvelopeOut) SetBindings(bindings MessageBindings) {
	o.MessageBindings = bindings
}

func (o *EnvelopeOut) ResetPayload() {
	o.Payload.Reset()
}

// "Fallback" variant for envelope when no implementation has been selected
type EnvelopeIn struct {
	Payload        bytes.Buffer
	MessageHeaders run.Header

	Exchange string
	Queue    string
}

func (i *EnvelopeIn) Read(p []byte) (n int, err error) {
	return i.Payload.Read(p)
}

func (i *EnvelopeIn) Headers() run.Header {
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

