package kafka

import (
	"bytes"
	"time"

	"github.com/bdragon300/asyncapi-codegen/pkg/run"
)

// Params below are passed to the New* implementation functions
type (
	ServerBindings struct {
		SchemaRegistryURL    string
		SchemaRegistryVendor string
	}

	ChannelBindings struct {
		Topic              string
		Partitions         int
		Replicas           int
		PublisherBindings  OperationBindings
		SubscriberBindings OperationBindings
		TopicConfiguration TopicConfiguration
	}

	OperationBindings struct {
		ClientID any // jsonschema contents
		GroupID  any // jsonschema contents
	}

	MessageBindings struct {
		Key                     any // TODO: jsonschema
		SchemaIDLocation        string
		SchemaIDPayloadEncoding string
		SchemaLookupStrategy    string
	}
	TopicConfiguration struct {
		CleanupPolicy     TopicCleanupPolicy
		RetentionMs       time.Duration
		RetentionBytes    int
		DeleteRetentionMs time.Duration
		MaxMessageBytes   int
	}

	TopicCleanupPolicy struct {
		Delete  bool
		Compact bool
	}
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{Payload: bytes.NewBuffer(make([]byte, 0))}
}

// "Fallback" variant for envelope when no implementation has been selected
type EnvelopeOut struct {
	Payload         *bytes.Buffer
	MessageHeaders  run.Headers
	MessageBindings MessageBindings

	Key       string
	Topic     string
	Partition int       // negative if not set
	Timestamp time.Time // If not set then will be set automatically
}

func (o *EnvelopeOut) Write(p []byte) (n int, err error) {
	return o.Payload.Write(p)
}

func (o *EnvelopeOut) SetHeaders(headers run.Headers) {
	o.MessageHeaders = headers
}

func (o *EnvelopeOut) Protocol() run.Protocol {
	return run.ProtocolKafka
}

func (o *EnvelopeOut) SetBindings(bindings MessageBindings) {
	o.MessageBindings = bindings
}

func (o *EnvelopeOut) ResetPayload() {
	o.Payload.Reset()
}

func (o *EnvelopeOut) SetTopic(topic string) {
	o.Topic = topic
}

func NewEnvelopeIn() *EnvelopeIn {
	return &EnvelopeIn{Payload: bytes.NewBuffer(make([]byte, 0))}
}

// "Fallback" variant for envelope when no implementation has been selected
type EnvelopeIn struct {
	Payload        *bytes.Buffer
	MessageHeaders run.Headers

	Topic     string
	Partition int // negative if not set
	Offset    int64
	Timestamp time.Time // If not set then will be set automatically
}

func (i *EnvelopeIn) Read(p []byte) (n int, err error) {
	return i.Payload.Read(p)
}

func (i *EnvelopeIn) Headers() run.Headers {
	return i.MessageHeaders
}

func (i *EnvelopeIn) Protocol() run.Protocol {
	return run.ProtocolKafka
}

func (i *EnvelopeIn) Commit() {
	panic("implement me")
}
