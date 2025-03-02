package run

import (
	"github.com/bdragon300/go-asyncapi/run/3rdparty/uritemplates"
)

// ParamString represents a URI template expression compliant with [RFC6570] and its parameters mapping.
// The main purpose is to keep them together and expand the URI template with the given parameters.
//
// [RFC6570]: https://tools.ietf.org/html/rfc6570
type ParamString struct {
	Expr       string
	Parameters map[string]string
}

// String returns the expanded URI template expression. If error occurs, it panics.
func (c ParamString) String() string {
	key, err := c.Expand()
	if err != nil {
		panic(err)
	}
	return key
}

// Expand returns the expanded URI template expression.
func (c ParamString) Expand() (string, error) {
	if len(c.Parameters) == 0 || c.Expr == "" {
		return c.Expr, nil
	}

	_, res, err := uritemplates.Expand(c.Expr, c.Parameters)
	return res, err
}
