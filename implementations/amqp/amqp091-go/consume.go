package amqp091go

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
	"github.com/bdragon300/asyncapi-codegen-go/pkg/run/amqp"
	amqp091 "github.com/rabbitmq/amqp091-go"
)

func NewConsumer(url string, bindings *amqp.ServerBindings) (*Consumer, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		Connection: conn,
		Bindings:   bindings,
	}, nil
}

type Consumer struct {
	*amqp091.Connection
	Bindings *amqp.ServerBindings
}

func (c Consumer) Subscriber(channelName string, bindings *amqp.ChannelBindings) (run.Subscriber[*EnvelopeIn], error) {
	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}

	qc := bindings.QueueConfiguration
	declare := qc.Durable != nil || qc.Exclusive != nil || qc.AutoDelete != nil || qc.VHost != ""
	var queueName string
	if bindings.ChannelType == amqp.ChannelTypeQueue {
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

	return &Subscriber{
		Channel:   ch,
		queueName: queueName,
		bindings:  bindings,
	}, nil
}

type Subscriber struct {
	*amqp091.Channel
	queueName string
	bindings  *amqp.ChannelBindings
}

func (s Subscriber) Receive(ctx context.Context, cb func(envelope *EnvelopeIn) error) (err error) {
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
