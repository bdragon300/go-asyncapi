package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoTypeDefinition represents a Go type definition. I.e. the Go code like:
//
//	type Foo int
type GoTypeDefinition struct {
	BaseType
	RedefinedType common.GolangType
}

func (p *GoTypeDefinition) String() string {
	if p.Import != "" {
		return fmt.Sprintf("GoTypeDefinition(%s.%s)->%s", p.Import, p.OriginalName, p.RedefinedType)
	}
	return fmt.Sprintf("GoTypeDefinition(%s)->%s", p.OriginalName, p.RedefinedType)
}

func (p *GoTypeDefinition) CanBeAddressed() bool {
	return true // In fact, type alias is a new type, so it is addressable by default, even if aliased type is not (e.g. interface)
}

func (p *GoTypeDefinition) CanBeDereferenced() bool {
	return false // Type alias is not a pointer itself
}

func (p *GoTypeDefinition) GoTemplate() string {
	return "code/lang/gotypedefinition"
}

func (p *GoTypeDefinition) InnerGolangType() common.GolangType {
	if v, ok := p.RedefinedType.(GolangTypeExtractor); ok {
		return v.InnerGolangType()
	}
	return p.RedefinedType
}

func (p *GoTypeDefinition) IsStruct() bool {
	if v, ok := any(p.RedefinedType).(golangStructType); ok {
		return v.IsStruct()
	}
	return false
}
