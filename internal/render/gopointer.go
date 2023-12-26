package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
)

type Pointer struct {
	Type         common.GolangType
	DirectRender bool
}

func (p Pointer) DirectRendering() bool {
	return p.DirectRender
}

func (p Pointer) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("Pointer", "", "", "definition", p.DirectRendering())
	defer ctx.LogReturn()

	return p.Type.RenderDefinition(ctx)
}

func (p Pointer) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	ctx.LogRender("Pointer", "", "", "usage", p.DirectRendering())
	defer ctx.LogReturn()

	isPtr := true
	switch v := p.Type.(type) {
	case *Interface: // Prevent pointer to interface
		isPtr = false
	case pointerGolangType:
		isPtr = !v.IsPointer() // Prevent appearing pointer to pointer
	case *Simple:
		isPtr = !v.IsIface
	}
	if isPtr {
		return []*jen.Statement{jen.Op("*").Add(utils.ToCode(p.Type.RenderUsage(ctx))...)}
	}
	return p.Type.RenderUsage(ctx)
}

func (p Pointer) TypeName() string {
	return p.Type.TypeName()
}

func (p Pointer) String() string {
	return p.Type.String()
}

func (p Pointer) WrappedGolangType() (common.GolangType, bool) {
	return p.Type, p.Type != nil
}

func (p Pointer) IsPointer() bool {
	return true
}
