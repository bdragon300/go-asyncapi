package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type ServerVariable struct {
	OriginalName string
	Description  string // TODO
	Enum        []string // TODO: implement validation
	Default     string
}

func (s *ServerVariable) Kind() common.ObjectKind {
	return common.ObjectKindOther
}

func (s *ServerVariable) Selectable() bool {
	return false
}

func (s *ServerVariable) Visible() bool {
	return true
}

func (s *ServerVariable) Name() string {
	return s.OriginalName
}

func (s *ServerVariable) String() string {
	return "ServerVariable " + s.OriginalName
}
