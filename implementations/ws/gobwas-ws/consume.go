package gobwasws

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	runWs "github.com/bdragon300/go-asyncapi/run/ws"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func NewConsumer(bindings *runWs.ServerBindings, responseTimeout time.Duration) (*ConsumeClient, error) {
	return &ConsumeClient{
		responseTimeout: responseTimeout,
		bindings:        bindings,
		connections:     make(map[string]chan *Connection),
		mu:              new(sync.RWMutex),
	}, nil
}

type ConsumeClient struct {
	http.ServeMux
	responseTimeout time.Duration
	bindings        *runWs.ServerBindings
	connections     map[string]chan *Connection
	mu              *sync.RWMutex
}

func (c *ConsumeClient) NewSubscriber(channelName string, bindings *runWs.ChannelBindings) (runWs.Subscriber, error) {
	c.ensureChannel(channelName, bindings)
	conn, ok := <-c.connections[channelName]
	if !ok {
		// Consumer has been closed while waiting for connection
		return nil, errors.New("consumer closed")
	}
	return conn, nil
}

func (c *ConsumeClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, ch := range c.connections {
		close(ch)
	}
	c.connections = nil
}

func (c *ConsumeClient) ensureChannel(channelName string, bindings *runWs.ChannelBindings) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.connections[channelName]; !ok { // HandleFunc panics if called more than once for the same channel
		c.connections[channelName] = make(chan *Connection)
		c.HandleFunc(channelName, func(w http.ResponseWriter, req *http.Request) {
			needMethod := bindings.Method
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

			netConn, rw, _, err := ws.UpgradeHTTP(req, w)
			if err != nil {
				// TODO: error log
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			ctx, cancel := context.WithTimeout(req.Context(), c.responseTimeout)
			defer cancel()

			conn := NewConnection(bindings, netConn, rw)
			select {
			case <-ctx.Done():
				// TODO: test when messages has came in between UpgradeHTTP and this, maybe it's needed to drain out?
				// TODO: error log
				defer conn.Close()
				_ = wsutil.WriteServerMessage(netConn, ws.OpClose, []byte("timeout exceeded"))
			case c.connections[channelName] <- conn:
			}
		})
	}
}
