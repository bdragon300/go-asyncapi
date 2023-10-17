package amqp

// TODO: fix local import
import (
	"bytes"

	"github.com/bdragon300/asyncapi-codegen/pkg/run"
)

type ServerBindings struct{}

type ChannelBindings struct {
	ExchangeName       string
	PublisherBindings  PublishOperationBindings
	SubscriberBindings SubscribeOperationBindings
}

type PublishOperationBindings struct{}

type SubscribeOperationBindings struct{}

type MessageBindings struct{}

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

