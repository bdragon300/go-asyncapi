package std

import (
	"context"
	"fmt"
	"net"

	runRawSocket "github.com/bdragon300/go-asyncapi/run/rawsocket"
)

const (
	ProtocolFamily         = "ip"
	DefaultLocalAddress    = "localhost"
	DefaultMaxEnvelopeSize = 1024
)

func NewConsumer(bindings *runRawSocket.ChannelBindings, protocolVersion string) (*Client, error) {
	return NewProducer("", bindings, protocolVersion)
}

func NewProducer(serverURL string, bindings *runRawSocket.ChannelBindings, protocolVersion string) (*Client, error) {
	if protocolVersion != "" && protocolVersion != "4" && protocolVersion != "6" {
		return nil, fmt.Errorf("invalid protocol version: %s", protocolVersion)
	}

	localAddress := DefaultLocalAddress
	if bindings != nil && bindings.LocalAddress != "" {
		localAddress = bindings.LocalAddress
	}
	return &Client{
		LocalAddress:         localAddress,
		DefaultRemoteAddress: serverURL,
		MaxEnvelopeSize:      DefaultMaxEnvelopeSize,
		protocolFamily:       ProtocolFamily + protocolVersion,
	}, nil
}

type Client struct {
	Config               net.ListenConfig
	LocalAddress         string
	DefaultRemoteAddress string
	// MaxEnvelopeSize is the maximum size of received envelopes. It should be set to the maximum
	// expected size of the IP datagram that can be received. If the size of the received datagram
	// exceeds this value, the datagram will be truncated and the rest of the data will be lost. By default, it is 1024.
	MaxEnvelopeSize int

	protocolFamily string
}

func (c *Client) Subscriber(ctx context.Context, _ string, _ *runRawSocket.ChannelBindings) (runRawSocket.Subscriber, error) {
	return c.channel(ctx)
}

func (c *Client) Publisher(ctx context.Context, _ string, _ *runRawSocket.ChannelBindings) (runRawSocket.Publisher, error) {
	return c.channel(ctx)
}

func (c *Client) channel(ctx context.Context) (*Channel, error) {
	conn, err := c.Config.ListenPacket(ctx, c.protocolFamily, c.LocalAddress)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveIPAddr(c.protocolFamily, c.DefaultRemoteAddress)
	if err != nil {
		return nil, fmt.Errorf("resolve remote address: %w", err)
	}

	return NewChannel(conn.(*net.IPConn), c.MaxEnvelopeSize, addr), nil
}