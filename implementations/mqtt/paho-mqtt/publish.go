package pahomqtt

import (
	"context"
	"errors"
	"sync"

	runMqtt "github.com/bdragon300/go-asyncapi/run/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type PublishChannel struct {
	Client mqtt.Client
	Topic  string

	bindings  *runMqtt.ChannelBindings
	instances int
	mu        *sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

type ImplementationRecord interface {
	RecordPaho() []byte
}

func (r *PublishChannel) Send(ctx context.Context, envelopes ...runMqtt.EnvelopeWriter) error {
	for _, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		tok := r.Client.Publish(r.Topic, byte(r.bindings.PublisherBindings.QoS), r.bindings.PublisherBindings.Retain, ir.RecordPaho())

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tok.Done():
			if tok.Error() != nil {
				return tok.Error()
			}
		case <-r.ctx.Done():
			return errors.New("channel closed")
		}
	}
	return nil
}

func (r *PublishChannel) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.instances <= 0 {
		return errors.New("channel already closed")
	}

	r.instances--
	if r.instances == 0 {
		r.cancel()
	}
	return nil
}
