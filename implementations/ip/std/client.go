package std

import (
	"context"
	"fmt"
	"net"
	"net/url"

	runIP "github.com/bdragon300/go-asyncapi/run/ip"
)

const DefaultMaxEnvelopeSize = 1024

func NewClient(localURL, remoteURL string) (*Client, error) {
	res := &Client{MaxEnvelopeSize: DefaultMaxEnvelopeSize}

	if localURL != "" {
		u, err := url.Parse(localURL)
		if err != nil {
			return nil, fmt.Errorf("parse localURL: %w", err)
		}
		res.localAddress = u.Hostname()
		res.localProtocolFamily = u.Scheme
		if p := u.Query().Get("proto"); p != "" {
			res.localProtocolFamily += ":" + p
		}
	}

	if remoteURL != "" {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return nil, fmt.Errorf("parse remoteURL: %w", err)
		}
		res.remoteAddress = u.Hostname()
	}

	return res, nil
}

type Client struct {
	Config              net.ListenConfig
	localAddress        string
	localProtocolFamily string
	remoteAddress       string

	// MaxEnvelopeSize is the maximum size of received envelopes. It should be set to the maximum
	// expected size of the IP datagram that can be received. If the size of the received datagram
	// exceeds this value, the datagram will be truncated and the rest of the data will be lost. By default, it is 1024.
	MaxEnvelopeSize int
}

func (c *Client) Subscriber(ctx context.Context, _ string, _ *runIP.ChannelBindings, _ *runIP.OperationBindings) (runIP.Subscriber, error) {
	return c.channel(ctx)
}

func (c *Client) Publisher(ctx context.Context, _ string, _ *runIP.ChannelBindings, _ *runIP.OperationBindings) (runIP.Publisher, error) {
	return c.channel(ctx)
}

func (c *Client) channel(ctx context.Context) (*Channel, error) {
	conn, err := c.Config.ListenPacket(ctx, c.localProtocolFamily, c.localAddress)
	if err != nil {
		return nil, err
	}

	var raddr net.Addr
	if c.remoteAddress != "" {
		raddr, err = net.ResolveIPAddr(c.localProtocolFamily, c.remoteAddress)
		if err != nil {
			return nil, fmt.Errorf("resolve remote address: %w", err)
		}
	}

	return NewChannel(conn.(*net.IPConn), c.MaxEnvelopeSize, raddr), nil
}
