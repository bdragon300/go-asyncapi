package kafka

import "github.com/bdragon300/asyncapi-codegen/generated/runtime"

// Pub
type (
	Producer       = runtime.Producer[ChannelBindings, EnvelopeWriter]
	Publisher      = runtime.Publisher[EnvelopeWriter]
	EnvelopeWriter interface {
		runtime.EnvelopeWriter
		SetMetadata(meta EnvelopeMeta)
		SetBindings(provider MessageBindings)
	}
)

type EnvelopeMarshaler interface {
	MarshalKafkaEnvelope(envelope EnvelopeWriter) error
}

// Sub
type (
	Consumer       = runtime.Consumer[ChannelBindings, EnvelopeReader]
	Subscriber     = runtime.Subscriber[EnvelopeReader]
	EnvelopeReader interface {
		runtime.EnvelopeReader
		Metadata() EnvelopeMeta
		Commit()
	}
)

type EnvelopeUnmarshaler interface {
	UnmarshalKafkaEnvelope(envelope EnvelopeReader) error
}
