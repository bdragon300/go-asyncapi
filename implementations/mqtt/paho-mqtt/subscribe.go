package pahomqtt

import (
	"context"
	"errors"
	"sync"

	"github.com/bdragon300/go-asyncapi/run"
	runMqtt "github.com/bdragon300/go-asyncapi/run/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SubscribeChannel struct {
	Client mqtt.Client
	Topic  string

	bindings      *runMqtt.ChannelBindings
	subscribeChan *run.FanOut[runMqtt.EnvelopeReader]
	instances     int
	mu            *sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func (r *SubscribeChannel) Receive(ctx context.Context, cb func(envelope runMqtt.EnvelopeReader) error) error {
	el := r.subscribeChan.Add(cb)
	defer r.subscribeChan.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.ctx.Done():
		return errors.New("channel closed")
	}
}

func (r *SubscribeChannel) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.instances <= 0 {
		return errors.New("channel already closed")
	}

	r.instances--
	if r.instances == 0 {
		r.cancel()
		if r.subscribeChan != nil {
			if tok := r.Client.Unsubscribe(r.Topic); tok.Wait() && tok.Error() != nil {
				return tok.Error()
			}
		}
	}
	return nil
}
