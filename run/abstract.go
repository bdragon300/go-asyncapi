package run

import (
	"context"
	"io"
)

type AbstractProducer[CB any, OB any, W AbstractEnvelopeWriter, P AbstractPublisher[W]] interface {
	Publisher(ctx context.Context, channelName string, chBindings *CB, opBindings *OB) (P, error)
	// There is no Close method here because the generated code does not responsible for creating Producers. It is the responsibility of the user.
}
type AbstractPublisher[W AbstractEnvelopeWriter] interface {
	Send(ctx context.Context, envelopes ...W) error
	Close() error
}
type AbstractEnvelopeWriter interface {
	io.Writer
	ResetPayload()
	SetHeaders(headers Headers)
	// SetContentType is here because it may be set in message definition in AsyncAPI. Also, some protocols may have
	// content type property of data, AMQP, for instance.
	SetContentType(contentType string)
}

type AbstractConsumer[CB any, OB any, R AbstractEnvelopeReader, S AbstractSubscriber[R]] interface {
	Subscriber(ctx context.Context, channelName string, chBindings *CB, opBindings *OB) (S, error)
	// There is no Close method here because the generated code does not responsible for creating Consumers. It is the responsibility of the user.
}
type AbstractSubscriber[R AbstractEnvelopeReader] interface {
	Receive(ctx context.Context, cb func(envelope R)) error
	Close() error
}
type AbstractEnvelopeReader interface {
	io.Reader
	Headers() Headers
}
