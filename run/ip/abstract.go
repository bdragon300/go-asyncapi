package ip

import (
	"context"
	"github.com/bdragon300/go-asyncapi/run"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"io"
)

// Pub
type (
	Producer interface {
		Publisher(ctx context.Context, channelName string, bindings *ChannelBindings) (Publisher, error)
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
		Subscriber(ctx context.Context, channelName string, bindings *ChannelBindings) (Subscriber, error)
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
