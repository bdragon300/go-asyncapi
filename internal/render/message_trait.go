package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

// MessageTrait represents a message trait object
type MessageTrait struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the message as it was defined in the AsyncAPI document.
	OriginalName string
	// ContentType is the message's content type if set.
	ContentType string

	// HeadersTypePromise is a Go struct for message headers. Nil if headers type is not set in document.
	HeadersTypePromise *lang.GolangTypePromise

	// BindingsPromise is a promise to message bindings contents. Nil if message bindings are not set.
	BindingsPromise *lang.Promise[*Bindings]

	// CorrelationIDPromise is a CorrelationID object defined for the message. Nil if correlationID is not defined.
	CorrelationIDPromise *lang.Promise[*CorrelationID]
}

func (m MessageTrait) Name() string {
	return m.OriginalName
}

func (m MessageTrait) Kind() common.ArtifactKind {
	return common.ArtifactKindOther // MessageTrait only affects to the building of the Message and does not render additional code
}

func (m MessageTrait) Selectable() bool {
	return false
}

func (m MessageTrait) Visible() bool {
	return false
}

func (m MessageTrait) String() string {
	return "MessageTrait(" + m.OriginalName + ")"
}
