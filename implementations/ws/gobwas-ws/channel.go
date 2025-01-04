package gobwasws

import (
	"context"
	"fmt"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
	runWs "github.com/bdragon300/go-asyncapi/run/ws"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func NewChannel(chb *runWs.ChannelBindings, opb *runWs.OperationBindings, conn net.Conn, clientSide bool) *Channel {
	res := Channel{
		Conn:            conn,
		clientSide:      clientSide,
		channelBindings: chb,
		operationBindings: opb,
		items:           run.NewFanOut[runWs.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancelCause(context.Background())
	go res.run()
	return &res
}

type ImplementationRecord interface {
	Bytes() []byte
	OpCode() ws.OpCode
}

type Channel struct {
	net.Conn

	// clientSide determines if this channel is a client-side or a server-side.
	// To prevent cache spoofing attack, the client-side application must additionally mask the payload in
	// outgoing websocket frames, whereas the server-side code must unmask the payload back in incoming frames
	// https://www.rfc-editor.org/rfc/rfc6455#section-5.3
	clientSide      bool
	channelBindings *runWs.ChannelBindings
	operationBindings *runWs.OperationBindings
	items           *run.FanOut[runWs.EnvelopeReader]
	ctx        context.Context
	cancel     context.CancelCauseFunc
}

func (s Channel) Receive(ctx context.Context, cb func(envelope runWs.EnvelopeReader)) error {
	el := s.items.Add(cb)
	defer s.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return context.Cause(s.ctx)
	}
}

func (s Channel) Send(ctx context.Context, envelopes ...runWs.EnvelopeWriter) error {
	for i, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		msg := ir.Bytes()

		select {
		case <-s.ctx.Done():
			return context.Cause(s.ctx)
		case <-ctx.Done():
			return ctx.Err()
		default:
			var err error
			if s.clientSide {
				err = wsutil.WriteClientMessage(s.Conn, ir.OpCode(), msg)
			} else {
				err = wsutil.WriteServerMessage(s.Conn, ir.OpCode(), msg)
			}
			if err != nil {
				return fmt.Errorf("envelope #%d: %w", i, err)
			}
		}
	}
	return nil
}

func (s Channel) Close() error {
	s.cancel(nil)
	return s.Conn.Close()
}

func (s Channel) run() {
	var err error
	defer func() {
		if err != nil {
			s.cancel(err)
		}
	}()

	for {
		var msgs []wsutil.Message
		if s.clientSide {
			msgs, err = wsutil.ReadServerMessage(s.Conn, nil)
		} else {
			msgs, err = wsutil.ReadClientMessage(s.Conn, nil)
		}
		if err != nil {
			return
		}
		for _, msg := range msgs {
			select {
			case <-s.ctx.Done():
				return
			default:
				s.items.Put(func() runWs.EnvelopeReader { return NewEnvelopeIn(msg) })
			}
		}
	}
}
