package lang

import (
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/dave/jennifer/jen"
)

type Server struct {
	Protocol string
}

func (s Server) AllowRender() bool {
	return true
}

func (s Server) RenderDefinition(_ *render.Context) []*jen.Statement {
	return nil
}

func (s Server) RenderUsage(_ *render.Context) []*jen.Statement {
	panic("not implemented")
}
