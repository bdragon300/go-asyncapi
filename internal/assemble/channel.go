package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/utils"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Channel struct {
	Name                     string
	AppliedServers           []string
	AppliedServerLinks       []*Link[*Server] // Avoid using a map to keep definition order in generated code
	AppliedToAllServersLinks *LinkList[*Server]
	AllProtocols             map[string]common.Assembler

	ParametersStruct *Struct // nil if no parameters
	BindingsStruct   *Struct // nil if bindings are not set
}

func (c Channel) AllowRender() bool {
	return true
}

func (c Channel) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement

	if c.ParametersStruct != nil {
		res = append(res, c.ParametersStruct.AssembleDefinition(ctx)...)
	}
	if c.BindingsStruct != nil {
		res = append(res, c.BindingsStruct.AssembleDefinition(ctx)...)
	}
	res = append(res, c.assembleChannelNameFunc(ctx)...)

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
			panic(fmt.Sprintf("%q protocol is not supported", p))
		}
		res = append(res, r.AssembleDefinition(ctx)...)
	}
	return res
}

func (c Channel) AssembleUsage(_ *common.AssembleContext) []*j.Statement {
	panic("not implemented")
}

func (c Channel) assembleChannelNameFunc(ctx *common.AssembleContext) []*j.Statement {
	// Channel1Name(params Chan1Parameters) runtime.ParamString
	return []*j.Statement{
		j.Func().Id(utils.ToGolangName(c.Name, true)+"Name").
			ParamsFunc(func(g *j.Group) {
				if c.ParametersStruct != nil {
					g.Id("params").Add(utils.ToCode(c.ParametersStruct.AssembleUsage(ctx))...)
				}
			}).
			Qual(ctx.RuntimePackage(""), "ParamString").
			BlockFunc(func(blockGroup *j.Group) {
				if c.ParametersStruct == nil {
					blockGroup.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(c.Name),
					}))
				} else {
					blockGroup.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, f := range c.ParametersStruct.Fields {
							d[j.Id("params").Dot(f.Name).Dot("Name").Call()] = j.Id("params").Dot(f.Name).Dot("String").Call()
						}
					}))
					blockGroup.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(c.Name),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				}
			}),
	}
}
