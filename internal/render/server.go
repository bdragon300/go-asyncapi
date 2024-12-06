package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type Server struct {
	Name           string
	SpecKey        string // Name as it is in the source document, without considering `x-go-name` extension
	TypeNamePrefix string // Name of server struct
	Dummy          bool

	URL             string
	Protocol        string
	ProtocolVersion string

	VariablesPromises  types.OrderedMap[string, *lang.Promise[*ServerVariable]]
	AllChannelsPromise *lang.ListPromise[*Channel]

	BindingsType    *lang.GoStruct           // nil if bindings are not defined for server
	BindingsPromise *lang.Promise[*Bindings] // nil if bindings are not defined for server as well
}

func (s Server) Kind() common.ObjectKind {
	return common.ObjectKindServer
}

func (s Server) Selectable() bool {
	return !s.Dummy
}

//func (s Server) D(ctx *common.RenderContext) []*j.Statement {
//	var res []*j.Statement
//	ctx.LogStartRender("Server", "", s.Name, "definition", s.Selectable())
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
//							d[j.Lit(entry.Key)] = j.Id(entry.Value.Target().Name)
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

//func (s Server) ID() string {
//	return s.Name
//}
//
func (s Server) String() string {
	return "Server " + s.Name
}

func (s Server) GetRelevantChannels() []*Channel {
	return lo.FilterMap(s.AllChannelsPromise.T(), func(p *Channel, _ int) (*Channel, bool) {
		// Empty/omitted servers field in channel means "all servers"
		ok := len(p.SpecServerNames) == 0 || lo.Contains(p.SpecServerNames, s.SpecKey)
		return p, ok && !p.Dummy
	})
}

func (c Server) BindingsProtocols() (res []string) {
	if c.BindingsPromise != nil {
		res = append(res, c.BindingsPromise.T().Values.Keys()...)
		res = append(res, c.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

func (c Server) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{Name: "ServerBindings", Import: context.Context.RuntimeModule(protoName)},
		EmptyCurlyBrackets: true,
	}
	if c.BindingsPromise != nil {
		if b, ok := c.BindingsPromise.T().Values.Get(protoName); ok {
			//ctx.Logger.Debug("Server bindings", "proto", protoName)
			res = b
		}
	}
	return res
}

type ProtoServer struct {
	*Server
	Type *lang.GoStruct // Nil if server is dummy or has unsupported protocol

	ProtoName string
}

func (p ProtoServer) String() string {
	return "ProtoServer " + p.Name
}