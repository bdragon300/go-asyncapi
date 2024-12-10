package tpl

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type renderContext interface {
	Imports() []common.ImportItem
	CurrentSelection() common.RenderSelectionConfig
}

func NewTemplateContext(renderContext renderContext, object common.CompileObject, selectionIndex, overallIndex int) TemplateContext {
	return TemplateContext{
		renderContext:  renderContext,
		object:         object,
		selectionIndex: selectionIndex,
		overallIndex: overallIndex,
	}
}

// TemplateContext is passed as value to the root template on selections processing.
type TemplateContext struct {
	renderContext renderContext
	object         common.CompileObject
	selectionIndex int
	overallIndex int
}

func (t TemplateContext) Imports() []common.ImportItem {
	return t.renderContext.Imports()
}

func (t TemplateContext) CurrentSelection() common.RenderSelectionConfig {
	return t.renderContext.CurrentSelection()
}

func (t TemplateContext) SelectionIndex() int {
	return t.selectionIndex
}

func (t TemplateContext) OverallIndex() int {
	return t.overallIndex
}

func (t TemplateContext) Object() any {
	return t.object.Renderable
}
