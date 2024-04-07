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

func NewConsumer(serverURL string, bindings *runKafka.ServerBindings, extraOpts []kgo.Opt) (*ConsumeClient, error) {
	return &ConsumeClient{
		serverURL: serverURL,
		bindings:  bindings,
		extraOpts: extraOpts,
	}, nil
}

type ConsumeClient struct {
	serverURL string
	bindings  *runKafka.ServerBindings
	extraOpts []kgo.Opt
}

func (c ConsumeClient) Subscriber(_ context.Context, channelName string, bindings *runKafka.ChannelBindings) (runKafka.Subscriber, error) {
	// TODO: schema registry https://github.com/twmb/franz-go/blob/master/examples/schema_registry/schema_registry.go
	// TODO: bindings.ClientID, bindings.GroupID
	var opts []kgo.Opt

	u, err := url.Parse(c.serverURL)
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
	opts = append(opts, c.extraOpts...)

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &SubscribeChannel{
		Client:   cl,
		Topic:    topic,
		bindings: bindings,
	}, nil
}

type SubscribeChannel struct {
	*kgo.Client
	Topic             string
	IgnoreFetchErrors bool // TODO: add opts for Subscriber/Publisher interfaces
	bindings          *runKafka.ChannelBindings
}

func (s SubscribeChannel) Receive(ctx context.Context, cb func(envelope runKafka.EnvelopeReader)) error {
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
			cb(NewEnvelopeIn(r))
		})
		if batchError != nil {
			return fmt.Errorf("callback errors: %w", batchError)
		}
	}
}

func (s SubscribeChannel) Close() error {
	s.Client.Close()
	return nil
}
