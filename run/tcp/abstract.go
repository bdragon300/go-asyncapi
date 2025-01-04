package tcp

import (
	"context"
	"io"

	"github.com/bdragon300/go-asyncapi/run"
)

// Pub
type (
	Producer interface {
		Publisher(ctx context.Context, channelName string, chBindings *ChannelBindings, opBindings *OperationBindings) (Publisher, error)
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
	}
)

type EnvelopeMarshaler interface {
	MarshalEnvelopeTCP(envelope EnvelopeWriter) error
}

// Sub
type (
	Consumer interface {
		Subscriber(ctx context.Context, channelName string, chBindings *ChannelBindings, opBindings *OperationBindings) (Subscriber, error)
	}
	Subscriber interface {
		Receive(ctx context.Context, cb func(envelope EnvelopeReader)) error
		Close() error
	}
	EnvelopeReader interface {
		io.Reader
		Headers() run.Headers
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalEnvelopeTCP(envelope EnvelopeReader) error
}
