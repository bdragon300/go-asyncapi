package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	j "github.com/dave/jennifer/jen"
)

type Server struct {
	Name            string
	Protocol        string
	ProtoServer     common.Renderer
	BindingsStruct  *GoStruct           // nil if bindings are not defined for server
	BindingsPromise *Promise[*Bindings] // nil if bindings are not defined for server as well
}

func (s Server) DirectRendering() bool {
	return true
}

func (s Server) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	ctx.LogRender("Server", "", s.Name, "definition", s.DirectRendering())
	defer ctx.LogReturn()

	// Bindings struct and its methods according to server protocol
	if s.BindingsStruct != nil {
		res = append(res, s.BindingsStruct.RenderDefinition(ctx)...)

		if s.BindingsPromise != nil {
			tgt := s.BindingsPromise.Target()
			if r, ok := ctx.ProtoRenderers[s.Protocol]; ok {
				res = append(res, tgt.RenderBindingsMethod(ctx, s.BindingsStruct, s.Protocol, r.ProtocolAbbreviation())...)
			} else {
				ctx.Logger.Warnf("Skip protocol %q, since it is not supported", s.Protocol)
			}
		}
	}
	res = append(res, s.ProtoServer.RenderDefinition(ctx)...)
	return res
}

func (s Server) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (s Server) String() string {
	return s.Name
}
