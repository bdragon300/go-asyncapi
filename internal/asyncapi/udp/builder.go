package udp

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "udp"
}
