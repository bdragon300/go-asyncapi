package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

// CorrelationID represents the correlation ID object.
type CorrelationID struct {
	lang.BaseJSONPointed
	lang.BaseRuntimeExpression
	// OriginalName is the name of the correlation ID as it was defined in the AsyncAPI document.
	OriginalName string
	// Description is an optional correlation ID description. Renders as Go doc comment.
	Description string
	// Dummy is true when correlation ID object is ignored (x-ignore: true)
	Dummy bool
}

func (c *CorrelationID) Name() string {
	return c.OriginalName
}

func (c *CorrelationID) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (c *CorrelationID) Selectable() bool {
	return false
}

func (c *CorrelationID) Visible() bool {
	return !c.Dummy
}

func (c *CorrelationID) String() string {
	return "CorrelationID(" + c.OriginalName + ")"
}
