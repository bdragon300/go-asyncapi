package nethttp

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewConsumer(bindings *runHttp.ServerBindings, responseTimeout time.Duration) (consumer *ConsumeClient, err error) {
	return &ConsumeClient{
		responseTimeout: responseTimeout,
		bindings:        bindings,
		connections:     make(map[string]chan *Channel),
		mu:              &sync.RWMutex{},
	}, nil
}

type ConsumeClient struct {
	http.ServeMux
	responseTimeout time.Duration
	bindings        *runHttp.ServerBindings
	connections     map[string]chan *Channel
	mu              *sync.RWMutex
}

func (c *ConsumeClient) NewSubscriber(ctx context.Context, channelName string, bindings *runHttp.ChannelBindings) (runHttp.Subscriber, error) {
	c.ensureChannel(channelName, bindings)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-c.connections[channelName]:
		return conn, nil
	}
}

func (c *ConsumeClient) ensureChannel(channelName string, bindings *runHttp.ChannelBindings) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.connections[channelName]; !ok { // HandleFunc panics if called more than once for the same channel
		c.connections[channelName] = make(chan *Channel)
		c.HandleFunc(channelName, func(w http.ResponseWriter, req *http.Request) {
			needMethod := bindings.SubscriberBindings.Method
			if needMethod != "" && strings.ToUpper(needMethod) != req.Method {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			c.mu.RLock()
			defer c.mu.RUnlock()
			if _, ok := c.connections[channelName]; !ok {
				http.Error(w, "channel not found", http.StatusNotFound)
				return
			}

			hj, ok := w.(http.Hijacker)
			if !ok {
				// TODO: error log
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			// TODO: test connection keepalive after hijack
			// After hijack, the connection is no longer managed by the http server
			netConn, rw, err := hj.Hijack()
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			var origin *url.URL
			if o := req.Header.Get("Origin"); o != "" && o != "null" {
				origin, _ = url.Parse(o) // TODO: error log
			}

			ctx, cancel := context.WithTimeout(req.Context(), c.responseTimeout)
			defer cancel()

			conn := NewChannel(bindings, origin, netConn, rw)
			select {
			case <-ctx.Done():
				// TODO: error log
				defer conn.Close()
				http.Error(w, "timeout exceeded", http.StatusRequestTimeout)
			case c.connections[channelName] <- conn:
			}
		})
	}
}
