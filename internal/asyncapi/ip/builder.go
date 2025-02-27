package ip

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "ip"
}
