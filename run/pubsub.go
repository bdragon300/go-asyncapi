package run

import (
	"context"
	"errors"
	"io"
)

type AbstractProducer[B any, W AbstractEnvelopeWriter, P AbstractPublisher[W]] interface {
	NewPublisher(channelName string, bindings *B) (P, error)
}
type AbstractPublisher[W AbstractEnvelopeWriter] interface {
	Send(ctx context.Context, envelopes ...W) error
	Close() error
}
type AbstractEnvelopeWriter interface {
	io.Writer
	ResetPayload()
	SetHeaders(headers Headers)
	SetContentType(contentType string)
}

type AbstractConsumer[B any, R AbstractEnvelopeReader, S AbstractSubscriber[R]] interface {
	NewSubscriber(channelName string, bindings *B) (S, error)
}
type AbstractSubscriber[R AbstractEnvelopeReader] interface {
	Receive(ctx context.Context, cb func(envelope R) error) error
	Close() error
}
type AbstractEnvelopeReader interface {
	io.Reader
	Headers() Headers
}

type PublisherFanOut[W AbstractEnvelopeWriter, P AbstractPublisher[W]] struct {
	Publishers []P
}

func (p PublisherFanOut[W, P]) Send(ctx context.Context, envelopes ...W) error {
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

func (p PublisherFanOut[W, P]) Close() (err error) {
	for _, pub := range p.Publishers {
		err = errors.Join(pub.Close())
	}
	return
}

type SubscriberFanIn[R AbstractEnvelopeReader, S AbstractSubscriber[R]] struct {
	Subscribers      []S
	StopOnFirstError bool
}

func (s SubscriberFanIn[R, S]) Receive(ctx context.Context, cb func(envelope R) error) error {
	if len(s.Subscribers) == 1 {
		return s.Subscribers[0].Receive(ctx, cb)
	}

	poolCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pool := NewErrorPool()

	for i := 0; i < len(s.Subscribers); i++ {
		i := i
		pool.Go(func() error {
			return s.Subscribers[i].Receive(poolCtx, func(envelope R) error {
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

func (s SubscriberFanIn[R, S]) Close() (err error) {
	for _, sub := range s.Subscribers {
		err = errors.Join(sub.Close())
	}
	return
}

func GatherPublishers[W AbstractEnvelopeWriter, PUB AbstractPublisher[W], B any, PRD AbstractProducer[B, W, PUB]](
	chName ParamString,
	channelBindings *B,
	producers []PRD,
) ([]PUB, error) {
	pubsCh := make(chan PUB, len(producers))
	pool := NewErrorPool()
	for _, prod := range producers {
		prod := prod
		pool.Go(func() error {
			p, e := prod.NewPublisher(chName.String(), channelBindings)
			pubsCh <- p
			return e
		})
	}
	err := pool.Wait()
	close(pubsCh)

	pubs := make([]PUB, 0, len(producers))
	for pub := range pubsCh {
		pubs = append(pubs, pub)
		if err != nil {
			err = errors.Join(err, pub.Close())
		}
	}
	return pubs, nil
}

func GatherSubscribers[R AbstractEnvelopeReader, S AbstractSubscriber[R], B any, C AbstractConsumer[B, R, S]](
	chName ParamString,
	channelBindings *B,
	consumers []C,
) ([]S, error) {
	subsCh := make(chan S, len(consumers))
	pool := NewErrorPool()
	for _, cons := range consumers {
		cons := cons
		pool.Go(func() error {
			s, e := cons.NewSubscriber(chName.String(), channelBindings)
			subsCh <- s
			return e
		})
	}
	err := pool.Wait()
	close(subsCh)

	subs := make([]S, 0, len(consumers))
	for sub := range subsCh {
		subs = append(subs, sub)
		if err != nil {
			err = errors.Join(err, sub.Close())
		}
	}
	return subs, nil
}
