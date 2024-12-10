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

func (p *GoPointer) D() string {
	//ctx.LogStartRender("GoPointer", "", "", "definition", p.IsDefinition())
	//defer ctx.LogFinishRender()
	return p.Type.D()
}

func (p *GoPointer) U() string {
	//ctx.LogStartRender("GoPointer", "", "", "usage", p.IsDefinition())
	//defer ctx.LogFinishRender()

	drawPtr := true
	if v, ok := p.Type.(GolangPointerType); ok {
		drawPtr = !v.IsPointer() // Prevent appearing pointer to pointer
	}
	if drawPtr {
		return "*" + p.Type.U()
	}
	return p.Type.U()
}

func (p *GoPointer) TypeName() string {
	return p.Type.TypeName()
}

func (p *GoPointer) String() string {
	return "GoPointer -> " + p.Type.String()
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

func (p *GoPointer) DefinitionInfo() (*common.GolangTypeDefinitionInfo, error) {
	return p.Type.DefinitionInfo()
}

func (p *GoPointer) SetDefinitionInfo(info *common.GolangTypeDefinitionInfo) {
	p.Type.SetDefinitionInfo(info)
}