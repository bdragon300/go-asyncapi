package run

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
)

type panicError struct {
	val   any
	stack []byte
}

func (p panicError) Error() string {
	return fmt.Sprintf("Panic: %v\n%s", p.val, p.stack)
}

func NewErrorPool() *ErrorPool {
	return &ErrorPool{
		running: &atomic.Int64{},
		errs: make(chan error),
	}
}

type ErrorPool struct {
	running *atomic.Int64
	errs chan error
}

func (p *ErrorPool) Go(cb func() error) {
	p.running.Add(1)
	go func() {
		defer p.running.Add(-1)
		defer func() {
			if v := recover(); v != nil {
				p.errs <- panicError{
					val:   v,
					stack: debug.Stack(),
				}
			}
		}()
		p.errs <- cb()
	}()
}

func (p *ErrorPool) Wait() (err error) {
	for p.running.Load() > 0 {
		e := <-p.errs
		switch e.(type) {
		case panicError:
			panic(e.Error()) // Rethrow a panic occurred in a goroutine
		default:
			err = errors.Join(e)
		}
	}
	return
}
