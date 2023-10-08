package amqp

// TODO: fix local import
import (
	"bytes"

	"github.com/bdragon300/asyncapi-codegen/generated/runtime"
)

type ServerBindings struct{}

type ChannelBindings struct {
	ExchangeName       string
	PublisherBindings  OperationBindings
	SubscriberBindings OperationBindings
}

type OperationBindings struct{}

type MessageBindings struct{}

type EnvelopeMeta struct {
	Exchange string
	Queue    string
}

type EnvelopeOut struct {
	Payload         bytes.Buffer
	MessageHeaders  runtime.Header
	MessageMetadata EnvelopeMeta
	MessageBindings MessageBindings
}

func (o *EnvelopeOut) Write(p []byte) (n int, err error) {
	return o.Payload.Write(p)
}

func (o *EnvelopeOut) SetHeaders(headers runtime.Header) {
	o.MessageHeaders = headers
}

func (o *EnvelopeOut) Protocol() runtime.Protocol {
	return runtime.ProtocolAMQP
}

func (o *EnvelopeOut) SetMetadata(meta EnvelopeMeta) {
	o.MessageMetadata = meta
}

func (o *EnvelopeOut) SetBindings(bindings MessageBindings) {
	o.MessageBindings = bindings
}

func (o *EnvelopeOut) ResetPayload() {
	o.Payload.Reset()
}

type EnvelopeIn struct {
	Payload         bytes.Buffer
	MessageHeaders  runtime.Header
	MessageMetadata EnvelopeMeta
}

func (i *EnvelopeIn) Read(p []byte) (n int, err error) {
	return i.Read(p)
}

func (i *EnvelopeIn) Headers() runtime.Header {
	return i.MessageHeaders
}

func (i *EnvelopeIn) Protocol() runtime.Protocol {
	return runtime.ProtocolAMQP
}

func (i *EnvelopeIn) Metadata() EnvelopeMeta {
	return i.MessageMetadata
}

func (i *EnvelopeIn) Ack() {
	panic("implement me")
}

func (i *EnvelopeIn) Nack() {
	panic("implement me")
}

