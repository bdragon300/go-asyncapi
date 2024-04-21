package run

import (
	"container/list"
	"context"
	"errors"
	"reflect"
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
}

func (s SubscriberFanIn[R, S]) Receive(ctx context.Context, cb func(envelope R)) error {
	if len(s.Subscribers) == 1 {
		return s.Subscribers[0].Receive(ctx, cb)
	}

	poolCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	pool := NewErrorPool()

	for i := 0; i < len(s.Subscribers); i++ {
		i := i
		pool.Go(func() error {
			return s.Subscribers[i].Receive(poolCtx, cb)
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
	ctx context.Context,
	chName ParamString,
	channelBindings *B,
	producers []PRD,
) ([]PUB, error) {
	pubsCh := make(chan PUB, len(producers))
	pool := NewErrorPool()
	for _, prod := range producers {
		prod := prod
		pool.Go(func() error {
			p, e := prod.Publisher(ctx, chName.String(), channelBindings)
			pubsCh <- p
			return e
		})
	}
	err := pool.Wait()
	close(pubsCh)

	pubs := make([]PUB, 0, len(producers))
	var zero PUB
	for pub := range pubsCh {
		pubs = append(pubs, pub)
		if err != nil && !reflect.DeepEqual(pub, zero) {
			err = errors.Join(err, pub.Close())  // Close subscribers on error to avoid resource leak
		}
	}
	return pubs, err
}

func GatherSubscribers[R AbstractEnvelopeReader, SUB AbstractSubscriber[R], B any, C AbstractConsumer[B, R, SUB]](
	ctx context.Context,
	chName ParamString,
	channelBindings *B,
	consumers []C,
) ([]SUB, error) {
	subsCh := make(chan SUB, len(consumers))
	pool := NewErrorPool()
	for _, cons := range consumers {
		cons := cons
		pool.Go(func() error {
			s, e := cons.Subscriber(ctx, chName.String(), channelBindings)
			subsCh <- s
			return e
		})
	}
	err := pool.Wait()
	close(subsCh)

	subs := make([]SUB, 0, len(consumers))
	var zero SUB
	for sub := range subsCh {
		subs = append(subs, sub)
		if err != nil && !reflect.DeepEqual(sub, zero) {
			err = errors.Join(err, sub.Close())  // Close subscribers on error to avoid resource leak
		}
	}
	return subs, err
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

func (cm *FanOut[MessageT]) Add(cb func(msg MessageT)) *list.Element {
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
		item.Value.(func(msg MessageT))(msg)
	}
}
