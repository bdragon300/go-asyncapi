package amqp

// TODO: fix local import
import "github.com/bdragon300/asyncapi-codegen/assets/runtime"

type AMQPProducer interface {
	Publisher(params AMQPChannelParams) (runtime.Publisher[AMQPOutEnvelope], error)
}

type AMQPConsumer interface {
	Publisher(params AMQPChannelParams) (runtime.Publisher[AMQPInEnvelope], error)
}

type AMQPConsumerInfo struct {
	Topics []string
}

type AMQPMeta struct {
	Exchange string
	Queue    string
}

type AMQPOutEnvelope struct {
	Payload  []byte
	Metadata AMQPMeta
}

type AMQPInEnvelope struct {
	Payload  []byte
	Metadata AMQPMeta
}

type AMQPServerParams struct {
	URL             string
	ProtocolVersion string
}

type AMQPChannelParams struct{}

