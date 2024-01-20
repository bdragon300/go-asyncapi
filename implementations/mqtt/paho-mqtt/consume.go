package pahomqtt

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	runMqtt "github.com/bdragon300/go-asyncapi/run/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewConsumer(serverURL string, bindings *runMqtt.ServerBindings, initClientOptions *mqtt.ClientOptions) (*ConsumeClient, error) {
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

	return &ConsumeClient{
		clientOptions: co,
		bindings:      bindings,
	}, nil
}

type ConsumeClient struct {
	clientOptions *mqtt.ClientOptions
	bindings      *runMqtt.ServerBindings
}

func (c ConsumeClient) NewSubscriber(channelName string, bindings *runMqtt.ChannelBindings) (runMqtt.Subscriber, error) {
	cl := mqtt.NewClient(c.clientOptions)
	tok := cl.Connect()
	if tok.Wait() && tok.Error() != nil {
		return nil, tok.Error()
	}

	ctx, cancel := context.WithCancel(context.Background())
	return SubscribeClient{
		Client:   cl,
		Topic:    channelName,
		bindings: bindings,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

type SubscribeClient struct {
	mqtt.Client
	Topic string

	bindings *runMqtt.ChannelBindings
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s SubscribeClient) Receive(ctx context.Context, cb func(envelope runMqtt.EnvelopeReader) error) (err error) {
	tok := s.Subscribe(s.Topic, byte(s.bindings.SubscriberBindings.QoS), func(client mqtt.Client, message mqtt.Message) {
		if e := cb(NewEnvelopeIn(message)); e != nil {
			err = errors.Join(err, fmt.Errorf("callback error: %w", e))
			panic(err)
		}
	})
	defer func() { // TODO: check if this is the right place to unsubscribe, maybe a separate method?
		tok := s.Unsubscribe(s.Topic)
		if tok.Wait() && tok.Error() != nil {
			err = errors.Join(err, tok.Error())
		}
	}()

	select {
	case <-s.ctx.Done():
		return errors.Join(err, s.ctx.Err())
	case <-ctx.Done():
		return errors.Join(err, ctx.Err())
	case <-tok.Done():
		return errors.Join(err, tok.Error()) // FIXME: exits just after subscription?
	}
}

func (s SubscribeClient) Close() error {
	s.cancel()
	s.Disconnect(0)
	return nil
}
