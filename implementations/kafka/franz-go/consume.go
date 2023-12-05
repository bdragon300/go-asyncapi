package franzgo

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/pkg/run"
	"github.com/bdragon300/asyncapi-codegen-go/pkg/run/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
)

func NewConsumer(url string, bindings *kafka.ServerBindings) (*Consumer, error) {
	return &Consumer{
		URL:       url,
		Bindings:  bindings,
	}, nil
}

type Consumer struct {
	URL       string
	Bindings  *kafka.ServerBindings
	ExtraOpts []kgo.Opt
}

func (c Consumer) Subscriber(channelName string, bindings *kafka.ChannelBindings) (run.Subscriber[*EnvelopeIn], error) {
	// TODO: schema registry https://github.com/twmb/franz-go/blob/master/examples/schema_registry/schema_registry.go
	// TODO: bindings.ClientID, bindings.GroupID
	var opts []kgo.Opt

	u, err := url.Parse(c.URL)
	if err != nil {
		return nil, fmt.Errorf("url parse: %w", err)
	}
	opts = append(opts, kgo.SeedBrokers(strings.Split(u.Host, ",")...))

	topic := channelName
	if bindings.Topic != "" {
		topic = bindings.Topic
	}
	if topic != "" {
		opts = append(opts, kgo.ConsumeTopics(topic))
	}
	opts = append(opts, c.ExtraOpts...)

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return Subscriber{
		Client:   cl,
		Topic:    topic,
		bindings: bindings,
	}, nil
}

type Subscriber struct {
	Client            *kgo.Client
	Topic             string
	IgnoreFetchErrors bool // TODO: add opts for Subscriber/Publisher interfaces
	bindings          *kafka.ChannelBindings
}

func (s Subscriber) Receive(ctx context.Context, cb func(envelope *EnvelopeIn) error) error {
	for {
		fetches := s.Client.PollFetches(ctx)
		if fetches.Err0() != nil {
			return fetches.Err0()
		}
		var batchError error

		if !s.IgnoreFetchErrors {
			fetches.EachError(func(topic string, partition int32, err error) {
				batchError = errors.Join(batchError, fmt.Errorf("topic=%q, partition=%v: %w", topic, partition, err))
			})
		}
		if batchError != nil {
			return fmt.Errorf("fetch errors: %w", batchError)
		}

		fetches.EachRecord(func(r *kgo.Record) {
			envelope := NewEnvelopeIn(r)
			batchError = errors.Join(batchError, cb(envelope))
		})
		if batchError != nil {
			return fmt.Errorf("callback errors: %w", batchError)
		}
	}
}

func (s Subscriber) Close() error {
	s.Client.Close()
	return nil
}
