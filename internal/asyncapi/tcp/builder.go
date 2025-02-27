package tcp

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "tcp"
}
