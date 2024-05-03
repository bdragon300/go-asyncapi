package std

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/bdragon300/go-asyncapi/run"
	runIP "github.com/bdragon300/go-asyncapi/run/ip"
)

func NewChannel(conn *net.IPConn, bufferSize int, remoteAddress net.Addr) *Channel {
	res := Channel{
		IPConn:        conn,
		remoteAddress: remoteAddress,
		bufferSize:    bufferSize,
		items:         run.NewFanOut[runIP.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancelCause(context.Background())
	go res.run() // TODO: run once Receive is called (everywhere do this)
	return &res
}

type Channel struct {
	*net.IPConn

	remoteAddress    net.Addr // Ignored if includeIPHeaders is true, see TCP/IP stack docs
	bufferSize       int
	includeIPHeaders bool
	items            *run.FanOut[runIP.EnvelopeReader]
	ctx              context.Context
	cancel           context.CancelCauseFunc
}

type ImplementationRecord interface {
	Bytes() []byte
	HeaderBytes() ([]byte, error)
}

func (c *Channel) Send(_ context.Context, envelopes ...runIP.EnvelopeWriter) error {
	for i, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		msg := ir.Bytes()
		if c.includeIPHeaders {
			headers, err := ir.HeaderBytes()
			if err != nil {
				return fmt.Errorf("header bytes in envelope #%d: %w", i, err)
			}
			msg = append(headers, msg...)
		}

		if _, err := c.IPConn.WriteTo(msg, c.remoteAddress); err != nil {
			return err
		}
	}

	return nil
}

func (c *Channel) Receive(ctx context.Context, cb func(envelope runIP.EnvelopeReader)) error {
	el := c.items.Add(cb)
	defer c.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ctx.Done():
		return context.Cause(c.ctx)
	}
}

func (c *Channel) SetIncludeIPHeaders(include bool) error {
	c.includeIPHeaders = include

	sock, err := c.IPConn.SyscallConn()
	if err != nil {
		return fmt.Errorf("get syscall conn: %w", err)
	}
	var includeInt int
	if include {
		includeInt = 1
	}
	return sock.Control(func(fd uintptr) {
		// FIXME: fix this in more elegant way, we should know here the IP version instead of guessing
		err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, includeInt)
		if err != nil {
			if err2 := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, unix.IPV6_HDRINCL, includeInt); err2 != nil {
				panic(errors.Join(err, err2).Error())
			}
		}
	})
}

func (c *Channel) Close() error {
	c.cancel(nil)
	return c.IPConn.Close()
}

func (c *Channel) run() {
	var ok bool

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		buf := make([]byte, c.bufferSize) // TODO: sync.Pool
		n, err := c.IPConn.Read(buf)
		if err != nil {
			c.cancel(err)
			return
		}

		var po, ver int
		if po, ok = ipv4PayloadOffset(n, buf); ok {
			ver = 4
		} else if po, ok = ipv6PayloadOffset(n, buf); ok {
			ver = 6
		}
		c.items.Put(func() runIP.EnvelopeReader { return NewEnvelopeIn(buf[:po], buf[po:n], ver) })
	}
}

func ipv4PayloadOffset(n int, b []byte) (int, bool) {
	if len(b) < 20 {
		return 0, false
	}
	// Check an IP version
	if b[0]>>4 != 4 {
		return 0, false
	}
	ihl := int(b[0]&0x0f) << 2 // Internet Header Length
	if 20 > ihl || ihl > len(b) {
		return 0, true
	}
	if ihl > n {
		return n, true
	}
	return ihl, true
}

func ipv6PayloadOffset(n int, b []byte) (int, bool) {
	if len(b) < 40 {
		return 0, false
	}
	// Check an IP version
	if b[0]>>4 != 6 {
		return 0, false
	}
	if n < 40 {
		return n, true
	}
	return 40, true
}
