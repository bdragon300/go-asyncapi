package amqp091go

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bdragon300/asyncapi-codegen-go/run"
	runAmqp "github.com/bdragon300/asyncapi-codegen-go/run/amqp"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

func NewProducer(serverURL string, bindings *runAmqp.ServerBindings) (*ProduceClient, error) {
	conn, err := amqp091.Dial(serverURL)
	if err != nil {
		return nil, err
	}
	return &ProduceClient{
		Connection: conn,
		Bindings:   bindings,
	}, nil
}

type ProduceClient struct {
	*amqp091.Connection
	Bindings *runAmqp.ServerBindings
}

func (p ProduceClient) Publisher(channelName string, bindings *runAmqp.ChannelBindings) (runAmqp.Publisher, error) {
	ch, err := p.Channel()
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
	return &PublishClient{
		Channel:      ch,
		routingKey:   channelName,
		exchangeName: exchangeName,
		bindings:     bindings,
	}, nil
}

type PublishClient struct {
	*amqp091.Channel
	routingKey   string
	exchangeName string
	bindings     *runAmqp.ChannelBindings
}

func (p PublishClient) Send(ctx context.Context, envelopes ...runAmqp.EnvelopeWriter) error {
	var err error
	for _, envelope := range envelopes {
		e := envelope.(*EnvelopeOut)
		e.DeliveryMode = uint8(p.bindings.PublisherBindings.DeliveryMode)
		e.Priority = uint8(p.bindings.PublisherBindings.Priority)
		e.Timestamp = time.Time{}
		if p.bindings.PublisherBindings.Timestamp {
			e.Timestamp = time.Now()
		}
		e.ReplyTo = p.bindings.PublisherBindings.ReplyTo
		e.UserId = p.bindings.PublisherBindings.UserID
		if p.bindings.PublisherBindings.Expiration > 0 {
			e.Expiration = p.bindings.PublisherBindings.Expiration.String()
		}
		if len(p.bindings.PublisherBindings.CC) > 0 {
			e.Headers["CC"] = p.bindings.PublisherBindings.CC
		}
		if len(p.bindings.PublisherBindings.BCC) > 0 {
			e.Headers["BCC"] = p.bindings.PublisherBindings.BCC
		}

		err = errors.Join(err, p.Channel.PublishWithContext(
			ctx, p.exchangeName, p.routingKey, p.bindings.PublisherBindings.Mandatory, false, *e.Publishing,
		))
	}
	return err
}
