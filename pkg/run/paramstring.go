package run

import "github.com/bdragon300/asyncapi-codegen-go/pkg/run/3rdparty/uritemplates"

type ParamString struct {
	Expr       string
	Parameters map[string]string
}

func (c ParamString) String() string {
	if len(c.Parameters) == 0 {
		return c.Expr
	}
	key, err := c.Expand()
	if err != nil {
		panic(err)
	}
	return key
}

func (c ParamString) Expand() (string, error) {
	_, key, err := uritemplates.Expand(c.Expr, c.Parameters)
	return key, err
}
