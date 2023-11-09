package assemble

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/dave/jennifer/jen"
)

type Server struct {
	Name           string
	Protocol       string
	ProtoServer    common.Assembler
	BindingsStruct *Struct // nil if no bindings set in spec
}

func (s Server) AllowRender() bool {
	return true
}

func (s Server) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	res = append(res, s.BindingsStruct.AssembleDefinition(ctx)...)
	res = append(res, s.ProtoServer.AssembleDefinition(ctx)...)
	return res
}

func (s Server) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	panic("not implemented")
}

func (s Server) String() string {
	return "Server " + s.Name
}
