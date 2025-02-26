package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

// Parameter represents a channel parameter object.
type Parameter struct {
	// OriginalName is the name of the parameter as it was defined in the AsyncAPI document.
	OriginalName string
	// Type is the Go type of the parameter. Usually, it's ``string''.
	Type common.GolangType
}

func (p *Parameter) Name() string {
	return p.OriginalName
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

func (p *Parameter) String() string {
	return "Parameter " + p.OriginalName
}
