package kafka

// TODO: fix local import
import (
	"time"

	"github.com/bdragon300/asyncapi-codegen/generated/runtime"
)

type Producer interface {
	Publisher(topic string, bindings *ChannelBindings) (runtime.Publisher[OutEnvelope], error)
}

type Consumer interface {
	Subscriber(topic string, bindings *ChannelBindings) (runtime.Subscriber[InEnvelope], error)
}

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
	ClientID string
	GroupID  string
}

type TopicCleanupPolicy struct {
	Delete  bool
	Compact bool
}

type EnvelopeMeta struct {
	Topic     string
	Partition int       // negative if not set
	Timestamp time.Time // If not set then will be set automatically

	// From message bindings
	Key                     []byte
	SchemaIDLocation        string
	SchemaIDPayloadEncoding string
	SchemaLookupStrategy    string
}

type OutEnvelope struct {
	Payload  []byte
	Headers  map[string][]byte
	Metadata EnvelopeMeta
}

type InEnvelope struct {
	Payload  []byte
	Headers  map[string][]byte
	Metadata EnvelopeMeta
}
