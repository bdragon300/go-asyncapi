package amqp091go

import (
	"io"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
	"github.com/bdragon300/asyncapi-codegen-go/pkg/run/amqp"
	amqp091 "github.com/rabbitmq/amqp091-go"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{
		Publishing: &amqp091.Publishing{},
	}
}

type EnvelopeOut struct {
	*amqp091.Publishing
	deliveryTag string
}

func (e *EnvelopeOut) Write(p []byte) (n int, err error) {
	e.Body = append(e.Body, p...)
	return len(p), nil
}

func (e *EnvelopeOut) ResetPayload() {
	e.Body = e.Body[:0]
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	e.Publishing.Headers = map[string]any(headers)
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.ContentType = contentType
}

func (e *EnvelopeOut) SetBindings(bindings amqp.MessageBindings) {
	e.Publishing.ContentEncoding = bindings.ContentEncoding
	e.Type = bindings.MessageType
}

func (e *EnvelopeOut) SetDeliveryTag(tag string) {
	e.deliveryTag = tag
}

type EnvelopeIn struct {
	*amqp091.Delivery
	reader io.Reader
}

func (e EnvelopeIn) Read(p []byte) (n int, err error) {
	return e.reader.Read(p)
}

func (e EnvelopeIn) Headers() run.Headers {
	return map[string]any(e.Delivery.Headers)
}

func (e EnvelopeIn) Ack() error {
	return e.Delivery.Ack(false)
}

func (e EnvelopeIn) Nack(requeue bool) error {
	return e.Delivery.Nack(false, requeue)
}