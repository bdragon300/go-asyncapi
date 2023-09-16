package runtime

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

type panicError struct {
	val   any
	stack []byte
}

func (p panicError) Error() string {
	return fmt.Sprintf("Panic: %v\n%s", p.val, p.stack)
}

func NewErrorPool() *ErrorPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &ErrorPool{
		wg:     &sync.WaitGroup{},
		errs:   make(chan error),
		ctx:    ctx,
		cancel: cancel,
	}
}

type ErrorPool struct {
	wg     *sync.WaitGroup
	errs   chan error
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *ErrorPool) Go(cb func() error) {
	select {
	case <-p.ctx.Done():
		panic("Pool has already finished")
	default:
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() {
			if v := recover(); v != nil {
				p.errs <- &panicError{
					val:   v,
					stack: debug.Stack(),
				}
				p.cancel()
			}
		}()
		select {
		case <-p.ctx.Done():
		case p.errs <- cb():
		}
	}()
}

func (p *ErrorPool) Wait() (err error) {
	p.wg.Wait()
	p.cancel()
	close(p.errs)

	for e := range p.errs {
		switch e.(type) {
		case panicError:
			panic(e.Error()) // Repanic a panic occurred in a goroutine
		default:
			if err != nil {
				err = errors.Join(e)
			}
		}
	}
	return
}
