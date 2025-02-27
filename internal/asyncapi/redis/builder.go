package redis

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "redis"
}
