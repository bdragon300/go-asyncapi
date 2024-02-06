package std

import (
	"bytes"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
	runRawSocket "github.com/bdragon300/go-asyncapi/run/rawsocket"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{payload: bytes.NewBuffer(nil)}
}

type EnvelopeOut struct {
	payload     *bytes.Buffer
	headers     run.Headers
	contentType string
	remoteAddr  net.Addr
}

func (e *EnvelopeOut) Write(p []byte) (n int, err error) {
	return e.payload.Write(p)
}

func (e *EnvelopeOut) ResetPayload() { // TODO: rename to Reset and include buffer into envelope. And envelopein as well
	e.payload.Reset()
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	e.headers = headers
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.contentType = contentType
}

func (e *EnvelopeOut) SetBindings(_ runRawSocket.MessageBindings) {}

func (e *EnvelopeOut) RecordStd() []byte {
	return e.payload.Bytes()
}

func (e *EnvelopeOut) SetRemoteAddr(addr net.Addr) {
	e.remoteAddr = addr
}

func (e *EnvelopeOut) RemoteAddr() net.Addr {
	return e.remoteAddr
}

func NewEnvelopeIn(msg []byte, addr net.Addr) *EnvelopeIn {
	return &EnvelopeIn{payload: bytes.NewReader(msg), remoteAddr: addr}
}

type EnvelopeIn struct {
	payload    *bytes.Reader
	remoteAddr net.Addr
}

func (e *EnvelopeIn) Read(p []byte) (n int, err error) {
	return e.payload.Read(p)
}

func (e *EnvelopeIn) Headers() run.Headers {
	return nil
}

func (e *EnvelopeIn) RemoteAddr() net.Addr {
	return e.remoteAddr
}
