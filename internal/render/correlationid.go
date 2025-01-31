package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type CorrelationIDStructFieldKind string

const (
	CorrelationIDStructFieldKindPayload CorrelationIDStructFieldKind = "payload"
	CorrelationIDStructFieldKindHeaders CorrelationIDStructFieldKind = "headers"
)

// CorrelationID never renders itself, only as a part of message struct
type CorrelationID struct {
	OriginalName    string
	Description     string
	StructFieldKind CorrelationIDStructFieldKind // Type field name to store the value to or to load the value from
	LocationPath    []string                     // JSONPointer path to the field in the message, should be non-empty
	Dummy           bool
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

func (c *CorrelationID) Name() string {
	return c.OriginalName
}

func (c *CorrelationID) String() string {
	return "CorrelationID " + c.OriginalName
}
