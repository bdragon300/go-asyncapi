package assemble

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type ServerParts struct {
	Publish   common.Assembler
	Subscribe common.Assembler
	Common    common.Assembler
}

type Server struct {
	Protocol string
	Parts    ServerParts
}

func (s Server) AllowRender() bool {
	return true
}

func (s Server) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	res = append(res, s.Parts.Publish.AssembleDefinition(ctx)...)
	res = append(res, s.Parts.Subscribe.AssembleDefinition(ctx)...)
	res = append(res, s.Parts.Common.AssembleDefinition(ctx)...)

	return res
}

func (s Server) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	panic("not implemented")
}
