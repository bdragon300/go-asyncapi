package nethttp

import (
	"bufio"
	"net"
	"net/url"

	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewProducer(serverURL string, bindings *runHttp.ServerBindings) (*ProduceClient, error) {
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
	bindings  *runHttp.ServerBindings
	serverURL *url.URL
}

func (p ProduceClient) NewPublisher(channelName string, bindings *runHttp.ChannelBindings) (runHttp.Publisher, error) {
	port := p.serverURL.Port()
	if port == "" {
		port = "80"
	}
	netConn, err := net.Dial("tcp", net.JoinHostPort(p.serverURL.Hostname(), port))
	if err != nil {
		return nil, err
	}

	rw := bufio.NewReadWriter(bufio.NewReader(netConn), bufio.NewWriter(netConn))
	return NewConnection(bindings, p.serverURL.JoinPath(channelName), netConn, rw), nil
}
