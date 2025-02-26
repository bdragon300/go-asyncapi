package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoPointer is a type representing a pointer type to another Go type. It acts as a wrapper for the inner type.
type GoPointer struct {
	Type common.GolangType
}

func (p *GoPointer) Name() string {
	return p.Type.Name()
}

func (p *GoPointer) Kind() common.ObjectKind {
	return p.Type.Kind()
}

func (p *GoPointer) Selectable() bool {
	return p.Type.Selectable()
}

func (p *GoPointer) Visible() bool {
	return p.Type.Visible()
}

func (p *GoPointer) String() string {
	return "GoPointer -> " + p.Type.String()
}

func (p *GoPointer) CanBeAddressed() bool {
	// Prevent appearing pointer to pointer in type definition.
	// Also, taking the address of pointer typically is not useful anywhere in the generated code
	return false
}

func (p *GoPointer) CanBeDereferenced() bool {
	return true
}

func (p *GoPointer) GoTemplate() string {
	return "code/lang/gopointer"
}

func (p *GoPointer) InnerGolangType() common.GolangType {
	if v, ok := p.Type.(GolangTypeExtractor); ok {
		return v.InnerGolangType()
	}
	return p.Type
}
