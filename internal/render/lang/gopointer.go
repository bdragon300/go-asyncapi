package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
)

type GoPointer struct {
	Type          common.GolangType
	// HasDefinition forces the GoPointer to be rendered as definition, even if type it points marked not to do so.
	HasDefinition bool
}

func (p GoPointer) Kind() common.ObjectKind {
	return p.Type.Kind()
}

func (p GoPointer) Selectable() bool {
	return p.HasDefinition
}

func (p GoPointer) RenderContext() common.RenderContext {
	return context.Context
}

func (p GoPointer) D() string {
	ctx.LogStartRender("GoPointer", "", "", "definition", p.IsDefinition())
	defer ctx.LogFinishRender()

	return p.Type.D()
}

func (p GoPointer) U() string {
	ctx.LogStartRender("GoPointer", "", "", "usage", p.IsDefinition())
	defer ctx.LogFinishRender()

	isPtr := true
	switch v := p.Type.(type) {
	case GolangPointerType:
		isPtr = !v.IsPointer() // Prevent appearing pointer to pointer
	case *GoSimple:
		isPtr = !v.IsInterface
	}
	if isPtr {
		return "*" + p.Type.U()
	}
	return p.Type.U()
}

func (p GoPointer) TypeName() string {
	return p.Type.TypeName()
}

func (p GoPointer) String() string {
	return "GoPointer -> " + p.Type.String()
}

func (p GoPointer) WrappedGolangType() (common.GolangType, bool) {
	return p.Type, p.Type != nil
}

func (p GoPointer) IsPointer() bool {
	return true
}
