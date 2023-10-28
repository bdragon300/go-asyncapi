package kafka

import (
	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
)

// Pub
type (
	Producer       = run.Producer[ChannelBindings, EnvelopeWriter]
	Publisher      = run.Publisher[EnvelopeWriter]
	EnvelopeWriter interface {
		run.EnvelopeWriter
		SetBindings(provider MessageBindings)
		SetTopic(topic string)
	}
)

type EnvelopeMarshaler interface {
	MarshalKafkaEnvelope(envelope EnvelopeWriter) error
}

// Sub
type (
	Consumer       = run.Consumer[ChannelBindings, EnvelopeReader]
	Subscriber     = run.Subscriber[EnvelopeReader]
	EnvelopeReader interface {
		run.EnvelopeReader
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalKafkaEnvelope(envelope EnvelopeReader) error
}
