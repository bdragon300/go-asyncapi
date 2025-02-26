package types

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// CompileError is an error that occurred during the compilation stage. Contains the failed document entity path and
// optionally the protocol name.
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

// MultilineError is an error that contains the multiline content along with the error message.
type MultilineError struct {
	Err     error
	Content []byte
}

func (e MultilineError) Error() string {
	return e.Err.Error()
}

func (e MultilineError) Unwrap() error {
	return e.Err
}

// ContentLines returns the multiline content with line numbers.
func (e MultilineError) ContentLines() string {
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
