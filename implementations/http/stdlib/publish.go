package stdlib

import (
	"context"
	"net/http"
	"net/url"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
	runHttp "github.com/bdragon300/asyncapi-codegen-go/pkg/run/http"
)

func NewProducer(serverURL string, bindings *runHttp.ServerBindings) (*Producer, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	return &Producer{
		URL:      u,
		Bindings: bindings,
	}, nil
}

type Producer struct {
	URL      *url.URL
	Bindings *runHttp.ServerBindings
}

func (p Producer) Publisher(channelName string, bindings *runHttp.ChannelBindings) (run.Publisher[*EnvelopeOut], error) {
	return &Publisher{
		channelName:    channelName,
		url:            p.URL,
		bindings:       bindings,
		NewRequest:     NewRequest,
		HandleResponse: nil,
	}, nil
}

type Publisher struct {
	channelName    string
	url            *url.URL
	bindings       *runHttp.ChannelBindings
	NewRequest     func(ctx context.Context, method, url string, e *EnvelopeOut) (*http.Request, error)
	HandleResponse func(r *http.Response) error
}

func (p Publisher) Send(ctx context.Context, envelopes ...*EnvelopeOut) error {
	pool := run.NewErrorPool()
	method := p.bindings.PublisherBindings.Method
	if method == "" {
		method = "GET"
	}

	for _, e := range envelopes {
		e := e
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

func (p Publisher) Close() error {
	return nil
}
