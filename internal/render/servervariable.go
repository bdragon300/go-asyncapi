package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

// ServerVariable represents the server variable object.
type ServerVariable struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the server variable as it was defined in the AsyncAPI document.
	OriginalName string
	// Description is an optional server variable description. Renders as Go doc comment.
	Description string // TODO
	// Enum is enum of possible values.
	Enum []string // TODO: implement validation
	// Default is the default value of the server variable.
	Default string
}

func (s *ServerVariable) Name() string {
	return s.OriginalName
}

func (s *ServerVariable) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (s *ServerVariable) Selectable() bool {
	return false
}

func (s *ServerVariable) Visible() bool {
	return true
}

func (s *ServerVariable) String() string {
	return "ServerVariable(" + s.OriginalName + ")"
}
