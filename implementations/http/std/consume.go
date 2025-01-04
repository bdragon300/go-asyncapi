package std

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/bdragon300/go-asyncapi/run"
	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewConsumer(bindings *runHttp.ServerBindings) (consumer *ConsumeClient, err error) {
	return &ConsumeClient{
		bindings:    bindings,
		subscribers: make(map[string]*run.FanOut[*EnvelopeIn]),
		mu:          &sync.RWMutex{},
	}, nil
}

type ConsumeClient struct {
	http.ServeMux
	bindings    *runHttp.ServerBindings
	subscribers map[string]*run.FanOut[*EnvelopeIn]
	mu          *sync.RWMutex
}

func (c *ConsumeClient) Subscriber(_ context.Context, address string, bindings *runHttp.ChannelBindings) (runHttp.Subscriber, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ensureChannel(address, bindings)
	subscriber := NewSubscriber(bindings)
	element := c.subscribers[address].Add(func(msg *EnvelopeIn) {
		subscriber.items.Put(func() runHttp.EnvelopeReader {
			return NewEnvelopeIn(msg.Clone(context.Background()), msg.ResponseWriter)
		})
	})

	// Remove a subscriber from the list when it has been closed
	go func() {
		<-subscriber.ctx.Done()
		c.mu.Lock()
		defer c.mu.Unlock()
		c.subscribers[address].Remove(element)
	}()

	return subscriber, nil
}

func (c *ConsumeClient) ensureChannel(channelName string, bindings *runHttp.ChannelBindings) {
	if _, ok := c.subscribers[channelName]; !ok { // HandleFunc panics if called more than once for the same channel
		c.subscribers[channelName] = run.NewFanOut[*EnvelopeIn]()
		c.HandleFunc(channelName, func(w http.ResponseWriter, req *http.Request) {
			if bindings != nil {
				needMethod := bindings.SubscriberBindings.Method
				if needMethod != "" && strings.ToUpper(needMethod) != req.Method {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
			}

			c.mu.RLock()
			defer c.mu.RUnlock()
			if _, ok := c.subscribers[channelName]; !ok {
				http.Error(w, "channel not found", http.StatusNotFound)
				return
			}
			c.subscribers[channelName].Put(func() *EnvelopeIn { return NewEnvelopeIn(req, w) })
		})
	}
}
