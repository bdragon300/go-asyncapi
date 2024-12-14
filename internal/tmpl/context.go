package tmpl

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type renderContext interface {
	CurrentSelection() common.RenderSelectionConfig
}

type fileHeaderStorage interface {
	Imports() []common.ImportItem
	PackageName() string
}

func NewTemplateContext(renderContext renderContext, object common.Renderable, headerStorage fileHeaderStorage) *TemplateContext {
	return &TemplateContext{
		renderContext:     renderContext,
		fileHeaderStorage: headerStorage,
		object:            object,
	}
}

// TemplateContext is passed as value to the root template on selections processing.
type TemplateContext struct {
	renderContext renderContext
	// object contains an actual object to be rendered, without any promises.
	object            common.Renderable
	fileHeaderStorage fileHeaderStorage
}

func (t TemplateContext) Imports() []common.ImportItem {
	return t.fileHeaderStorage.Imports()
}

func (t TemplateContext) PackageName() string {
	return t.fileHeaderStorage.PackageName()
}

func (t TemplateContext) CurrentSelection() common.RenderSelectionConfig {
	return t.renderContext.CurrentSelection()
}

func (t TemplateContext) Object() common.Renderable {
	if t.object == nil {
		return nil
	}
	return t.object
}
