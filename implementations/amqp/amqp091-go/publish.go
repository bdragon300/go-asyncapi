package amqp091go

import (
	"context"
	"errors"
	"time"

	runAmqp "github.com/bdragon300/go-asyncapi/run/amqp"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

type PublishChannel struct {
	*amqp091.Channel
	routingKey   string
	exchangeName string
	bindings     *runAmqp.ChannelBindings
}

type ImplementationRecord interface {
	AsAMQP091Record() *amqp091.Publishing
	DeliveryTag() string
}

func (p PublishChannel) Send(ctx context.Context, envelopes ...runAmqp.EnvelopeWriter) error {
	var err error
	for _, envelope := range envelopes {
		rm := envelope.(ImplementationRecord)
		record := rm.AsAMQP091Record()
		record.DeliveryMode = uint8(p.bindings.PublisherBindings.DeliveryMode)
		record.Priority = uint8(p.bindings.PublisherBindings.Priority)
		record.Timestamp = time.Time{}
		if p.bindings.PublisherBindings.Timestamp {
			record.Timestamp = time.Now()
		}
		record.ReplyTo = p.bindings.PublisherBindings.ReplyTo
		record.UserId = p.bindings.PublisherBindings.UserID
		if p.bindings.PublisherBindings.Expiration > 0 {
			record.Expiration = p.bindings.PublisherBindings.Expiration.String()
		}
		if len(p.bindings.PublisherBindings.CC) > 0 {
			record.Headers["CC"] = p.bindings.PublisherBindings.CC
		}
		if len(p.bindings.PublisherBindings.BCC) > 0 {
			record.Headers["BCC"] = p.bindings.PublisherBindings.BCC
		}

		err = errors.Join(err, p.Channel.PublishWithContext(
			ctx, p.exchangeName, p.routingKey, p.bindings.PublisherBindings.Mandatory, false, *record,
		))
	}
	return err
}
