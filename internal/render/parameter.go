package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

// Parameter represents a channel parameter object.
type Parameter struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the parameter as it was defined in the AsyncAPI document.
	OriginalName string
}

func (p *Parameter) Name() string {
	return p.OriginalName
}

func (p *Parameter) Kind() common.ArtifactKind {
	return common.ArtifactKindParameter
}

func (p *Parameter) Selectable() bool {
	return true
}

func (p *Parameter) Visible() bool {
	return true
}

func (p *Parameter) String() string {
	return "Parameter(" + p.OriginalName + ")"
}

func (p *Parameter) Pinnable() bool {
	return true
}
