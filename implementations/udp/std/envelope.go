package std

import (
	"bytes"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
	runUDP "github.com/bdragon300/go-asyncapi/run/udp"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{Buffer: bytes.NewBuffer(nil)}
}

type EnvelopeOut struct {
	*bytes.Buffer
	headers     run.Headers
	contentType string
	remoteAddr  net.Addr
}

func (e *EnvelopeOut) ResetPayload() { // TODO: rename to Reset
	e.Buffer.Reset()
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	e.headers = headers
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.contentType = contentType
}

func (e *EnvelopeOut) SetBindings(_ runUDP.MessageBindings) {}

func (e *EnvelopeOut) RecordStd() []byte {
	return e.Buffer.Bytes()
}

func (e *EnvelopeOut) SetRemoteAddr(addr net.Addr) {
	e.remoteAddr = addr
}

func (e *EnvelopeOut) RemoteAddr() net.Addr {
	return e.remoteAddr
}

func NewEnvelopeIn(msg []byte, addr net.Addr) *EnvelopeIn {
	return &EnvelopeIn{Reader: bytes.NewReader(msg), remoteAddr: addr}
}

type EnvelopeIn struct {
	*bytes.Reader
	remoteAddr net.Addr
}

func (e *EnvelopeIn) Headers() run.Headers {
	return nil
}

func (e *EnvelopeIn) RemoteAddr() net.Addr {
	return e.remoteAddr
}
