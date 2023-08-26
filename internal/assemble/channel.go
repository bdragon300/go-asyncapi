package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Channel struct {
	AppliedServers      []*LinkQuery[*Server]
	AppliedToAllServers *LinkQueryList[*Server]
	SupportedProtocols  map[string]common.Assembled
}

func (c Channel) AllowRender() bool {
	return true
}

func (c Channel) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement

	protocols := lo.Uniq(lo.Map(c.AppliedServers, func(item *LinkQuery[*Server], index int) string {
		return item.Link().Protocol
	}))
	if c.AppliedToAllServers != nil {
		protocols = lo.Uniq(lo.Map(c.AppliedToAllServers.Links(), func(item *Server, index int) string {
			return item.Protocol
		}))
	}
	for _, p := range protocols {
		if r, ok := c.SupportedProtocols[p]; ok {
			res = append(res, r.AssembleDefinition(ctx)...)
		} else {
			panic(fmt.Sprintf("%q protocol is not supported", p))
		}
	}
	return res
}

func (c Channel) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	panic("not implemented")
}
