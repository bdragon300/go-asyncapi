package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Pub
type Producer[B any, E EnvelopeWriter] interface {
	Publisher(channelName string, bindings *B) (Publisher[E], error)
}
type Publisher[E EnvelopeWriter] interface {
	Send(ctx context.Context, envelopes ...E) error
	Close() error
}
type EnvelopeWriter interface {
	io.Writer
	ResetPayload()
	SetHeaders(headers Headers)
	Protocol() Protocol
}

// Sub
type Consumer[B any, E EnvelopeReader] interface {
	Subscriber(channelName string, bindings *B) (Subscriber[E], error)
}
type Subscriber[E EnvelopeReader] interface {
	Receive(ctx context.Context, cb func(envelope E) error) error
	Close() error
}
type EnvelopeReader interface {
	io.Reader
	Headers() Headers
	Protocol() Protocol
}

type Headers map[string]any

func (h Headers) ToByteValues() map[string][]byte {
	res := make(map[string][]byte, len(h))
	for k, v := range h {
		switch tv := v.(type) {
		case []byte:
			res[k] = tv
		case string:
			res[k] = []byte(tv)
		default:
			b, err := json.Marshal(tv) // FIXME: use special util function for type conversion
			if err != nil {
				panic(fmt.Sprintf("Cannot marshal header value of type %T: %v", v, err))
			}
			res[k] = b
		}
	}

	return res
}

type Parameter interface {
	Name() string
	String() string
}
