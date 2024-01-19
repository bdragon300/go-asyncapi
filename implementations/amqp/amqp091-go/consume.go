package amqp091go

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bdragon300/go-asyncapi/run"
	runAmqp "github.com/bdragon300/go-asyncapi/run/amqp"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

func NewConsumer(url string, bindings *runAmqp.ServerBindings) (*ConsumeClient, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	return &ConsumeClient{
		Connection: conn,
		bindings:   bindings,
	}, nil
}

type ConsumeClient struct {
	*amqp091.Connection
	bindings *runAmqp.ServerBindings
}

func (c ConsumeClient) Subscriber(channelName string, bindings *runAmqp.ChannelBindings) (runAmqp.Subscriber, error) {
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

	return &SubscribeClient{
		Channel:   ch,
		queueName: queueName,
		bindings:  bindings,
	}, nil
}

type SubscribeClient struct {
	*amqp091.Channel
	queueName string
	bindings  *runAmqp.ChannelBindings
}

func (s SubscribeClient) Receive(ctx context.Context, cb func(envelope runAmqp.EnvelopeReader) error) (err error) {
	// TODO: consumer tag in x- schema argument
	consumerTag := fmt.Sprintf("consumer-%s", time.Now().Format(time.RFC3339))
	deliveries, err := s.ConsumeWithContext(
		ctx,
		s.queueName,
		consumerTag,
		!s.bindings.SubscriberBindings.Ack,
		run.DerefOrZero(s.bindings.QueueConfiguration.Exclusive),
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, s.Cancel(consumerTag, false))
	}()
	for delivery := range deliveries {
		evlp := EnvelopeIn{
			Delivery: &delivery,
			reader:   bytes.NewReader(delivery.Body),
		}
		if err = cb(&evlp); err != nil {
			err = fmt.Errorf("message callback: %w", err)
			if s.bindings.SubscriberBindings.Ack {
				if e := s.Nack(delivery.DeliveryTag, false, false); e != nil {
					err = errors.Join(err, fmt.Errorf("nack: %w", e))
				}
			}
			return err
		}
		if s.bindings.SubscriberBindings.Ack {
			if e := s.Ack(delivery.DeliveryTag, false); e != nil {
				return fmt.Errorf("ack: %w", e)
			}
		}
	}
	return
}
