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
	return common.ObjectKindServerVariable
}

func (s *ServerVariable) Selectable() bool {
	return false
}

func (s *ServerVariable) GetOriginalName() string {
	return s.OriginalName
}

//
//func (s ServerVariable) D(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}
//
//func (s ServerVariable) U(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}
//
//func (s ServerVariable) ID() string {
//	return s.GetOriginalName
//}
//
func (s *ServerVariable) String() string {
	return "ServerVariable " + s.OriginalName
}
