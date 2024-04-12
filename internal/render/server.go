package render

import (
	"sort"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Server struct {
	Name         string
	RawName      string // Name as it is in the source document, without considering `x-go-name` extension
	GolangName   string // Name of server struct
	DirectRender bool   // Typically, it's true if server is defined in `servers` section, false if in `components` section
	Dummy        bool

	URL             string
	Protocol        string
	ProtocolVersion string
	ProtoServer     common.Renderer // nil if protocol is not supported by the tool

	Variables           types.OrderedMap[string, *Promise[*ServerVariable]]
	AllChannelsPromises []*Promise[*Channel]

	BindingsStruct  *GoStruct           // nil if bindings are not defined for server
	BindingsPromise *Promise[*Bindings] // nil if bindings are not defined for server as well
}

func (s Server) DirectRendering() bool {
	return s.DirectRender && !s.Dummy
}

func (s Server) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	ctx.LogStartRender("Server", "", s.Name, "definition", s.DirectRendering())
	defer ctx.LogFinishRender()

	if s.ProtocolVersion != "" {
		res = append(res, s.RenderProtocolVersionConst(ctx)...)
	}
	res = append(res, s.RenderURLFunc(ctx)...)

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

func (s Server) RenderProtocolVersionConst(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderProtocolVersionConst")

	return []*j.Statement{
		j.Const().Id(s.GolangName + "ProtocolVersion").Op("=").Lit(s.ProtocolVersion),
	}
}

func (s Server) RenderURLFunc(ctx *common.RenderContext) []*j.Statement {
	ctx.Logger.Trace("RenderURLFunc")

	// Server1URL(param1 string, param2 string) run.ParamString
	return []*j.Statement{
		j.Func().Id(s.GolangName+"URL").
			ParamsFunc(func(g *j.Group) {
				for _, entry := range s.Variables.Entries() {
					g.Id(utils.ToGolangName(entry.Key, false)).String()
				}
			}).
			Qual(ctx.RuntimeModule(""), "ParamString").
			BlockFunc(func(bg *j.Group) {
				if s.Variables.Len() > 0 {
					for _, entry := range s.Variables.Entries() {
						param := utils.ToGolangName(entry.Key, false)
						if entry.Value.Target().Default != "" {
							bg.If(j.Id(param).Op("==").Lit("")).
								Block(
									j.Id(param).Op("=").Lit(entry.Value.Target().Default),
								)
						}
					}
					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, entry := range s.Variables.Entries() {
							d[j.Lit(entry.Key)] = j.Id(entry.Value.Target().Name)
						}
					}))
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(s.URL),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				} else {
					bg.Return(j.Qual(ctx.RuntimeModule(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(s.URL),
					}))
				}
			}),
	}
}

func (s Server) ID() string {
	return s.Name
}

func (s Server) String() string {
	return "Server " + s.Name
}

func (s Server) GetRelevantChannels() []*Channel {
	return lo.FilterMap(s.AllChannelsPromises, func(p *Promise[*Channel], _ int) (*Channel, bool) {
		// Empty/omitted servers field in channel means "all servers"
		ok := len(p.Target().ExplicitServerNames) == 0 || lo.Contains(p.Target().ExplicitServerNames, s.RawName)
		return p.Target(), ok && !p.Target().Dummy
	})
}

func getServerProtocols(ctx *common.RenderContext, promises []*Promise[*Server]) []string {
	res := lo.Uniq(lo.FilterMap(promises, func(item *Promise[*Server], _ int) (string, bool) {
		_, ok := ctx.ProtoRenderers[item.Target().Protocol]
		if !ok {
			ctx.Logger.Warnf("Skip protocol %q since it is not supported", item.Target().Protocol)
		}
		return item.Target().Protocol, ok && !item.Target().Dummy
	}))
	sort.Strings(res)
	return res
}
