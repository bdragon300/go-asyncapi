package std

import (
	"context"
	"fmt"
	"net"
	"net/url"

	runUDP "github.com/bdragon300/go-asyncapi/run/udp"
)

const DefaultMaxEnvelopeSize = 1024

func NewClient(localURL, remoteURL string) (*Client, error) {
	var la, ra, pf string

	if localURL != "" {
		u, err := url.Parse(localURL)
		if err != nil {
			return nil, fmt.Errorf("parse localURL: %w", err)
		}
		la = u.Host
		pf = u.Scheme
	}

	if remoteURL != "" {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return nil, fmt.Errorf("parse remoteURL: %w", err)
		}
		ra = u.Host
		pf = u.Scheme
	}

	return &Client{
		LocalAddress:         la,
		DefaultRemoteAddress: ra,
		MaxEnvelopeSize:      DefaultMaxEnvelopeSize,
		protocolFamily:       pf,
	}, nil
}

type Client struct {
	Config               net.ListenConfig
	LocalAddress         string
	DefaultRemoteAddress string

	// MaxEnvelopeSize is the maximum size of received envelopes. It should be set to the maximum
	// expected size of the UDP datagram that can be received. If the size of the received datagram
	// exceeds this value, the datagram will be truncated and the rest of the data will be lost. By default, it is 1024.
	MaxEnvelopeSize int

	protocolFamily string
}

func (c Client) Publisher(ctx context.Context, _ string, _ *runUDP.ChannelBindings) (runUDP.Publisher, error) {
	return c.channel(ctx)
}

func (c Client) Subscriber(ctx context.Context, _ string, _ *runUDP.ChannelBindings) (runUDP.Subscriber, error) {
	return c.channel(ctx)
}

func (c *Client) channel(ctx context.Context) (*Channel, error) {
	conn, err := c.Config.ListenPacket(ctx, c.protocolFamily, c.LocalAddress)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveUDPAddr(c.protocolFamily, c.DefaultRemoteAddress)
	if err != nil {
		return nil, fmt.Errorf("resolve remote address: %w", err)
	}

	return NewChannel(conn.(*net.UDPConn), c.MaxEnvelopeSize, addr), nil
}

