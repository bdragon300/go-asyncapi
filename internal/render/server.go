package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	j "github.com/dave/jennifer/jen"
)

type Server struct {
	Name         string
	DirectRender bool // Typically, it's true if server is defined in `servers` section, false if in `components` section
	Dummy        bool
	Protocol     string
	ProtoServer  common.Renderer // nil if protocol is not supported by the tool

	Variables types.OrderedMap[string, *Promise[*ServerVariable]]

	BindingsStruct  *GoStruct           // nil if bindings are not defined for server
	BindingsPromise *Promise[*Bindings] // nil if bindings are not defined for server as well
}

func (s Server) DirectRendering() bool {
	return s.DirectRender && !s.Dummy
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
				res = append(res, tgt.RenderBindingsMethod(ctx, s.BindingsStruct, s.Protocol, r.ProtocolTitle())...)
			} else {
				ctx.Logger.Warnf("Skip protocol %q, since it is not supported", s.Protocol)
			}
		}
	}
	if s.ProtoServer != nil {
		res = append(res, s.ProtoServer.RenderDefinition(ctx)...)
	}
	return res
}

func (s Server) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (s Server) ID() string {
	return s.Name
}

func (s Server) String() string {
	return "Server " + s.Name
}
