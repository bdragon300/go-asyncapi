package amqp

import (
	"context"
	"io"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
)

// Pub
type (
	Producer interface {
		Publisher(channelName string, bindings *ChannelBindings) (Publisher, error)
	}
	Publisher interface {
		Send(ctx context.Context, envelopes ...EnvelopeWriter) error
		Close() error
	}
	EnvelopeWriter interface {
		io.Writer
		ResetPayload()
		SetHeaders(headers run.Headers)
		SetContentType(contentType string)
		SetBindings(bindings MessageBindings)

		SetDeliveryTag(tag string)
	}
)

type EnvelopeMarshaler interface {
	MarshalAMQPEnvelope(envelope EnvelopeWriter) error
}

// Sub
type (
	Consumer interface {
		Subscriber(channelName string, bindings *ChannelBindings) (Subscriber, error)
	}
	Subscriber interface {
		Receive(ctx context.Context, cb func(envelope EnvelopeReader) error) error
		Close() error
	}
	EnvelopeReader interface {
		io.Reader
		Headers() run.Headers

		Ack() error
		Nack(requeue bool) error
		Reject(requeue bool) error
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalAMQPEnvelope(envelope EnvelopeReader) error
}
