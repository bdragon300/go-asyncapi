package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

// MessageTrait represents a message trait object
type MessageTrait struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the message trait itself, typically this is a key where the message trait is registered in the document.
	OriginalName string
	// ContentType is the message trait's content type if set.
	ContentType string

	// HeadersTypePromise is a Go struct for message trait headers. Nil if headers type is not set in document.
	HeadersTypePromise *lang.GolangTypePromise

	// BindingsPromise is a promise to message trait bindings contents. Nil if message trait bindings are not set.
	BindingsPromise *lang.Promise[*Bindings]

	// CorrelationIDPromise is a CorrelationID object defined for the message trait. Nil if correlationID is not defined.
	CorrelationIDPromise *lang.Promise[*CorrelationID]
}

// HeadersType returns a Go type of headers defined for message trait in the document.
// If headers is not set, returns nil.
func (m *MessageTrait) HeadersType() common.GolangType {
	if m.HeadersTypePromise != nil {
		return common.DerefArtifact[common.GolangType](m.HeadersTypePromise.T())
	}
	return nil
}

// HasHeaders returns true if the message trait has headers defined in the document, false otherwise.
func (m *MessageTrait) HasHeaders() bool {
	return m.HeadersTypePromise != nil
}

// Bindings returns the Bindings object or nil if no bindings are set.
func (m *MessageTrait) Bindings() *Bindings {
	if m.BindingsPromise != nil {
		return m.BindingsPromise.T()
	}
	return nil
}

// CorrelationID returns the CorrelationID object or nil if no correlationID is set.
func (m *MessageTrait) CorrelationID() *CorrelationID {
	if m.CorrelationIDPromise != nil {
		return m.CorrelationIDPromise.T()
	}
	return nil
}

func (m *MessageTrait) Name() string {
	return m.OriginalName
}

func (m *MessageTrait) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (m *MessageTrait) Selectable() bool {
	return false // MessageTrait only affects to the Message rendering and does not produce additional code
}

func (m *MessageTrait) Visible() bool {
	return false
}

func (m *MessageTrait) String() string {
	return "MessageTrait(" + m.OriginalName + ")"
}
