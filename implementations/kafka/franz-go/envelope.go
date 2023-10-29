package franzgo

import (
	"bytes"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
	"github.com/bdragon300/asyncapi-codegen-go/pkg/run/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{
		payload: bytes.NewBuffer(make([]byte, 0)),
	}
}

type EnvelopeOut struct {
	*kgo.Record
	payload         *bytes.Buffer
	messageBindings kafka.MessageBindings
}

func (e *EnvelopeOut) Write(p []byte) (n int, err error) {
	return e.payload.Write(p)
}

func (e *EnvelopeOut) ResetPayload() {
	e.payload.Reset()
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	for k, v := range headers.ToByteValues() {
		e.Record.Headers = append(e.Record.Headers, kgo.RecordHeader{Key: k, Value: v})
	}
}

func (e *EnvelopeOut) Protocol() run.Protocol {
	return run.ProtocolKafka
}

func (e *EnvelopeOut) SetBindings(bindings kafka.MessageBindings) {
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

func (e EnvelopeIn) Protocol() run.Protocol {
	return run.ProtocolKafka
}