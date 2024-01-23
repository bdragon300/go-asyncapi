package gobwasws

import (
	"bytes"

	"github.com/bdragon300/go-asyncapi/run"
	runWs "github.com/bdragon300/go-asyncapi/run/ws"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{payload: bytes.NewBuffer(make([]byte, 0))}
}

type EnvelopeOut struct {
	opCode      ws.OpCode
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

func (e *EnvelopeOut) SetBindings(_ runWs.MessageBindings) {}

func (e *EnvelopeOut) SetOpCode(opCode byte) {
	e.opCode = ws.OpCode(opCode)
}

func NewEnvelopeIn(msg wsutil.Message) *EnvelopeIn {
	return &EnvelopeIn{Message: msg, reader: bytes.NewReader(msg.Payload)}
}

type EnvelopeIn struct {
	wsutil.Message
	reader *bytes.Reader
}

func (e *EnvelopeIn) Read(p []byte) (n int, err error) {
	return e.reader.Read(p)
}

func (e *EnvelopeIn) Headers() run.Headers {
	return nil
}
