package tmpl

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type renderContext interface {
	CurrentSelection() common.ConfigSelectionItem
	Package() string
}

type importsProvider interface {
	Imports() []common.ImportItem
}

func NewTemplateContext(renderContext renderContext, object common.Renderable, importsProvider importsProvider) *TemplateContext {
	return &TemplateContext{
		renderContext:   renderContext,
		importsProvider: importsProvider,
		object:          object,
	}
}

// TemplateContext is passed as value to the root template on selections processing.
type TemplateContext struct {
	renderContext renderContext
	object          common.Renderable
	importsProvider importsProvider
}

func (t TemplateContext) Imports() []common.ImportItem {
	return t.importsProvider.Imports()
}

func (t TemplateContext) PackageName() string {
	return t.renderContext.Package()
}

func (t TemplateContext) CurrentSelection() common.ConfigSelectionItem {
	return t.renderContext.CurrentSelection()
}

func (t TemplateContext) Object() common.Renderable {
	return t.object
}

type ImplementationsTemplateContext struct {
	Protocol string
	Name     string
}