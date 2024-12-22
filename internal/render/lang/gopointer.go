package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoPointer struct {
	Type   common.GolangType
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

func (p *GoPointer) GoTemplate() string {
	return "lang/gopointer"
}

func (p *GoPointer) String() string {
	return "GoPointer -> " + p.Type.String()
}

func (p *GoPointer) Name() string {
	return p.Type.Name()
}

func (p *GoPointer) UnwrapGolangType() (common.GolangType, bool) {
	if v, ok := p.Type.(GolangTypeWrapperType); ok {
		return v.UnwrapGolangType()
	}
	return p.Type, p.Type != nil
}

func (p *GoPointer) IsPointer() bool {
	return true
}