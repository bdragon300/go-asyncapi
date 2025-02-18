package udp

import (
	"context"
	"io"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
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

		SetRemoteAddr(addr net.Addr)
	}
)

type EnvelopeMarshaler interface {
	MarshalEnvelopeUDP(envelope EnvelopeWriter) error
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

		RemoteAddr() net.Addr
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalEnvelopeUDP(envelope EnvelopeReader) error
}
