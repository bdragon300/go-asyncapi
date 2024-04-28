package std

import (
	"context"
	"errors"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
	runRawSocket "github.com/bdragon300/go-asyncapi/run/rawsocket"
)

func NewChannel(conn *net.IPConn, bufferSize int, defaultRemoteAddress net.Addr) *Channel {
	res := Channel{
		IPConn:               conn,
		defaultRemoteAddress: defaultRemoteAddress,
		bufferSize:           bufferSize,
		items:                run.NewFanOut[runRawSocket.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancelCause(context.Background())
	go res.run() // TODO: run once Receive is called (everywhere do this)
	return &res
}

type Channel struct {
	*net.IPConn

	defaultRemoteAddress net.Addr
	bufferSize           int // Including IP headers
	items                *run.FanOut[runRawSocket.EnvelopeReader]
	ctx                  context.Context
	cancel               context.CancelCauseFunc
}

type ImplementationRecord interface {
	Bytes() []byte
	RemoteAddr() net.Addr
}

func (c *Channel) Send(_ context.Context, envelopes ...runRawSocket.EnvelopeWriter) error {
	for _, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		addr := ir.RemoteAddr()
		if addr == nil {
			addr = c.defaultRemoteAddress
		}
		if _, err := c.IPConn.WriteTo(ir.Bytes(), addr); err != nil {
			return err
		}
	}

	return nil
}

func (c *Channel) Receive(ctx context.Context, cb func(envelope runRawSocket.EnvelopeReader)) error {
	el := c.items.Add(cb)
	defer c.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return context.Cause(c.ctx)
	}
}

func (c *Channel) Close() error {
	c.cancel(errors.New("close channel"))
	return c.IPConn.Close()
}

func (c *Channel) run() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		buf := make([]byte, c.bufferSize) // TODO: sync.Pool
		n, addr, err := c.IPConn.ReadFrom(buf)
		if err != nil {
			c.cancel(err)
			return
		}
		c.items.Put(func() runRawSocket.EnvelopeReader { return NewEnvelopeIn(buf[:n], addr) })
	}
}
