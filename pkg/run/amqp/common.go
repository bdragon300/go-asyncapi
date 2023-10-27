package amqp

import "github.com/bdragon300/asyncapi-codegen/pkg/run"

// Pub
type (
	Producer       = run.Producer[ChannelBindings, EnvelopeWriter]
	Publisher      = run.Publisher[EnvelopeWriter]
	EnvelopeWriter interface {
		run.EnvelopeWriter
		SetBindings(provider MessageBindings)
		SetExchange(exchange string)
		SetQueue(queue string)
	}
)

type EnvelopeMarshaler interface {
	MarshalAMQPEnvelope(envelope EnvelopeWriter) error
}

// Sub
type (
	Consumer       = run.Consumer[ChannelBindings, EnvelopeReader]
	Subscriber     = run.Subscriber[EnvelopeReader]
	EnvelopeReader interface {
		run.EnvelopeReader
		Ack()
		Nack()
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalAMQPEnvelope(envelope EnvelopeReader) error
}
