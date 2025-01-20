package run

import (
	"github.com/bdragon300/go-asyncapi/run/3rdparty/uritemplates"
)

type ParamString struct {
	Expr       string
	Parameters map[string]string
}

func (c ParamString) String() string {
	key, err := c.Expand()
	if err != nil {
		panic(err)
	}
	return key
}

func (c ParamString) Expand() (string, error) {
	if len(c.Parameters) == 0 || c.Expr == "" {
		return c.Expr, nil
	}

	_, res, err := uritemplates.Expand(c.Expr, c.Parameters)
	return res, err
}
