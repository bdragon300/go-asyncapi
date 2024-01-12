package franzgo

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	runKafka "github.com/bdragon300/go-asyncapi/run/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewConsumer(url string, bindings *runKafka.ServerBindings) (*ConsumeClient, error) {
	return &ConsumeClient{
		URL:      url,
		Bindings: bindings,
	}, nil
}

type ConsumeClient struct {
	URL       string
	Bindings  *runKafka.ServerBindings
	ExtraOpts []kgo.Opt
}

func (c ConsumeClient) Subscriber(channelName string, bindings *runKafka.ChannelBindings) (runKafka.Subscriber, error) {
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

	return &SubscribeClient{
		Client:   cl,
		Topic:    topic,
		bindings: bindings,
	}, nil
}

type SubscribeClient struct {
	Client            *kgo.Client
	Topic             string
	IgnoreFetchErrors bool // TODO: add opts for Subscriber/Publisher interfaces
	bindings          *runKafka.ChannelBindings
}

func (s SubscribeClient) Receive(ctx context.Context, cb func(envelope runKafka.EnvelopeReader) error) error {
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

func (s SubscribeClient) Close() error {
	s.Client.Close()
	return nil
}
