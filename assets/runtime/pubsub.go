package runtime

import (
	"context"
)

type Publisher[E any] interface {
	Send(ctx context.Context, envelopes ...*E) error
	Close() error
}

type Subscriber[E any] interface {
	Receive(ctx context.Context, cb func(envelope *E) error) error
	Close() error
}

func NewPublisherFanOut[E any](publishers []Publisher[E]) *PublisherFanOut[E] {
	return &PublisherFanOut[E]{publishers: publishers}
}

type PublisherFanOut[E any] struct {
	publishers []Publisher[E]
}

func (p PublisherFanOut[E]) Send(ctx context.Context, envelopes ...*E) error {
	pool := NewErrorPool()

	for i := 0; i < len(p.publishers); i++ {
		i := i
		pool.Go(func() error {
			return p.publishers[i].Send(ctx, envelopes...)
		})
	}
	return pool.Wait()
}

func (p PublisherFanOut[E]) Close() error {
	return nil
}

func NewSubscriberFanIn[E any](subscribers []Subscriber[E], bufSize int, stopOnFirstError bool) *SubscriberFanIn[E] {
	return &SubscriberFanIn[E]{subscribers: subscribers, bufSize: bufSize, stopOnFirstError: stopOnFirstError}
}

type SubscriberFanIn[E any] struct {
	subscribers      []Subscriber[E]
	bufSize          int
	stopOnFirstError bool
}

func (s SubscriberFanIn[E]) Receive(ctx context.Context, cb func(envelope *E) error) error {
	poolCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pool := NewErrorPool()

	for i := 0; i < len(s.subscribers); i++ {
		i := i
		pool.Go(func() error {
			// TODO: add option to ignore error in one connection or close all connections
			return s.subscribers[i].Receive(poolCtx, func(envelope *E) error {
				err := cb(envelope)
				if err != nil && s.stopOnFirstError {
					cancel()
				}
				return err
			})
		})
	}
	return pool.Wait()
}

func (s SubscriberFanIn[E]) Close() error {
	return nil
}

func ToSlice[T any](elements ...T) []T {
	return elements
}
