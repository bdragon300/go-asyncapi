package stdhttp

import (
	"container/list"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/bdragon300/go-asyncapi/run"
	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewConsumer(bindings *runHttp.ServerBindings) (consumer *ConsumeClient, err error) {
	return &ConsumeClient{
		Bindings:    bindings,
		subscribers: make(map[string]*list.List),
		mu:          &sync.RWMutex{},
	}, nil
}

type ConsumeClient struct {
	http.ServeMux
	Bindings *runHttp.ServerBindings

	subscribers map[string]*list.List // Subscribers by channel name
	mu          *sync.RWMutex
}

func (c *ConsumeClient) Subscriber(channelName string, bindings *runHttp.ChannelBindings) (runHttp.Subscriber, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.subscribers[channelName]; !ok {
		l := list.New()
		c.subscribers[channelName] = l
		c.HandleFunc(channelName, func(w http.ResponseWriter, req *http.Request) {
			c.mu.RLock()
			defer c.mu.RUnlock()

			if l.Len() == 0 {
				// No readers, drain out the body
				defer req.Body.Close()
				_, _ = io.Copy(io.Discard, req.Body)
				return
			}

			p := run.NewErrorPool()
			for item := l.Front(); item != nil; item = item.Next() {
				item := item
				p.Go(func() error {
					envelope := NewEnvelopeIn(req, w)
					return item.Value.(*SubscribeClient).receiveEnvelope(envelope)
				})
			}
			_ = p.Wait() // TODO: do smth with error?
		})
	}

	subCtx, subCtxCancel := context.WithCancel(context.Background())
	sub := SubscribeClient{
		bindings:  bindings,
		callbacks: list.New(),
		ctx:       subCtx,
		ctxCancel: subCtxCancel,
		mu:        &sync.RWMutex{},
	}
	el := c.subscribers[channelName].PushBack(&sub)
	go func() {
		<-subCtx.Done()
		c.mu.Lock()
		defer c.mu.Unlock()

		c.subscribers[channelName].Remove(el)
	}()

	return &sub, nil
}

type SubscribeClient struct {
	bindings  *runHttp.ChannelBindings
	callbacks *list.List
	ctx       context.Context
	ctxCancel context.CancelFunc
	mu        *sync.RWMutex
}

func (s *SubscribeClient) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	envelope := NewEnvelopeIn(request, writer)
	_ = s.receiveEnvelope(envelope) // TODO: to smth with error?
}

func (s *SubscribeClient) Receive(ctx context.Context, cb func(envelope runHttp.EnvelopeReader) error) error {
	var el *list.Element
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		el = s.callbacks.PushBack(cb)
	}()
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.callbacks.Remove(el)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return errors.New("subscriber closed")
	}
}

func (s *SubscribeClient) Close() error {
	s.ctxCancel()
	return nil
}

func (s *SubscribeClient) receiveEnvelope(envelope *EnvelopeIn) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p := run.NewErrorPool()
	for item := s.callbacks.Front(); item != nil; item = item.Next() {
		p.Go(func() error {
			return item.Value.(func(envelope *EnvelopeIn) error)(envelope)
		})
	}
	return p.Wait()
}
