package runtime

import (
	"context"
	"io"
)

// Pub
type Producer[B any, E EnvelopeWriter] interface {
	Publisher(topic string, bindings *B) (Publisher[E], error)
}
type Publisher[E EnvelopeWriter] interface {
	Send(ctx context.Context, envelopes ...E) error
	Close() error
}
type EnvelopeWriter interface {
	io.Writer
	ResetPayload()
	SetHeaders(headers Header)
	Protocol() Protocol
}

// Sub
type Consumer[B any, E EnvelopeReader] interface {
	Subscriber(topic string, bindings *B) (Subscriber[E], error)
}
type Subscriber[E EnvelopeReader] interface {
	Receive(ctx context.Context, cb func(envelope E) error) error
	Close() error
}
type EnvelopeReader interface {
	io.Reader
	Headers() Header
	Protocol() Protocol
}

type Header map[string]any

type Parameter interface {
	Name() string
	String() string
}
