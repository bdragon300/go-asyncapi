package kafka

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "kafka"
}
