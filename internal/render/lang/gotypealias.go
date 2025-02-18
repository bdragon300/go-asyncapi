package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoTypeAlias struct {
	BaseType
	AliasedType common.GolangType
}

func (p *GoTypeAlias) GoTemplate() string {
	return "code/lang/gotypealias"
}

func (p *GoTypeAlias) InnerGolangType() common.GolangType {
	if v, ok := p.AliasedType.(GolangTypeExtractor); ok {
		return v.InnerGolangType()
	}
	return p.AliasedType
}

func (p *GoTypeAlias) Addressable() bool {
	return true // In fact, type alias is a new type, so it is addressable by default, even if aliased type is not (e.g. interface)
}

func (p *GoTypeAlias) IsPointer() bool {
	return false // Type alias is not a pointer itself
}

func (p *GoTypeAlias) IsStruct() bool {
	if v, ok := any(p.AliasedType).(golangStructType); ok {
		return v.IsStruct()
	}
	return false
}

func (p *GoTypeAlias) String() string {
	if p.Import != "" {
		return fmt.Sprintf("GoTypeAlias /%s.%s -> %s", p.Import, p.OriginalName, p.AliasedType)
	}
	return fmt.Sprintf("GoTypeAlias %s -> %s", p.OriginalName, p.AliasedType)
}
