package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoMap is a type representing a Go map.
type GoMap struct {
	BaseType
	// KeyType is the type of the map key.
	KeyType common.GolangType
	// ValueType is the type of the map value.
	ValueType common.GolangType

	StructFieldRenderInfo StructFieldRenderInfo
}

func (m *GoMap) String() string {
	if m.Import != "" {
		return fmt.Sprintf("GoMap(%s.%s)", m.Import, m.OriginalName)
	}
	return "GoMap(" + m.OriginalName + ")"
}

func (m *GoMap) GoTemplate() string {
	return "code/lang/gomap"
}

func (m *GoMap) StructRenderInfo() StructFieldRenderInfo {
	return m.StructFieldRenderInfo
}
