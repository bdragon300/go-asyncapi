package franzgo

import (
	"bytes"

	"github.com/bdragon300/asyncapi-codegen-go/run"
	runKafka "github.com/bdragon300/asyncapi-codegen-go/run/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{
		Record: &kgo.Record{},
	}
}

type EnvelopeOut struct {
	*kgo.Record
	messageBindings runKafka.MessageBindings
}

func (e *EnvelopeOut) Write(p []byte) (n int, err error) {
	e.Value = append(e.Value, p...)
	return len(p), nil
}

func (e *EnvelopeOut) ResetPayload() {
	e.Value = e.Value[:0]
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	for k, v := range headers.ToByteValues() {
		e.Record.Headers = append(e.Record.Headers, kgo.RecordHeader{Key: k, Value: v})
	}
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.Record.Headers = append(e.Record.Headers, kgo.RecordHeader{Key: "Content-Type", Value: []byte(contentType)})
}

func (e *EnvelopeOut) SetBindings(bindings runKafka.MessageBindings) {
	e.messageBindings = bindings
}

func (e *EnvelopeOut) SetTopic(topic string) {
	e.Topic = topic
}

func NewEnvelopeIn(r *kgo.Record) *EnvelopeIn {
	return &EnvelopeIn{
		Record: r,
		rd:     bytes.NewReader(r.Value),
	}
}

type EnvelopeIn struct {
	*kgo.Record
	rd *bytes.Reader
}

func (e EnvelopeIn) Read(p []byte) (n int, err error) {
	return e.rd.Read(p)
}

func (e EnvelopeIn) Headers() run.Headers {
	res := make(run.Headers, len(e.Record.Headers))
	for _, h := range e.Record.Headers {
		res[h.Key] = h.Value
	}
	return res
}
