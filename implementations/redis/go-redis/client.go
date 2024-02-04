package goredis

import (
	"context"

	runRedis "github.com/bdragon300/go-asyncapi/run/redis"
	"github.com/redis/go-redis/v9"
)

func NewClient(serverURL string) (*Client, error) {
	opts, err := redis.ParseURL(serverURL)
	if err != nil {
		return nil, err
	}
	cl := redis.NewClient(opts)
	return &Client{Client: cl}, nil
}

type Client struct {
	*redis.Client
}

func (c *Client) Publisher(_ context.Context, channelName string, _ *runRedis.ChannelBindings) (runRedis.Publisher, error) {
	return &PublishChannel{Client: c.Client, Name: channelName}, nil
}

func (c *Client) Subscriber(ctx context.Context, channelName string, _ *runRedis.ChannelBindings) (runRedis.Subscriber, error) {
	return &SubscriberChannel{
		PubSub: c.Client.Subscribe(ctx, channelName),
		Name:   channelName,
	}, nil
}
