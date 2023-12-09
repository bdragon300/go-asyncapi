package stdlib

import (
	"context"
	"net/http"
	"net/url"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
	runHttp "github.com/bdragon300/asyncapi-codegen-go/pkg/run/http"
)

func NewProducer(serverURL string, bindings *runHttp.ServerBindings) (*ProduceClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	return &ProduceClient{
		URL:      u,
		Bindings: bindings,
	}, nil
}

type ProduceClient struct {
	URL      *url.URL
	Bindings *runHttp.ServerBindings
}

func (p ProduceClient) Publisher(channelName string, bindings *runHttp.ChannelBindings) (runHttp.Publisher, error) {
	return &PublishClient{
		channelName:    channelName,
		url:            p.URL,
		bindings:       bindings,
		NewRequest:     NewRequest,
		HandleResponse: nil,
	}, nil
}

type PublishClient struct {
	channelName    string
	url            *url.URL
	bindings       *runHttp.ChannelBindings
	NewRequest     func(ctx context.Context, method, url string, e *EnvelopeOut) (*http.Request, error)
	HandleResponse func(r *http.Response) error
}

func (p PublishClient) Send(ctx context.Context, envelopes ...runHttp.EnvelopeWriter) error {
	pool := run.NewErrorPool()
	method := p.bindings.PublisherBindings.Method
	if method == "" {
		method = "GET"
	}

	for _, envelope := range envelopes {
		e := envelope.(*EnvelopeOut)
		u := e.URL.JoinPath(p.channelName, e.path)
		pool.Go(func() error {
			req, err := p.NewRequest(ctx, method, u.String(), e)
			if err != nil {
				return err
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			if p.HandleResponse != nil {
				return p.HandleResponse(resp)
			}
			return nil
		})
	}
	return pool.Wait()
}

func (p PublishClient) Close() error {
	return nil
}

func NewRequest(ctx context.Context, method, url string, e *EnvelopeOut) (*http.Request, error) {
	var body *EnvelopeOut
	if e.body.Len() > 0 {
		body = e
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if e.Proto != "" {
		req.Proto = e.Proto
	}
	if e.ProtoMajor != 0 {
		req.ProtoMajor = e.ProtoMajor
		req.ProtoMinor = e.ProtoMinor
	}
	if e.Header != nil {
		req.Header = e.Header
	}
	if e.ContentLength > 0 {
		req.ContentLength = e.ContentLength
	}
	req.TransferEncoding = e.TransferEncoding
	req.Close = e.Close
	if e.Host != "" {
		req.Host = e.Host
	}
	req.Form = e.Form
	req.PostForm = e.PostForm
	req.MultipartForm = e.MultipartForm
	req.Trailer = e.Trailer
	req.Response = e.Response

	return req, nil
}
