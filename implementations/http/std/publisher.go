package std

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewPublisher(bindings *runHttp.ChannelBindings, channelURL *url.URL) *Publisher {
	return &Publisher{
		Client:     http.DefaultClient,
		channelURL: channelURL,
		bindings:   bindings,
	}
}

type Publisher struct {
	Client     *http.Client
	channelURL *url.URL
	bindings   *runHttp.ChannelBindings
}

type ImplementationRecord interface {
	AsStdRecord() *http.Request
	// TODO: Bindings?
}

func (p Publisher) Send(ctx context.Context, envelopes ...runHttp.EnvelopeWriter) error {
	method := "GET"
	if p.bindings != nil && p.bindings.PublisherBindings.Method != "" {
		method = p.bindings.PublisherBindings.Method
	}

	for i, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		req := ir.AsStdRecord().WithContext(ctx)
		if req.URL == nil {
			req.URL = p.channelURL
		}

		req.Method = method
		if _, err := p.Client.Do(req); err != nil { // TODO: implement request-response interface in channels
			return fmt.Errorf("envelope #%d: %w", i, err)
		}
	}
	return nil
}

func (p Publisher) Close() error {
	return nil
}
