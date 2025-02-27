package mqtt

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "mqtt"
}
