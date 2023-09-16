package assemble

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type Server struct {
	Protocol    string
	ProtoServer common.Assembler
}

func (s Server) AllowRender() bool {
	return true
}

func (s Server) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return s.ProtoServer.AssembleDefinition(ctx)
}

func (s Server) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	panic("not implemented")
}
