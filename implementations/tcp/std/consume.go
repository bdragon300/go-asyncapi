package std

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"

	runTCP "github.com/bdragon300/go-asyncapi/run/tcp"
)

const (
	ProtocolFamily         = "tcp"
	DefaultMaxEnvelopeSize = 1024
)

type Decoder interface {
	Decode(v any) error
}

func NewConsumer(bindings *runTCP.ChannelBindings, protocolVersion string) (*ConsumeClient, error) {
	if protocolVersion != "" && protocolVersion != "4" && protocolVersion != "6" {
		return nil, fmt.Errorf("invalid protocol version: %s", protocolVersion)
	}

	listenAddress := ""
	if bindings != nil {
		listenAddress = net.JoinHostPort(bindings.LocalAddress, strconv.Itoa(bindings.LocalPort))
	}

	listener, err := net.Listen(ProtocolFamily+protocolVersion, listenAddress)

	return &ConsumeClient{
		TCPListener:     listener.(*net.TCPListener),
		MaxEnvelopeSize: DefaultMaxEnvelopeSize,
	}, err
}

type ConsumeClient struct {
	*net.TCPListener
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
