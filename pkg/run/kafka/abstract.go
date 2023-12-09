package kafka

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
		Protocol() run.Protocol
		SetBindings(bindings MessageBindings)

		SetTopic(topic string)
	}
)

type EnvelopeMarshaler interface {
	MarshalKafkaEnvelope(envelope EnvelopeWriter) error
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
		Protocol() run.Protocol
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalKafkaEnvelope(envelope EnvelopeReader) error
}
