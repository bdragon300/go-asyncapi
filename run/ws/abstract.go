package ws

import (
	"context"
	"io"

	"github.com/bdragon300/go-asyncapi/run"
)

// Pub
type (
	Producer interface {
		NewPublisher(channelName string, bindings *ChannelBindings) (Publisher, error)
	}
	Publisher interface {
		Send(ctx context.Context, envelopes ...EnvelopeWriter) error
		Close() error
	}
	EnvelopeWriter interface {
		io.Writer
		ResetPayload()
		SetHeaders(headers run.Headers)
		SetContentType(contentType string)  // TODO: remove
		SetBindings(bindings MessageBindings)

		SetOpCode(opCode byte)
	}
)

type EnvelopeMarshaler interface {
	MarshalWebSocketEnvelope(envelope EnvelopeWriter) error
}

// Sub
type (
	Consumer interface {
		NewSubscriber(channelName string, bindings *ChannelBindings) (Subscriber, error)
	}
	Subscriber interface {
		Receive(ctx context.Context, cb func(envelope EnvelopeReader) error) error
		Close() error
	}
	EnvelopeReader interface {
		io.Reader
		Headers() run.Headers
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalWebSocketEnvelope(envelope EnvelopeReader) error
}
