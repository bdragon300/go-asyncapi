package gobwasws

import (
	"context"
	"net/url"

	runWs "github.com/bdragon300/go-asyncapi/run/ws"
	"github.com/gobwas/ws"
)

func NewProducer(bindings *runWs.ServerBindings, serverURL *url.URL) (*ProduceClient, error) {
	return &ProduceClient{
		bindings:  bindings,
		serverURL: serverURL,
	}, nil
}

type ProduceClient struct {
	bindings  *runWs.ServerBindings
	serverURL *url.URL
}

func (p ProduceClient) NewPublisher(channelName string, bindings *runWs.ChannelBindings) (runWs.Publisher, error) {
	conn, _, _, err := ws.Dial(context.Background(), p.serverURL.String())
	if err != nil {
		return nil, err
	}

	return NewConnection(bindings, channelName, conn), nil
}
