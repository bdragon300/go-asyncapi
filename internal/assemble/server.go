package assemble

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type Server struct {
	Protocol string
}

func (s Server) AllowRender() bool {
	return true
}

func (s Server) AssembleDefinition(_ *common.AssembleContext) []*jen.Statement {
	return nil
}

func (s Server) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	panic("not implemented")
}
