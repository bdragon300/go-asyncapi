package lang

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Channel struct {
	AppliedServers      []*LinkerQuery[*Server]
	AppliedToAllServers *LinkerQueryList[*Server]
	SupportedProtocols  map[string]render.LangRenderer
}

func (c Channel) AllowRender() bool {
	return true
}

func (c Channel) RenderDefinition(ctx *render.Context) []*jen.Statement {
	var res []*jen.Statement

	protocols := lo.Uniq(lo.Map(c.AppliedServers, func(item *LinkerQuery[*Server], index int) string {
		return item.Link().Protocol
	}))
	if c.AppliedToAllServers != nil {
		protocols = lo.Uniq(lo.Map(c.AppliedToAllServers.Links(), func(item *Server, index int) string {
			return item.Protocol
		}))
	}
	for _, p := range protocols {
		if r, ok := c.SupportedProtocols[p]; ok {
			res = append(res, r.RenderDefinition(ctx)...)
		} else {
			panic(fmt.Sprintf("%q protocol is not supported", p))
		}
	}
	return res
}

func (c Channel) RenderUsage(_ *render.Context) []*jen.Statement {
	panic("not implemented")
}
