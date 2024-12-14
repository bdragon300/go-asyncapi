package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoMap struct {
	BaseType
	KeyType   common.GolangType
	ValueType common.GolangType
}

func (m *GoMap) GoTemplate() string {
	return "lang/gomap"
}

func (m *GoMap) String() string {
	if m.Import != "" {
		return "GoMap /" + m.Import + "." + m.OriginalName
	}
	return "GoMap " + m.OriginalName
}