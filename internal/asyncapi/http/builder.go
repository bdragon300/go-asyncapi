package http

type ProtoBuilder struct{}

func (pb ProtoBuilder) Protocol() string {
	return "http"
}
