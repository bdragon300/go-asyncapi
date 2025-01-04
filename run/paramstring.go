package run

import (
	"github.com/bdragon300/go-asyncapi/run/3rdparty/uritemplates"
	"net/url"
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
	if len(c.Parameters) == 0 {
		return c.Expr, nil
	}

	_, res, err := uritemplates.Expand(c.Expr, c.Parameters)
	return res, err
}

type ParamURL struct {
	Host 	 string
	Pathname string
	Parameters   map[string]string
}

func (c ParamURL) String() string {
	key, err := c.Expand()
	if err != nil {
		panic(err)
	}
	return key
}

func (c ParamURL) Expand() (string, error) {
	s, err := url.JoinPath(c.Host, c.Pathname)
	if err != nil {
		return "", err
	}

	_, res, err := uritemplates.Expand(s, c.Parameters)
	return res, err
}