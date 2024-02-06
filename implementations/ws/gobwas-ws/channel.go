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

func NewChannel(bindings *runWs.ChannelBindings, conn net.Conn, rw *bufio.ReadWriter) *Channel {
	res := Channel{
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
	Bytes() []byte
	OpCode() ws.OpCode
}

type Channel struct {
	net.Conn
	ReadWriter *bufio.ReadWriter

	bindings *runWs.ChannelBindings
	items    *run.FanOut[runWs.EnvelopeReader]
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s Channel) Receive(ctx context.Context, cb func(envelope runWs.EnvelopeReader)) error {
	el := s.items.Add(cb)
	defer s.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return errors.New("channel closed")
	}
}

func (s Channel) Send(_ context.Context, envelopes ...runWs.EnvelopeWriter) error {
	select {
	case <-s.ctx.Done():
		return errors.New("channel closed")
	default:
	}

	for i, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		msg := ir.Bytes()
		err := wsutil.WriteServerMessage(s.ReadWriter, ir.OpCode(), msg)
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
