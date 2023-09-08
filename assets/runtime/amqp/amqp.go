package amqp

import "context"

type ConsumerInfo struct {
	Topics []string
}

type EnvelopeMeta struct {
	Exchange string
	Queue    string
}

type OutEnvelope struct {
	Payload  []byte
	Metadata EnvelopeMeta
}

type InEnvelope struct {
	Payload  []byte
	Metadata EnvelopeMeta
}

type ServerParams struct {
	URL             string
	ProtocolVersion string
}

type ChannelParams struct{}

type Producer interface {
	Produce(ctx context.Context, params ChannelParams, msgs []OutEnvelope) error
}

type Consumer interface {
	Consume(ctx context.Context, params ChannelParams) (<-chan *InEnvelope, error)
}
