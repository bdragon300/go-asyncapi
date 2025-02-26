package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type CorrelationIDStructFieldKind string

const (
	CorrelationIDStructFieldKindPayload CorrelationIDStructFieldKind = "payload"
	CorrelationIDStructFieldKindHeaders CorrelationIDStructFieldKind = "headers"
)

// CorrelationID represents the correlation ID object.
type CorrelationID struct {
	// OriginalName is the name of the correlation ID as it was defined in the AsyncAPI document.
	OriginalName string
	// Description is an optional correlation ID description. Renders as Go doc comment.
	Description string
	// Dummy is true when correlation ID object is ignored (x-ignore: true)
	Dummy bool
	// StructFieldName describes which field in target message struct to use: payload or headers.
	StructFieldKind CorrelationIDStructFieldKind
	// LocationPath is JSONPointer fragment with field location in message, split in parts by "/".
	LocationPath []string
}

func (c *CorrelationID) Name() string {
	return c.OriginalName
}

func (c *CorrelationID) Kind() common.ObjectKind {
	return common.ObjectKindOther
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
