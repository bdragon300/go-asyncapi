package kafka

import (
	"time"
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
