package ip

import (
	"context"
	"io"

	"github.com/bdragon300/go-asyncapi/run"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// Pub
type (
	Producer interface {
		Publisher(ctx context.Context, address string, chBindings *ChannelBindings, opBindings *OperationBindings) (Publisher, error)
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
	MarshalEnvelopeIP(envelope EnvelopeWriter) error
}

// Sub
type (
	Consumer interface {
		Subscriber(ctx context.Context, address string, chBindings *ChannelBindings, opBindings *OperationBindings) (Subscriber, error)
	}
	Subscriber interface {
		Receive(ctx context.Context, cb func(envelope EnvelopeReader)) error
		Close() error
	}
	EnvelopeReader interface {
		io.Reader
		Headers() run.Headers

		Headers4() (*ipv4.Header, error)
		Headers6() (*ipv6.Header, error)
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalEnvelopeIP(envelope EnvelopeReader) error
}
