package runtime

import (
	"context"
	"errors"
)

type PublisherFanOut[E EnvelopeWriter] struct {
	Publishers []Publisher[E]
}

func (p PublisherFanOut[E]) Send(ctx context.Context, envelopes ...E) error {
	if len(p.Publishers) == 1 {
		return p.Publishers[0].Send(ctx, envelopes...)
	}

	pool := NewErrorPool()

	for i := 0; i < len(p.Publishers); i++ {
		i := i
		pool.Go(func() error {
			return p.Publishers[i].Send(ctx, envelopes...)
		})
	}
	return pool.Wait()
}

func (p PublisherFanOut[E]) Close() (err error) {
	for _, pub := range p.Publishers {
		if pub != nil {
			err = errors.Join(pub.Close())
		}
	}
	return
}

type SubscriberFanIn[E EnvelopeReader] struct {
	Subscribers      []Subscriber[E]
	StopOnFirstError bool
}

func (s SubscriberFanIn[E]) Receive(ctx context.Context, cb func(envelope E) error) error {
	if len(s.Subscribers) == 1 {
		return s.Subscribers[0].Receive(ctx, cb)
	}

	poolCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pool := NewErrorPool()

	for i := 0; i < len(s.Subscribers); i++ {
		i := i
		pool.Go(func() error {
			return s.Subscribers[i].Receive(poolCtx, func(envelope E) error {
				err := cb(envelope)
				if err != nil && s.StopOnFirstError {
					cancel()
				}
				return err
			})
		})
	}
	return pool.Wait()
}

func (s SubscriberFanIn[E]) Close() (err error) {
	for _, sub := range s.Subscribers {
		if sub != nil {
			err = errors.Join(sub.Close())
		}
	}
	return
}

func GatherPublishers[E EnvelopeWriter, B any, P Producer[B, E]](chName ParamString, channelBindings *B, producers []P) ([]Publisher[E], error) {
	pubs := make([]Publisher[E], len(producers))
	pool := NewErrorPool()
	for ind, prod := range producers {
		prod := prod
		ind := ind
		pool.Go(func() error {
			var e error
			pubs[ind], e = prod.Publisher(chName.String(), channelBindings)
			return e
		})
	}
	if err := pool.Wait(); err != nil {
		for _, pub := range pubs {
			err = errors.Join(err, pub.Close())
		}
		return nil, err
	}

	return pubs, nil
}

func GatherSubscribers[E EnvelopeReader, B any, C Consumer[B, E]](chName ParamString, channelBindings *B, consumers []C) ([]Subscriber[E], error) {
	subs := make([]Subscriber[E], len(consumers))
	pool := NewErrorPool()
	for ind, cons := range consumers {
		cons := cons
		ind := ind
		pool.Go(func() error {
			var e error
			subs[ind], e = cons.Subscriber(chName.String(), channelBindings)
			return e
		})
	}
	if err := pool.Wait(); err != nil {
		for _, sub := range subs {
			err = errors.Join(err, sub.Close())
		}
		return nil, err
	}

	return subs, nil
}
