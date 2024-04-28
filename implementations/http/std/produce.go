package std

import (
	"context"
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

func (p ProduceClient) Publisher(_ context.Context, channelName string, bindings *runHttp.ChannelBindings) (runHttp.Publisher, error) {
	return NewPublisher(bindings, p.serverURL.JoinPath(channelName)), nil
}
