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

func (c *Client) Subscriber(ctx context.Context, address string, chb *runMqtt.ChannelBindings, opb *runMqtt.OperationBindings) (runMqtt.Subscriber, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.subscribers[address]; ok && v.instances > 0 {
		v.instances++
		return v, nil
	}

	var qos byte
	if opb != nil {
		qos = byte(opb.QoS)
	}

	subCh := run.NewFanOut[runMqtt.EnvelopeReader]()
	tok := c.Client.Subscribe(address, qos, func(_ mqtt.Client, message mqtt.Message) {
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
		Client:          c.Client,
		Topic:           address,
		channelBindings: chb,
		operationBindings: opb,
		subscribeChan:   subCh,
		instances:       1,
		mu:              c.mu,
		ctx:             ctx2,
		cancel:          cancel,
	}
	c.subscribers[address] = &r
	return &r, nil
}

func (c *Client) Publisher(_ context.Context, address string, chb *runMqtt.ChannelBindings, opb *runMqtt.OperationBindings) (runMqtt.Publisher, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.publishers[address]; ok && v.instances > 0 {
		v.instances++
		return v, nil
	}

	ctx2, cancel := context.WithCancel(context.Background())
	r := PublishChannel{
		Client:          c.Client,
		Topic:           address,
		channelBindings: chb,
		operationBindings: opb,
		instances:       1,
		mu:              c.mu,
		ctx:             ctx2,
		cancel:          cancel,
	}
	c.publishers[address] = &r
	return &r, nil
}
