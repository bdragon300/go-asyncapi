package stdhttp

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
	runHttp "github.com/bdragon300/asyncapi-codegen-go/pkg/run/http"
)

func NewEnvelopeOut() *EnvelopeOut {
	return &EnvelopeOut{
		body: bytes.NewBuffer(make([]byte, 0)),
	}
}

type EnvelopeOut struct {
	http.Request    // Client request stub, just for setting the options of a real request
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
			panic(fmt.Sprintf("Header values must be string or []string, got: %T", value))
		}
	}
}

func (e *EnvelopeOut) SetContentType(contentType string) {
	e.Header.Set("Content-Type", contentType)
}

func (e *EnvelopeOut) Protocol() run.Protocol {
	return run.ProtocolHTTP
}

func (e *EnvelopeOut) SetBindings(bindings runHttp.MessageBindings) {
	e.messageBindings = bindings
}

func (e *EnvelopeOut) SetPath(path string) {
	e.path = path
}

func NewEnvelopeIn(req *http.Request, w http.ResponseWriter) *EnvelopeIn {
	return &EnvelopeIn{
		Request:        req,
		responseWriter: w,
	}
}

type EnvelopeIn struct {
	*http.Request
	responseWriter http.ResponseWriter
}

func (e *EnvelopeIn) Read(p []byte) (n int, err error) {
	return e.Request.Body.Read(p)
}

func (e *EnvelopeIn) Headers() run.Headers {
	res := make(run.Headers)
	for name, val := range e.Request.Header {
		res[name] = val
	}
	return res
}

func (e *EnvelopeIn) Protocol() run.Protocol {
	return run.ProtocolHTTP
}

func (e *EnvelopeIn) ResponseWriter() http.ResponseWriter {
	return e.responseWriter
}

func (e *EnvelopeIn) RespondError(code int, error string) {
	http.Error(e.responseWriter, error, code)
}
