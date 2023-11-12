package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/dave/jennifer/jen"
)

type Server struct {
	Name           string
	Protocol       string
	ProtoServer    common.Renderer
	BindingsStruct *Struct // nil if no bindings set in spec
}

func (s Server) DirectRendering() bool {
	return true
}

func (s Server) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	ctx.LogRender("Server", "", s.Name, "definition", s.DirectRendering())
	defer ctx.LogReturn()

	res = append(res, s.BindingsStruct.RenderDefinition(ctx)...)
	res = append(res, s.ProtoServer.RenderDefinition(ctx)...)
	return res
}

func (s Server) RenderUsage(_ *common.RenderContext) []*jen.Statement {
	panic("not implemented")
}

func (s Server) String() string {
	return s.Name
}
