package pahomqtt

import (
	"context"
	"net/url"

	runMqtt "github.com/bdragon300/go-asyncapi/run/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewProducer(serverURL string, bindings *runMqtt.ServerBindings, initClientOptions *mqtt.ClientOptions) (*ProduceClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	co := initClientOptions
	if co == nil {
		co = mqtt.NewClientOptions()
	}

	co.AddBroker(u.String())
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

	return &ProduceClient{
		clientOptions: co,
		bindings:      bindings,
	}, nil
}

type ProduceClient struct {
	clientOptions *mqtt.ClientOptions
	bindings      *runMqtt.ServerBindings
}

func (p ProduceClient) NewPublisher(channelName string, bindings *runMqtt.ChannelBindings) (runMqtt.Publisher, error) {
	cl := mqtt.NewClient(p.clientOptions)
	tok := cl.Connect()
	if tok.Wait() && tok.Error() != nil {
		return nil, tok.Error()
	}

	return PublishClient{
		Client:   cl,
		Topic:    channelName,
		bindings: bindings,
	}, nil
}

type ImplementationRecord interface {
	RecordPaho() *mqtt.Message
}

type PublishClient struct {
	mqtt.Client
	Topic    string
	bindings *runMqtt.ChannelBindings
}

func (p PublishClient) Send(ctx context.Context, envelopes ...runMqtt.EnvelopeWriter) error {
	for _, envelope := range envelopes {
		ir := envelope.(ImplementationRecord)
		tok := p.Publish(p.Topic, byte(p.bindings.PublisherBindings.QoS), p.bindings.PublisherBindings.Retain, ir.RecordPaho())

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tok.Done():
			if tok.Error() != nil {
				return tok.Error()
			}
		}
	}
	return nil
}

func (p PublishClient) Close() error {
	p.Disconnect(0)
	return nil
}
