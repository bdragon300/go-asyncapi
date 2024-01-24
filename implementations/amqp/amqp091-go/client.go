package amqp091go

import (
	"context"
	"errors"
	"fmt"

	"github.com/bdragon300/go-asyncapi/run"
	runAmqp "github.com/bdragon300/go-asyncapi/run/amqp"
	"github.com/rabbitmq/amqp091-go"
)

func NewClient(serverURL string, bindings *runAmqp.ServerBindings) (*Client, error) {
	conn, err := amqp091.Dial(serverURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		Connection: conn,
		bindings:   bindings,
	}, nil
}

type Client struct {
	*amqp091.Connection
	bindings *runAmqp.ServerBindings
}

func (c Client) NewPublisher(_ context.Context, channelName string, bindings *runAmqp.ChannelBindings) (runAmqp.Publisher, error) {
	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}

	ec := bindings.ExchangeConfiguration
	declare := ec.Type != "" || ec.Durable != nil || ec.AutoDelete != nil || ec.VHost != ""
	var exchangeName string // By default, publish to the default exchange with empty name
	if bindings.ChannelType == runAmqp.ChannelTypeRoutingKey {
		exchangeName = channelName
	}
	if ec.Name != nil {
		exchangeName = *ec.Name
	}
	if declare {
		err = ch.ExchangeDeclare(
			exchangeName,
			string(ec.Type),
			run.DerefOrZero(ec.Durable),
			run.DerefOrZero(ec.AutoDelete),
			false,
			false,
			nil,
		)
		if err != nil {
			// TODO: close channel
			return nil, fmt.Errorf("exchange declare: %w", err)
		}
	}
	return &PublishChannel{
		Channel:      ch,
		routingKey:   channelName,
		exchangeName: exchangeName,
		bindings:     bindings,
	}, nil
}

func (c Client) NewSubscriber(_ context.Context, channelName string, bindings *runAmqp.ChannelBindings) (runAmqp.Subscriber, error) {
	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}

	qc := bindings.QueueConfiguration
	declare := qc.Durable != nil || qc.Exclusive != nil || qc.AutoDelete != nil || qc.VHost != ""
	var queueName string
	if bindings.ChannelType == runAmqp.ChannelTypeQueue {
		queueName = channelName
	}
	if qc.Name != "" {
		queueName = qc.Name
	}
	if declare {
		_, err = ch.QueueDeclare(
			queueName,
			run.DerefOrZero(qc.Durable),
			run.DerefOrZero(qc.AutoDelete),
			run.DerefOrZero(qc.Exclusive),
			false,
			nil,
		)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("queue declare: %w", err), ch.Close())
		}
	}
	exchangeName := run.DerefOrZero(bindings.ExchangeConfiguration.Name)
	// TODO: binding key in x- schema argument
	if err = ch.QueueBind(queueName, queueName, exchangeName, false, nil); err != nil {
		return nil, errors.Join(fmt.Errorf("queue bind: %w", err), ch.Close())
	}

	return &SubscribeChannel{
		Channel:   ch,
		queueName: queueName,
		bindings:  bindings,
	}, nil
}
