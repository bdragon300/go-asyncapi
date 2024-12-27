package types

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type CompileError struct {
	Err   error
	Path  string
	Proto string
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

type Error error
type ErrorWithContent struct {
	Err error
	Content []byte
}

func (e ErrorWithContent) Error() string {
	return e.Err.Error()
}

func (e ErrorWithContent) Unwrap() error {
	return e.Err
}

func (e ErrorWithContent) ContentLines() string {
	var b strings.Builder
	rd := bufio.NewReader(bytes.NewReader(e.Content))

	for line := 1; ; line++ {
		s, err := rd.ReadString('\n')
		if err != nil {
			break // Suppose that the only error can appear here is io.EOF
		}
		b.WriteString(fmt.Sprintf("%-4d│ ", line))
		b.WriteString(s)
	}

	return b.String()
}

