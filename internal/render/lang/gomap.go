package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoMap is a type representing a Go map.
type GoMap struct {
	BaseType
	// KeyType is the type of the map key.
	KeyType common.GolangType
	// ValueType is the type of the map value.
	ValueType common.GolangType
}

func (m *GoMap) String() string {
	if m.Import != "" {
		return "GoMap /" + m.Import + "." + m.OriginalName
	}
	return "GoMap " + m.OriginalName
}

func (m *GoMap) GoTemplate() string {
	return "code/lang/gomap"
}
