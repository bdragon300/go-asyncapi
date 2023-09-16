package kafka

// TODO: fix local import
import (
	"time"

	"github.com/bdragon300/asyncapi-codegen/generated/runtime"
)

type Producer interface {
	Publisher(params ChannelParams) (runtime.Publisher[OutEnvelope], error)
}

type Consumer interface {
	Subscriber(params ChannelParams) (runtime.Subscriber[InEnvelope], error)
}

// Params below are passed to the New* implementation functions
type ServerParams struct {
	URL             string
	ProtocolVersion string
	// From server bindings
	SchemaRegistryURL    string
	SchemaRegistryVendor string
}

type ChannelParams struct {
	// From channel/operation bindings
	Topic      string
	Partitions int
	Replicas   int
	ClientID   string
	GroupID    string

	// TopicConfiguration
	CleanupPolicy     [2]string
	RetentionMs       time.Duration
	RetentionBytes    int
	DeleteRetentionMs time.Duration
	MaxMessageBytes   int
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
