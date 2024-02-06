package goredis

import (
	"strings"

	"github.com/bdragon300/go-asyncapi/run"
	runRedis "github.com/bdragon300/go-asyncapi/run/redis"
	"github.com/redis/go-redis/v9"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{Builder: &strings.Builder{}}
}

type EnvelopeOut struct {
	*strings.Builder
	headers     run.Headers
	contentType string
}

func (e *EnvelopeOut) ResetPayload() {
	e.Builder.Reset()
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	e.headers = headers
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.contentType = contentType
}

func (e *EnvelopeOut) SetBindings(_ runRedis.MessageBindings) {}

func (e *EnvelopeOut) RecordGoRedis() any {
	return e.Builder.String()
}

func NewEnvelopeIn(msg *redis.Message) *EnvelopeIn {
	return &EnvelopeIn{Message: msg, reader: strings.NewReader(msg.Payload)}
}

type EnvelopeIn struct {
	*redis.Message
	reader *strings.Reader
}

func (e *EnvelopeIn) Read(p []byte) (n int, err error) {
	return e.reader.Read(p)
}

func (e *EnvelopeIn) Headers() run.Headers {
	return nil
}
