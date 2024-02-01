package run

import (
	"context"
	"io"
)

type AbstractProducer[B any, W AbstractEnvelopeWriter, P AbstractPublisher[W]] interface {
	Publisher(ctx context.Context, channelName string, bindings *B) (P, error)
}
type AbstractPublisher[W AbstractEnvelopeWriter] interface {
	Send(ctx context.Context, envelopes ...W) error
	Close() error
}
type AbstractEnvelopeWriter interface {
	io.Writer
	ResetPayload()
	SetHeaders(headers Headers)
	// SetContentType here because it could be set in message definition in AsyncAPI. Also, some protocols may have
	// content type property of data, e.g. AMQP.
	SetContentType(contentType string)
}

type AbstractConsumer[B any, R AbstractEnvelopeReader, S AbstractSubscriber[R]] interface {
	Subscriber(ctx context.Context, channelName string, bindings *B) (S, error)
}
type AbstractSubscriber[R AbstractEnvelopeReader] interface {
	Receive(ctx context.Context, cb func(envelope R)) error
	Close() error
}
type AbstractEnvelopeReader interface {
	io.Reader
	Headers() Headers
}
