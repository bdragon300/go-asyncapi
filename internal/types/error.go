package types

import (
	"errors"
	"fmt"
)

type CompileError struct {
	Err   error
	Path  string
	Proto string // TODO: check everywhere if this field is set where needed
}

func (c CompileError) Error() string {
	if c.Proto != "" {
		return fmt.Sprintf("path=%q proto=%q: %v", c.Path, c.Proto, c.Err)
	}
	return fmt.Sprintf("path=%q: %v", c.Path, c.Err)
}

func (c CompileError) Unwrap() error {
	return c.Err
}

func (c CompileError) Is(e error) bool {
	v, ok := e.(CompileError)
	return ok && v.Path == c.Path || errors.Is(c.Err, e)
}
