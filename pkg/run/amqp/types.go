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

const(
	ExchangeTypeDefault ExchangeType = "default"
	ExchangeTypeTopic ExchangeType = "topic"
	ExchangeTypeDirect ExchangeType = "direct"
	ExchangeTypeFanout ExchangeType = "fanout"
	ExchangeTypeHeaders ExchangeType = "headers"
)

type ChannelType string

const (
	ChannelTypeRoutingKey ChannelType = "routingKey"
	ChannelTypeQueue ChannelType = "queue"
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
		Name       *string  // Empty name points to default broker exchange
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

//func NewEnvelopeOut() *EnvelopeOut {
//	return &EnvelopeOut{Payload: bytes.NewBuffer(make([]byte, 0))}
//}
//
//// "Fallback" variant for envelope when no implementation has been selected
//type EnvelopeOut struct {
//	Payload         *bytes.Buffer
//	MessageHeaders  run.Headers
//	MessageBindings MessageBindings
//	ContentType string
//
//	DeliveryTag string
//}
//
//func (o *EnvelopeOut) Write(p []byte) (n int, err error) {
//	return o.Payload.Write(p)
//}
//
//func (o *EnvelopeOut) SetHeaders(headers run.Headers) {
//	o.MessageHeaders = headers
//}
//
//func (o *EnvelopeOut) SetContentType(contentType string) {
//	o.ContentType = contentType
//}
//
//func (o *EnvelopeOut) Protocol() run.Protocol {
//	return run.ProtocolAMQP
//}
//
//func (o *EnvelopeOut) SetBindings(bindings MessageBindings) {
//	o.MessageBindings = bindings
//}
//
//func (o *EnvelopeOut) SetDeliveryTag(tag string) {
//	o.DeliveryTag = tag
//}
//
//func (o *EnvelopeOut) ResetPayload() {
//	o.Payload.Reset()
//}
//
//func NewEnvelopeIn() *EnvelopeIn {
//	return &EnvelopeIn{Payload: bytes.NewBuffer(make([]byte, 0))}
//}
//
//// "Fallback" variant for envelope when no implementation has been selected
//type EnvelopeIn struct {
//	Payload        *bytes.Buffer
//	MessageHeaders run.Headers
//
//	Exchange string
//	Queue    string
//}
//
//func (i *EnvelopeIn) Read(p []byte) (n int, err error) {
//	return i.Payload.Read(p)
//}
//
//func (i *EnvelopeIn) Headers() run.Headers {
//	return i.MessageHeaders
//}
//
//func (i *EnvelopeIn) Protocol() run.Protocol {
//	return run.ProtocolAMQP
//}
//
//func (i *EnvelopeIn) Ack() {
//	panic("implement me")
//}
//
//func (i *EnvelopeIn) Nack() {
//	panic("implement me")
//}
//
