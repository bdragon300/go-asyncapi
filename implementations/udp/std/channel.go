package std

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/bdragon300/go-asyncapi/run"
	runUDP "github.com/bdragon300/go-asyncapi/run/udp"
)

func NewChannel(conn *net.UDPConn, bufferSize int, defaultRemoteAddress net.Addr) *Channel {
	res := Channel{
		UDPConn:              conn,
		defaultRemoteAddress: defaultRemoteAddress,
		bufferSize:           bufferSize,
		items:                run.NewFanOut[runUDP.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancelCause(context.Background())
	go res.run() // TODO: run once Receive is called (everywhere do this)
	return &res
}

type Channel struct {
	*net.UDPConn

	defaultRemoteAddress net.Addr
	bufferSize           int
	items                *run.FanOut[runUDP.EnvelopeReader]
	ctx                  context.Context
	cancel               context.CancelCauseFunc
}

type ImplementationRecord interface {
	RecordStd() []byte
	RemoteAddr() net.Addr
}

func (c *Channel) Send(_ context.Context, envelopes ...runUDP.EnvelopeWriter) error {
	for _, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		addr := ir.RemoteAddr()
		if addr == nil {
			addr = c.defaultRemoteAddress
		}
		if _, err := c.UDPConn.WriteTo(ir.RecordStd(), addr); err != nil {
			return err
		}
	}
	return nil
}

func (c *Channel) Receive(ctx context.Context, cb func(envelope runUDP.EnvelopeReader)) error {
	el := c.items.Add(cb)
	defer c.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return fmt.Errorf("channel closed: %w", c.ctx.Err())
	}
}

func (c *Channel) Close() error {
	c.cancel(errors.New("close channel"))
	return c.UDPConn.Close()
}

func (c *Channel) run() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		buf := make([]byte, c.bufferSize) // TODO: sync.Pool
		n, addr, err := c.UDPConn.ReadFrom(buf)
		if err != nil {
			c.cancel(err)
			return
		}
		c.items.Put(NewEnvelopeIn(buf[:n], addr))
	}
}
