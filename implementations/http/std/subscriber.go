package std

import (
	"context"

	"github.com/bdragon300/go-asyncapi/run"
	runHttp "github.com/bdragon300/go-asyncapi/run/http"
)

func NewSubscriber(bindings *runHttp.ChannelBindings) *Subscriber {
	res := Subscriber{
		bindings: bindings,
		items:    run.NewFanOut[runHttp.EnvelopeReader](),
	}
	res.ctx, res.cancel = context.WithCancelCause(context.Background())

	return &res
}

type Subscriber struct {
	bindings *runHttp.ChannelBindings
	items    *run.FanOut[runHttp.EnvelopeReader]
	ctx      context.Context
	cancel   context.CancelCauseFunc
}

func (s *Subscriber) Receive(ctx context.Context, cb func(envelope runHttp.EnvelopeReader)) error {
	el := s.items.Add(cb)
	defer s.items.Remove(el)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return context.Cause(s.ctx)
	}
}

func (s *Subscriber) Close() error {
	s.cancel(nil)
	return nil
}
