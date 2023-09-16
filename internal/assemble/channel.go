package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type Channel struct {
	Name                     string
	AppliedServers           []string
	AppliedServerLinks       []*Link[*Server] // Avoid using a map to keep definition order in generated code
	AppliedToAllServersLinks *LinkList[*Server]
	SupportedProtocols       map[string]common.Assembler
}

func (c Channel) AllowRender() bool {
	return true
}

func (c Channel) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement

	protocols := lo.Uniq(lo.Map(c.AppliedServerLinks, func(item *Link[*Server], index int) string {
		return item.Target().Protocol
	}))
	if c.AppliedToAllServersLinks != nil {
		protocols = lo.Uniq(lo.Map(c.AppliedToAllServersLinks.Targets(), func(item *Server, index int) string {
			return item.Protocol
		}))
	}
	for _, p := range protocols {
		r, ok := c.SupportedProtocols[p]
		if !ok {
			panic(fmt.Sprintf("%q protocol is not supported", p))
		}
		res = append(res, r.AssembleDefinition(ctx)...)
	}
	return res
}

func (c Channel) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	panic("not implemented")
}
