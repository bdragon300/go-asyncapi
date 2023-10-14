package kafka

// TODO: fix local import
import (
	"bytes"
	"time"

	"github.com/bdragon300/asyncapi-codegen/pkg/run"
)

// Params below are passed to the New* implementation functions
type ServerBindings struct {
	SchemaRegistryURL    string
	SchemaRegistryVendor string
}

type ChannelBindings struct {
	Topic              string
	Partitions         int
	Replicas           int
	PublisherBindings  OperationBindings // TODO: implement when validation will get implemented
	SubscriberBindings OperationBindings

	// TopicConfiguration
	CleanupPolicy     TopicCleanupPolicy
	RetentionMs       time.Duration
	RetentionBytes    int
	DeleteRetentionMs time.Duration
	MaxMessageBytes   int
}

type OperationBindings struct {
	ClientID any // jsonschema contents
	GroupID  any // jsonschema contents
}

type MessageBindings struct {
	Key                     any // TODO: jsonschema
	SchemaIDLocation        string
	SchemaIDPayloadEncoding string
	SchemaLookupStrategy    string
}

type TopicCleanupPolicy struct {
	Delete  bool
	Compact bool
}

type EnvelopeMeta struct {
	Topic     string
	Partition int       // negative if not set
	Timestamp time.Time // If not set then will be set automatically
}

type EnvelopeOut struct {
	Payload         bytes.Buffer
	MessageHeaders  run.Header
	MessageMetadata EnvelopeMeta
	MessageBindings MessageBindings
}

func (o *EnvelopeOut) Write(p []byte) (n int, err error) {
	return o.Payload.Write(p)
}

func (o *EnvelopeOut) SetHeaders(headers run.Header) {
	o.MessageHeaders = headers
}

func (o *EnvelopeOut) Protocol() run.Protocol {
	return run.ProtocolKafka
}

func (o *EnvelopeOut) SetMetadata(meta EnvelopeMeta) {
	o.MessageMetadata = meta
}

func (o *EnvelopeOut) SetBindings(bindings MessageBindings) {
	o.MessageBindings = bindings
}

func (o *EnvelopeOut) ResetPayload() {
	o.Payload.Reset()
}

type EnvelopeIn struct {
	Payload         bytes.Buffer
	MessageHeaders  run.Header
	MessageMetadata EnvelopeMeta
}

func (i *EnvelopeIn) Read(p []byte) (n int, err error) {
	return i.Payload.Read(p)
}

func (i *EnvelopeIn) Headers() run.Header {
	return i.MessageHeaders
}

func (i *EnvelopeIn) Protocol() run.Protocol {
	return run.ProtocolKafka
}

func (i *EnvelopeIn) Metadata() EnvelopeMeta {
	return i.MessageMetadata
}

func (i *EnvelopeIn) Commit() {
	panic("implement me")
}
