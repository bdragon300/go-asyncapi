package stdhttp

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/bdragon300/go-asyncapi/run"
	runHttp "github.com/bdragon300/go-asyncapi/run/http"
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

type ImplementationRecord interface {
	RecordNetHTTP() *http.Request
	Path() string
	Body() io.Reader
	// TODO: Bindings?
}

type PublishClient struct {
	channelName    string
	url            *url.URL
	bindings       *runHttp.ChannelBindings
	NewRequest     func(ctx context.Context, method, url string, e ImplementationRecord) (*http.Request, error)
	HandleResponse func(r *http.Response) error
}

func (p PublishClient) Send(ctx context.Context, envelopes ...runHttp.EnvelopeWriter) error {
	pool := run.NewErrorPool()
	method := p.bindings.PublisherBindings.Method
	if method == "" {
		method = "GET"
	}

	for _, envelope := range envelopes {
		rm := envelope.(ImplementationRecord)
		record := rm.RecordNetHTTP()
		u := record.URL.JoinPath(p.channelName, rm.Path())
		pool.Go(func() error {
			req, err := p.NewRequest(ctx, method, u.String(), rm)
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

func NewRequest(ctx context.Context, method, url string, e ImplementationRecord) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, e.Body())
	if err != nil {
		return nil, err
	}

	recordReq := e.RecordNetHTTP()
	if recordReq.Proto != "" {
		req.Proto = recordReq.Proto
	}
	if recordReq.ProtoMajor != 0 {
		req.ProtoMajor = recordReq.ProtoMajor
		req.ProtoMinor = recordReq.ProtoMinor
	}
	if recordReq.Header != nil {
		req.Header = recordReq.Header
	}
	if recordReq.ContentLength > 0 {
		req.ContentLength = recordReq.ContentLength
	}
	req.TransferEncoding = recordReq.TransferEncoding
	req.Close = recordReq.Close
	if recordReq.Host != "" {
		req.Host = recordReq.Host
	}
	req.Form = recordReq.Form
	req.PostForm = recordReq.PostForm
	req.MultipartForm = recordReq.MultipartForm
	req.Trailer = recordReq.Trailer
	req.Response = recordReq.Response

	return req, nil
}
