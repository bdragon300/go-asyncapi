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

type SubscribeChannel struct {
	*amqp091.Channel
	queueName string
	bindings  *runAmqp.ChannelBindings
}

func (s SubscribeChannel) Receive(ctx context.Context, cb func(envelope runAmqp.EnvelopeReader) error) (err error) {
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
