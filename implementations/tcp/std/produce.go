package std

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"

	runTCP "github.com/bdragon300/go-asyncapi/run/tcp"
)

func NewProducer(serverURL string, bindings *runTCP.ChannelBindings, protocolVersion string) (*ProduceClient, error) {
	if protocolVersion != "" && protocolVersion != "4" && protocolVersion != "6" {
		return nil, fmt.Errorf("invalid protocol version: %s", protocolVersion)
	}
	protocolFamily := ProtocolFamily + protocolVersion

	d := net.Dialer{}
	if bindings != nil {
		address := net.JoinHostPort(bindings.LocalAddress, strconv.Itoa(bindings.LocalPort))
		la, err := net.ResolveTCPAddr(protocolFamily, address)
		if err != nil {
			return nil, err
		}
		d.LocalAddr = la
	}
	return &ProduceClient{
		Dialer:          d,
		Address:         serverURL,
		MaxEnvelopeSize: DefaultMaxEnvelopeSize,
		protocolFamily:  protocolFamily,
	}, nil
}

type ProduceClient struct {
	net.Dialer
	Address         string
	Scanner         *bufio.Scanner
	MaxEnvelopeSize int

	protocolFamily string
}

func (p ProduceClient) Publisher(ctx context.Context, _ string, _ *runTCP.ChannelBindings) (runTCP.Publisher, error) {
	conn, err := p.DialContext(ctx, p.protocolFamily, p.Address)
	if err != nil {
		return nil, err
	}

	return NewChannel(conn.(*net.TCPConn), p.Scanner, p.MaxEnvelopeSize), nil
}
