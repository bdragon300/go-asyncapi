package runtime

import (
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
	return &ErrorPool{
		wg:   &sync.WaitGroup{},
		errs: make(chan error),
	}
}

type ErrorPool struct {
	wg   *sync.WaitGroup
	errs chan error
}

func (p *ErrorPool) Go(cb func() error) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() {
			if v := recover(); v != nil {
				p.errs <- &panicError{
					val:   v,
					stack: debug.Stack(),
				}
			}
		}()
		p.errs <- cb()
	}()
}

func (p *ErrorPool) Wait() (err error) {
	p.wg.Wait()
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
