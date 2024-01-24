package run

import (
	"context"
	"io"
)

type AbstractProducer[B any, W AbstractEnvelopeWriter, P AbstractPublisher[W]] interface {
	NewPublisher(ctx context.Context, channelName string, bindings *B) (P, error)
}
type AbstractPublisher[W AbstractEnvelopeWriter] interface {
	Send(ctx context.Context, envelopes ...W) error
	Close() error
}
type AbstractEnvelopeWriter interface {
	io.Writer
	ResetPayload()
	SetHeaders(headers Headers)
	SetContentType(contentType string)
}

type AbstractConsumer[B any, R AbstractEnvelopeReader, S AbstractSubscriber[R]] interface {
	NewSubscriber(ctx context.Context, channelName string, bindings *B) (S, error)
}
type AbstractSubscriber[R AbstractEnvelopeReader] interface {
	Receive(ctx context.Context, cb func(envelope R) error) error
	Close() error
}
type AbstractEnvelopeReader interface {
	io.Reader
	Headers() Headers
}
