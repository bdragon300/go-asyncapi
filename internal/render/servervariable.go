package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	j "github.com/dave/jennifer/jen"
)

type ServerVariable struct {
	Name        string
	Enum        []string // TODO: implement validation
	Default     string
	Description string // TODO
}

func (s ServerVariable) DirectRendering() bool {
	return false
}

func (s ServerVariable) RenderDefinition(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (s ServerVariable) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (s ServerVariable) ID() string {
	return s.Name
}

func (s ServerVariable) String() string {
	return "ServerVariable " + s.Name
}
