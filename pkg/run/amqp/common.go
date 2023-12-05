package amqp

import "github.com/bdragon300/asyncapi-codegen-go/pkg/run"

// Pub
type (
	Producer       = run.Producer[ChannelBindings, EnvelopeWriter]
	Publisher      = run.Publisher[EnvelopeWriter]
	EnvelopeWriter interface {
		run.EnvelopeWriter
		SetBindings(bindings MessageBindings)
		SetDeliveryTag(tag string)
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
		Ack() error
		Nack(requeue bool) error
		Reject(requeue bool) error
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalAMQPEnvelope(envelope EnvelopeReader) error
}
