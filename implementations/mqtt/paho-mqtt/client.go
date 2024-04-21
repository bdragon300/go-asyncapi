package pahomqtt

import (
	"context"
	"net/url"
	"sync"

	"github.com/bdragon300/go-asyncapi/run"

	runMqtt "github.com/bdragon300/go-asyncapi/run/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewClient(ctx context.Context, serverURL string, bindings *runMqtt.ServerBindings, initClientOptions *mqtt.ClientOptions) (*Client, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	co := initClientOptions
	if co == nil {
		co = mqtt.NewClientOptions()
	}

	co.AddBroker(u.String())
	if bindings != nil {
		co.SetCleanSession(bindings.CleanSession)
		if bindings.ClientID != "" {
			co.SetClientID(bindings.ClientID)
		}
		if bindings.KeepAlive != 0 {
			co.SetKeepAlive(bindings.KeepAlive)
		}
		if bindings.LastWill != nil {
			co.SetWill(bindings.LastWill.Topic, bindings.LastWill.Message, byte(bindings.LastWill.QoS), bindings.LastWill.Retain)
		}
	}

	cl := mqtt.NewClient(co)
	tok := cl.Connect()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-tok.Done():
		if tok.Error() != nil {
			return nil, tok.Error()
		}
	}

	return &Client{
		Client:      cl,
		bindings:    bindings,
		mu:          &sync.Mutex{},
		publishers:  make(map[string]*PublishChannel),
		subscribers: make(map[string]*SubscribeChannel),
	}, nil
}

type Client struct {
	mqtt.Client
	bindings    *runMqtt.ServerBindings
	mu          *sync.Mutex
	publishers  map[string]*PublishChannel
	subscribers map[string]*SubscribeChannel
}

func (c *Client) Subscriber(ctx context.Context, channelName string, bindings *runMqtt.ChannelBindings) (runMqtt.Subscriber, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.subscribers[channelName]; ok && v.instances > 0 {
		v.instances++
		return v, nil
	}

	var qos byte
	if bindings != nil {
		qos = byte(bindings.SubscriberBindings.QoS)
	}

	subCh := run.NewFanOut[runMqtt.EnvelopeReader]()
	tok := c.Client.Subscribe(channelName, qos, func(_ mqtt.Client, message mqtt.Message) {
		subCh.Put(func() runMqtt.EnvelopeReader { return NewEnvelopeIn(message) })
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-tok.Done():
		if tok.Error() != nil {
			return nil, tok.Error()
		}
	}

	ctx2, cancel := context.WithCancel(context.Background())
	r := SubscribeChannel{
		Client:        c.Client,
		Topic:         channelName,
		bindings:      bindings,
		subscribeChan: subCh,
		instances:     1,
		mu:            c.mu,
		ctx:           ctx2,
		cancel:        cancel,
	}
	c.subscribers[channelName] = &r
	return &r, nil
}

func (c *Client) Publisher(_ context.Context, channelName string, bindings *runMqtt.ChannelBindings) (runMqtt.Publisher, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.publishers[channelName]; ok && v.instances > 0 {
		v.instances++
		return v, nil
	}

	ctx2, cancel := context.WithCancel(context.Background())
	r := PublishChannel{
		Client:    c.Client,
		Topic:     channelName,
		bindings:  bindings,
		instances: 1,
		mu:        c.mu,
		ctx:       ctx2,
		cancel:    cancel,
	}
	c.publishers[channelName] = &r
	return &r, nil
}
