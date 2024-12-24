package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

type Server struct {
	OriginalName string
	Dummy          bool
	IsComponent bool // true if server is defined in `components` section

	URL             string
	Protocol        string
	ProtocolVersion string

	VariablesPromises  types.OrderedMap[string, *lang.Promise[*ServerVariable]]
	// All channels defined in `channel` section this server is applied to. Can be *Channel or promise to *Channel
	AllChannelsPromise *lang.ListPromise[common.Renderable]

	BindingsType    *lang.GoStruct           // nil if bindings are not defined for server
	BindingsPromise *lang.Promise[*Bindings] // nil if bindings are not defined for server as well

	ProtoServer *ProtoServer // nil if server is dummy or has unsupported protocol
}

func (s *Server) Kind() common.ObjectKind {
	return common.ObjectKindServer
}

func (s *Server) Selectable() bool {
	return !s.Dummy && !s.IsComponent // Select only the servers defined in the `channels` section`
}

func (s *Server) Visible() bool {
	return !s.Dummy
}

func (s *Server) SelectProtoObject(protocol string) common.Renderable {
	if s.ProtoServer.Selectable() && s.ProtoServer.Protocol == protocol {
		return s.ProtoServer
	}
	return nil
}

func (s *Server) Name() string {
	return utils.CapitalizeUnchanged(s.OriginalName)
}

func (s *Server) GetBoundChannels() []common.Renderable {
	type renderableWrapper interface {
		UnwrapRenderable() common.Renderable
	}

	currentName := common.GetContext().GetObjectName(s)
	r := lo.Filter(s.AllChannelsPromise.T(), func(r common.Renderable, _ int) bool {
		if !r.Visible() {
			return false
		}
		if w, ok := r.(renderableWrapper); ok {
			r = w.UnwrapRenderable()
		}
		if ch, ok := r.(*Channel); ok {
			// Empty/omitted servers field in channel means "all servers"
			return len(ch.BoundServerNames) == 0 || lo.Contains(ch.BoundServerNames, currentName)
		}
		return false
	})
	return r
}

//func (p *ProtoServer) GetBoundProtoChannels(protoName string) []*ProtoChannel {
//	channels := p.GetBoundChannels()
//	res := lo.FlatMap(channels, func(ch common.Renderable, _ int) []*ProtoChannel {
//		return lo.Map(ch.SelectProtoObject([]string{p.Protocol}), func(item common.Renderable, _ int) *ProtoChannel {
//			return item.(*ProtoChannel)
//		})
//	})
//	return res
//}

//func (s Server) D(ctx *common.RenderContext) []*j.Statement {
//	var res []*j.Statement
//	ctx.LogStartRender("Server", "", s.GetOriginalName, "definition", s.Selectable())
//	defer ctx.LogFinishRender()
//
//	if s.ProtocolVersion != "" {
//		res = append(res, s.RenderProtocolVersionConst(ctx)...)
//	}
//	res = append(res, s.RenderURLFunc(ctx)...)
//
//	// Bindings struct and its methods according to server protocol
//	if s.BindingsType != nil {
//		res = append(res, s.BindingsType.D(ctx)...)
//
//		if s.BindingsPromise != nil {
//			tgt := s.BindingsPromise.Target()
//			if r, ok := ctx.ProtoRenderers[s.Protocol]; ok {
//				res = append(res, tgt.RenderBindingsMethod(ctx, s.BindingsType, s.Protocol, r.ProtocolTitle())...)
//			} else {
//				ctx.Logger.Warnf("Skip protocol %q, since it is not supported", s.Protocol)
//			}
//		}
//	}
//	if s.ProtoServer != nil {
//		res = append(res, s.ProtoServer.D(ctx)...)
//	}
//	return res
//}

//func (s Server) U(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}

//func (s Server) RenderProtocolVersionConst(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderProtocolVersionConst")
//
//	return []*j.Statement{
//		j.Const().Id(s.TypeNamePrefix + "ProtocolVersion").Op("=").Lit(s.ProtocolVersion),
//	}
//}

//func (s Server) RenderURLFunc(ctx *common.RenderContext) []*j.Statement {
//	ctx.Logger.Trace("RenderURLFunc")
//
//	// Server1URL(param1 string, param2 string) run.ParamString
//	return []*j.Statement{
//		j.Func().Id(s.TypeNamePrefix+"URL").
//			ParamsFunc(func(g *j.Group) {
//				for _, entry := range s.VariablesPromises.Entries() {
//					g.Id(utils.ToGolangName(entry.Key, false)).String()
//				}
//			}).
//			Qual(ctx.RuntimeModule(""), "ParamString").
//			BlockFunc(func(bg *j.Group) {
//				if s.VariablesPromises.Len() > 0 {
//					for _, entry := range s.VariablesPromises.Entries() {
//						param := utils.ToGolangName(entry.Key, false)
//						if entry.Value.Target().Default != "" {
//							bg.If(j.Id(param).Op("==").Lit("")).
//								Block(
//									j.Id(param).Op("=").Lit(entry.Value.Target().Default),
//								)
//						}
//					}
//					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
//						for _, entry := range s.VariablesPromises.Entries() {
//							d[j.Lit(entry.Key)] = j.Id(entry.Value.Target().GetOriginalName)
//						}
//					}))
//					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
//						j.Id("Expr"):       j.Lit(s.URL),
//						j.Id("Parameters"): j.Id("paramMap"),
//					}))
//				} else {
//					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
//						j.Id("Expr"): j.Lit(s.URL),
//					}))
//				}
//			}),
//	}
//}

func (s *Server) String() string {
	return fmt.Sprintf("Server[%s] %s", s.Protocol, s.OriginalName)
}

func (s *Server) BindingsProtocols() (res []string) {
	if s.BindingsType == nil {
		return nil
	}
	if s.BindingsPromise != nil {
		res = append(res, s.BindingsPromise.T().Values.Keys()...)
		res = append(res, s.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

func (s *Server) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ServerBindings", Import: common.GetContext().RuntimeModule(protoName)},
		EmptyCurlyBrackets: true,
	}
	if s.BindingsPromise != nil {
		if b, ok := s.BindingsPromise.T().Values.Get(protoName); ok {
			//ctx.Logger.Debug("Server bindings", "proto", protoName)
			res = b
		}
	}
	return res
}

func (s *Server) Variables() (res types.OrderedMap[string, *ServerVariable]) {
	for _, entry := range s.VariablesPromises.Entries() {
		res.Set(entry.Key, entry.Value.T())
	}
	return
}

func (s *Server) AllChannels() (res []common.Renderable) {
	return s.AllChannelsPromise.T()
}

func (s *Server) Bindings() (res *Bindings) {
	if s.BindingsPromise != nil {
		return s.BindingsPromise.T()
	}
	return nil
}

type ProtoServer struct {
	*Server
	Type *lang.GoStruct // Nil if server is dummy or has unsupported protocol
}

func (p *ProtoServer) String() string {
	return "ProtoServer " + p.OriginalName
}

func (p *ProtoServer) Selectable() bool {
	return !p.Dummy
}
