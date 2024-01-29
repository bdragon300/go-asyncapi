package goredis

import (
	"context"

	runRedis "github.com/bdragon300/go-asyncapi/run/redis"
	"github.com/redis/go-redis/v9"
)

type SubscriberChannel struct {
	*redis.PubSub
	Name string
}

func (s SubscriberChannel) Receive(ctx context.Context, cb func(envelope runRedis.EnvelopeReader)) error {
	for {
		select {
		case msg, ok := <-s.PubSub.Channel():
			if !ok {
				return nil
			}
			cb(NewEnvelopeIn(msg))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
