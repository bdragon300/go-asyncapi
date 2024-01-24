package gobwasws

import (
	"bufio"
	"context"
	"fmt"
	"net/url"

	runWs "github.com/bdragon300/go-asyncapi/run/ws"
	"github.com/gobwas/ws"
)

func NewProducer(serverURL string, bindings *runWs.ServerBindings) (*ProduceClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	return &ProduceClient{
		bindings:  bindings,
		serverURL: u,
	}, nil
}

type ProduceClient struct {
	bindings  *runWs.ServerBindings
	serverURL *url.URL
}

func (p ProduceClient) NewPublisher(channelName string, bindings *runWs.ChannelBindings) (runWs.Publisher, error) {
	if bindings.Method != "" && bindings.Method != "GET" {
		return nil, fmt.Errorf("unsupported method %s", bindings.Method)
	}
	u := p.serverURL.JoinPath(channelName)
	netConn, rd, _, err := ws.Dial(context.Background(), u.String())
	if err != nil {
		return nil, err
	}

	rw := bufio.NewReadWriter(rd, bufio.NewWriter(netConn))
	return NewConnection(bindings, netConn, rw), nil
}
