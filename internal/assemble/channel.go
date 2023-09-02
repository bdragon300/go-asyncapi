package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/common"

	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type ChannelParts struct {
	Publish   common.Assembler
	Subscribe common.Assembler
	Common    common.Assembler
}

type Channel struct {
	Name                     string
	AppliedServers           []string
	AppliedServerLinks       []*LinkQuery[*Server] // Avoid using a map to keep definition order in generated code
	AppliedToAllServersLinks *LinkQueryList[*Server]
	SupportedProtocols       map[string]ChannelParts
}

func (c Channel) AllowRender() bool {
	return true
}

func (c Channel) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement

	protocols := lo.Uniq(lo.Map(c.AppliedServerLinks, func(item *LinkQuery[*Server], index int) string {
		return item.Link().Protocol
	}))
	if c.AppliedToAllServersLinks != nil {
		protocols = lo.Uniq(lo.Map(c.AppliedToAllServersLinks.Links(), func(item *Server, index int) string {
			return item.Protocol
		}))
	}
	for _, p := range protocols {
		if r, ok := c.SupportedProtocols[p]; ok {
			if r.Subscribe != nil {
				res = append(res, r.Subscribe.AssembleDefinition(ctx)...)
			}
			if r.Publish != nil {
				res = append(res, r.Publish.AssembleDefinition(ctx)...)
			}
			if r.Common != nil {
				res = append(res, r.Common.AssembleDefinition(ctx)...)
			}
		} else {
			panic(fmt.Sprintf("%q protocol is not supported", p))
		}
	}
	return res
}

func (c Channel) AssembleUsage(_ *common.AssembleContext) []*jen.Statement {
	panic("not implemented")
}
