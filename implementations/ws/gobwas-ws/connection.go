package gobwasws

import (
	"context"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
	"github.com/bdragon300/go-asyncapi/run/ws"
	ws2 "github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func NewConnection(bindings *ws.ChannelBindings, channelName string, conn net.Conn) *Connection {
	res := Connection{
		Conn:        conn,
		bindings:    bindings,
		channelName: channelName,
		items:       run.NewFanOut[ws.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancel(context.Background())
	go res.run()
	return &res
}

type ImplementationRecord interface {
	RecordGobwasWS() []byte
}

type Connection struct {
	net.Conn
	bindings    *ws.ChannelBindings
	channelName string
	items       *run.FanOut[ws.EnvelopeReader]
	ctx         context.Context
	cancel      context.CancelFunc
}

func (s Connection) Receive(ctx context.Context, cb func(envelope ws.EnvelopeReader) error) error {
	el := s.items.Add(cb)
	defer s.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return nil
	}
}

func (s Connection) Send(_ context.Context, envelopes ...ws.EnvelopeWriter) error {
	for _, envelope := range envelopes {
		msg := envelope.(ImplementationRecord).RecordGobwasWS()
		err := wsutil.WriteServerMessage(s.Conn, ws2.OpText, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Connection) Close() error {
	s.cancel()
	return s.Conn.Close()
}

func (s Connection) run() {
	for {
		// TODO: error log
		msgs, _ := wsutil.ReadClientMessage(s.Conn, nil)
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
