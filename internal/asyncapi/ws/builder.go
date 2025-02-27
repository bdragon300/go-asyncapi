package ws

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "amqp"
}
