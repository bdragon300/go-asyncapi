package amqp091go

import (
	"context"
	"errors"
	"time"

	runAmqp "github.com/bdragon300/go-asyncapi/run/amqp"

	"github.com/rabbitmq/amqp091-go"
)

type PublishChannel struct {
	*amqp091.Channel
	exchangeName    string
	channelBindings *runAmqp.ChannelBindings
	operationBindings *runAmqp.OperationBindings
}

type ImplementationRecord interface {
	AsAMQP091Record() *amqp091.Publishing
	RoutingKey() string
}

func (p PublishChannel) Send(ctx context.Context, envelopes ...runAmqp.EnvelopeWriter) error {
	var err error
	for _, envelope := range envelopes {
		rm := envelope.(ImplementationRecord)
		record := rm.AsAMQP091Record()
		record.DeliveryMode = uint8(p.operationBindings.DeliveryMode)
		record.Priority = uint8(p.operationBindings.Priority)
		record.Timestamp = time.Time{}
		if p.operationBindings.Timestamp {
			record.Timestamp = time.Now()
		}
		record.ReplyTo = p.operationBindings.ReplyTo
		record.UserId = p.operationBindings.UserID
		if p.operationBindings.Expiration > 0 {
			record.Expiration = p.operationBindings.Expiration.String()
		}
		if len(p.operationBindings.CC) > 0 {
			record.Headers["CC"] = p.operationBindings.CC
		}
		if len(p.operationBindings.BCC) > 0 {
			record.Headers["BCC"] = p.operationBindings.BCC
		}

		err = errors.Join(err, p.Channel.PublishWithContext(
			ctx, p.exchangeName, rm.RoutingKey(), p.operationBindings.Mandatory, false, *record,
		))
	}
	return err
}
