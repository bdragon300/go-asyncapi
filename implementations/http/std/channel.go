package std

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/bdragon300/go-asyncapi/run"
	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewChannel(bindings *runHttp.ChannelBindings, serverURL *url.URL, conn net.Conn, rw *bufio.ReadWriter) *Channel {
	res := Channel{
		Conn:       conn,
		ReadWriter: rw,
		bindings:   bindings,
		serverURL:  serverURL,
		items:      run.NewFanOut[runHttp.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancel(context.Background())
	go res.run()
	return &res
}

type ImplementationRecord interface {
	AsStdRecord() *http.Request
	Path() string
	// TODO: Bindings?
}

type Channel struct {
	net.Conn
	ReadWriter *bufio.ReadWriter

	bindings  *runHttp.ChannelBindings
	serverURL *url.URL
	items     *run.FanOut[runHttp.EnvelopeReader]
	ctx       context.Context
	cancel    context.CancelFunc
}

func (s Channel) Receive(ctx context.Context, cb func(envelope runHttp.EnvelopeReader)) error {
	el := s.items.Add(cb)
	defer s.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return errors.New("channel closed")
	}
}

func (s Channel) Send(ctx context.Context, envelopes ...runHttp.EnvelopeWriter) error {
	select {
	case <-s.ctx.Done():
		return errors.New("channel closed")
	default:
	}

	method := s.bindings.PublisherBindings.Method
	if method == "" {
		method = "GET"
	}

	for i, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		req := ir.AsStdRecord().WithContext(ctx)

		u := req.URL
		if u == nil {
			u = s.serverURL
		}
		u = u.JoinPath(ir.Path())
		req.URL = u

		req.Method = method
		err := req.Write(s.ReadWriter)
		if err != nil {
			return fmt.Errorf("envelope #%d: %w", i, err)
		}
	}
	return nil
}

func (s Channel) Close() error {
	s.cancel()
	return s.Conn.Close()
}

func (s Channel) run() {
	defer s.cancel()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			req, err := http.ReadRequest(s.ReadWriter.Reader)
			if err != nil {
				// TODO: error log
				continue
			}
			s.items.Put(NewEnvelopeIn(req))
		}
	}
}

