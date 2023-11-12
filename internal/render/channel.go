package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"

	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Channel struct {
	Name                     string
	AppliedServers           []string
	AppliedServerLinks       []*Link[*Server] // Avoid using a map to keep definition order in generated code
	AppliedToAllServersLinks *LinkList[*Server]
	AllProtocols             map[string]common.Renderer

	ParametersStruct *Struct // nil if no parameters
	BindingsStruct   *Struct // nil if bindings are not set
}

func (c Channel) DirectRendering() bool {
	return true
}

func (c Channel) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	ctx.LogRender("Channel", "", c.Name, "definition", c.DirectRendering())
	defer ctx.LogReturn()

	if c.ParametersStruct != nil {
		res = append(res, c.ParametersStruct.RenderDefinition(ctx)...)
	}
	if c.BindingsStruct != nil {
		res = append(res, c.BindingsStruct.RenderDefinition(ctx)...)
	}
	res = append(res, c.renderChannelNameFunc(ctx)...)

	protocols := lo.Uniq(lo.Map(c.AppliedServerLinks, func(item *Link[*Server], index int) string {
		return item.Target().Protocol
	}))
	if c.AppliedToAllServersLinks != nil {
		protocols = lo.Uniq(lo.Map(c.AppliedToAllServersLinks.Targets(), func(item *Server, index int) string {
			return item.Protocol
		}))
	}
	for _, p := range protocols {
		r, ok := c.AllProtocols[p]
		if !ok {
			ctx.Logger.Warnf("Skip protocol %q since it is not supported", p)
			continue
		}
		res = append(res, r.RenderDefinition(ctx)...)
	}
	return res
}

func (c Channel) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (c Channel) String() string {
	return "Channel " + c.Name
}

func (c Channel) renderChannelNameFunc(ctx *common.RenderContext) []*j.Statement {
	// Channel1Name(params Chan1Parameters) runtime.ParamString
	return []*j.Statement{
		j.Func().Id(utils.ToGolangName(c.Name, true)+"Name").
			ParamsFunc(func(g *j.Group) {
				if c.ParametersStruct != nil {
					g.Id("params").Add(utils.ToCode(c.ParametersStruct.RenderUsage(ctx))...)
				}
			}).
			Qual(ctx.RuntimePackage(""), "ParamString").
			BlockFunc(func(bg *j.Group) {
				if c.ParametersStruct == nil {
					bg.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(c.Name),
					}))
				} else {
					bg.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, f := range c.ParametersStruct.Fields {
							d[j.Id("params").Dot(f.Name).Dot("Name").Call()] = j.Id("params").Dot(f.Name).Dot("String").Call()
						}
					}))
					bg.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(c.Name),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				}
			}),
	}
}
