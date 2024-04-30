package std

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/url"

	runTCP "github.com/bdragon300/go-asyncapi/run/tcp"
)

const DefaultMaxEnvelopeSize = 1024

type Decoder interface {
	Decode(v any) error
}

func NewConsumer(listenURL string) (*ConsumeClient, error) {
	u, err := url.Parse(listenURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "tcp" && u.Scheme != "tcp4" && u.Scheme != "tcp6" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}
	listener, err := net.Listen(u.Scheme, u.Host)

	return &ConsumeClient{
		TCPListener:     listener.(*net.TCPListener),
		MaxEnvelopeSize: DefaultMaxEnvelopeSize,
	}, err
}

type ConsumeClient struct {
	*net.TCPListener
	// Scanner splits the incoming data into Envelopes. If equal to nil, the data will
	// be split on chunks of MaxEnvelopeSize bytes, which is equal to bufio.MaxScanTokenSize by default.
	Scanner         *bufio.Scanner
	MaxEnvelopeSize int
}

func (c *ConsumeClient) Subscriber(_ context.Context, _ string, _ *runTCP.ChannelBindings) (runTCP.Subscriber, error) {
	conn, err := c.AcceptTCP()
	if err != nil {
		return nil, err
	}

	return NewChannel(conn, c.Scanner, c.MaxEnvelopeSize), nil
}
