package run

import (
	"container/list"
	"context"
	"errors"
	"sync"
)

type PublisherFanOut[W AbstractEnvelopeWriter, P AbstractPublisher[W]] struct {
	Publishers []P
}

func (p PublisherFanOut[W, P]) Send(ctx context.Context, envelopes ...W) error {
	// TODO: use FanOut everywhere here
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
			p, e := prod.Publisher(context.TODO(), chName.String(), channelBindings) // TODO: context
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
			s, e := cons.Subscriber(context.TODO(), chName.String(), channelBindings) // TODO: context
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


func NewFanOut[MessageT any]() *FanOut[MessageT] {
	return &FanOut[MessageT]{
		receivers: list.New(),
		mu:        &sync.RWMutex{},
	}
}

type FanOut[MessageT any] struct {
	receivers *list.List
	mu        *sync.RWMutex
}

func (cm *FanOut[MessageT]) Add(cb func(msg MessageT) error) *list.Element {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	el := list.Element{Value: cb}
	cm.receivers.PushBack(&el)
	return &el
}

func (cm *FanOut[MessageT]) Remove(el *list.Element) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.receivers.Remove(el)
}

func (cm *FanOut[MessageT]) Put(msg MessageT) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.receivers.Len() == 0 {
		return
	}

	for item := cm.receivers.Front(); item != nil; item = item.Next() {
		item.Value.(func(msg MessageT) error)(msg)
	}
}
