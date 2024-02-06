package std

import (
	"bytes"

	"github.com/bdragon300/go-asyncapi/run"
	runTCP "github.com/bdragon300/go-asyncapi/run/tcp"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{payload: bytes.NewBuffer(nil)}
}

type EnvelopeOut struct {
	payload     *bytes.Buffer
	headers     run.Headers
	contentType string
}

func (e *EnvelopeOut) Write(p []byte) (n int, err error) {
	return e.payload.Write(p)
}

func (e *EnvelopeOut) ResetPayload() {
	e.payload.Reset()
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	e.headers = headers
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.contentType = contentType
}

func (e *EnvelopeOut) SetBindings(_ runTCP.MessageBindings) {}

func (e *EnvelopeOut) RecordStd() []byte {
	return e.payload.Bytes()
}

func NewEnvelopeIn(msg []byte) *EnvelopeIn {
	return &EnvelopeIn{payload: bytes.NewReader(msg)}
}

type EnvelopeIn struct {
	payload *bytes.Reader
}

func (e *EnvelopeIn) Read(p []byte) (n int, err error) {
	return e.payload.Read(p)
}

func (e *EnvelopeIn) Headers() run.Headers {
	return nil
}
