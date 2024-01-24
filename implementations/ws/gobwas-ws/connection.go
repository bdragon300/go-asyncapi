package gobwasws

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
	runWs "github.com/bdragon300/go-asyncapi/run/ws"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func NewConnection(bindings *runWs.ChannelBindings, conn net.Conn, rw *bufio.ReadWriter) *Connection {
	res := Connection{
		Conn:       conn,
		ReadWriter: rw,
		bindings:   bindings,
		items:      run.NewFanOut[runWs.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancel(context.Background())
	go res.run()
	return &res
}

type ImplementationRecord interface {
	RecordGobwasWS() []byte
	OpCode() ws.OpCode
}

type Connection struct {
	net.Conn
	ReadWriter *bufio.ReadWriter

	bindings *runWs.ChannelBindings
	items    *run.FanOut[runWs.EnvelopeReader]
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s Connection) Receive(ctx context.Context, cb func(envelope runWs.EnvelopeReader) error) error {
	el := s.items.Add(cb)
	defer s.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return errors.New("connection closed")
	}
}

func (s Connection) Send(_ context.Context, envelopes ...runWs.EnvelopeWriter) error {
	select {
	case <-s.ctx.Done():
		return errors.New("connection closed")
	default:
	}

	for i, envelope := range envelopes {
		msg := envelope.(ImplementationRecord).RecordGobwasWS()
		err := wsutil.WriteServerMessage(s.ReadWriter, ws.OpText, msg) // TODO: OpBinary?
		if err != nil {
			return fmt.Errorf("envelope #%d: %w", i, err)
		}
	}
	return nil
}

func (s Connection) Close() error {
	s.cancel()
	return s.Conn.Close()
}

func (s Connection) run() {
	defer s.cancel()

	for {
		// TODO: error log
		msgs, _ := wsutil.ReadClientMessage(s.ReadWriter, nil)
		for _, msg := range msgs {
			select {
			case <-s.ctx.Done():
				return
			default:
				s.items.Put(NewEnvelopeIn(msg))
			}
		}
	}
}
