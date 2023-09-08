package runtime

import (
	"sync"
)

func FanIn[T any](channelBufferCap int, upstreams ...<-chan T) <-chan T {
	out := make(chan T, channelBufferCap)
	var wg sync.WaitGroup

	// Start an output goroutine for each input channel in upstreams.
	wg.Add(len(upstreams))
	for _, c := range upstreams {
		go func(c <-chan T) {
			for n := range c {
				out <- n
			}
			wg.Done()
		}(c)
	}

	// Start a goroutine to close out once all the output goroutines are done.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func StreamBy[F any, T any](channelBufferCap int, upstream <-chan F, cb func(item F) (T, error)) (<-chan T, <-chan error) {
	resChan := make(chan T, channelBufferCap)
	errChan := make(chan error, channelBufferCap)

	go func() {
		defer func() { close(resChan); close(errChan) }()
		for item := range upstream {
			if res, err := cb(item); err != nil {
				select {
				case errChan <- err:
				default:
				}
			} else {
				resChan <- res
			}
		}
	}()

	return resChan, errChan
}
