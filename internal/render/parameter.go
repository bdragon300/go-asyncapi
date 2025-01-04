package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

type Parameter struct {
	OriginalName string
	Type         common.GolangType
}

func (p *Parameter) Kind() common.ObjectKind {
	return common.ObjectKindParameter
}

func (p *Parameter) Selectable() bool {
	return p.Type.Selectable()
}

func (p *Parameter) Visible() bool {
	return p.Type.Visible()
}

func (p *Parameter) Name() string {
	return utils.CapitalizeUnchanged(p.OriginalName)
}

func (p *Parameter) String() string {
	return "Parameter " + p.OriginalName
}
