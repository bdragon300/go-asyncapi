package std

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bdragon300/go-asyncapi/run"
	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewEnvelopeOut() *EnvelopeOut {
	req := &http.Request{ // Taken from http.NewRequestWithContext
		Method:     "GET",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	return &EnvelopeOut{
		Request: req.WithContext(context.Background()),
		body:    bytes.NewBuffer(make([]byte, 0)),
	}
}

type EnvelopeOut struct {
	*http.Request
	messageBindings runHttp.MessageBindings
	body            *bytes.Buffer
	path            string
}

func (e *EnvelopeOut) Write(p []byte) (n int, err error) {
	return e.body.Write(p)
}

func (e *EnvelopeOut) Read(p []byte) (n int, err error) {
	return e.body.Read(p)
}

func (e *EnvelopeOut) ResetPayload() {
	e.body.Reset()
}

func (e *EnvelopeOut) SetHeaders(headers run.Headers) {
	for name, value := range headers {
		switch v := value.(type) {
		case string:
			e.Header.Set(name, v)
		case []string:
			e.Header.Del(name)
			for _, item := range v {
				e.Header.Add(name, item)
			}
		default:
			panic(fmt.Sprintf("Header value must be string or []string, got: %T", value))
		}
	}
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.Header.Set("Content-Type", contentType)
}

func (e *EnvelopeOut) SetBindings(bindings runHttp.MessageBindings) {
	e.messageBindings = bindings
}

func (e *EnvelopeOut) SetPath(path string) {
	e.path = path
}

func (e *EnvelopeOut) AsStdRecord() *http.Request {
	reqCopy := *e.Request
	reqCopy.GetBody = func() (io.ReadCloser, error) {
		snapshot := e.body.Bytes()
		return io.NopCloser(bytes.NewReader(snapshot)), nil
	}
	return &reqCopy
}

func (e *EnvelopeOut) Path() string {
	return e.path
}

func NewEnvelopeIn(req *http.Request) *EnvelopeIn {
	return &EnvelopeIn{Request: req}
}

type EnvelopeIn struct {
	*http.Request
}

func (e *EnvelopeIn) Read(p []byte) (n int, err error) {
	n, err = e.Request.Body.Read(p)
	if errors.Is(err, io.EOF) {
		_ = e.Request.Body.Close()
	}
	return
}

func (e *EnvelopeIn) Headers() run.Headers {
	res := make(run.Headers)
	for name, val := range e.Request.Header {
		res[name] = val
	}
	return res
}
