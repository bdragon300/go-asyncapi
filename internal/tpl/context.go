package tpl

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"reflect"
)

type renderContext interface {
	Imports() []common.ImportItem
	CurrentSelection() common.RenderSelectionConfig
}

func NewTemplateContext(renderContext renderContext, object common.CompileObject, index int) TemplateContext {
	return TemplateContext{
		renderContext: renderContext,
		object:        object,
		index:         index,
	}
}

// TemplateContext is passed as value to the root template on selections processing.
type TemplateContext struct {
	renderContext renderContext
	object        common.CompileObject
	index         int
}

func (t TemplateContext) Imports() []common.ImportItem {
	return t.renderContext.Imports()
}

func (t TemplateContext) CurrentSelection() common.RenderSelectionConfig {
	return t.renderContext.CurrentSelection()
}

func (t TemplateContext) Index() int {
	return t.index
}

func (t TemplateContext) OtherObj() common.Renderable {
	if t.object.Kind() == common.ObjectKindOther {
		return t.object
	}
	return nil
}

func (t TemplateContext) SchemaObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindSchema)
}

func (t TemplateContext) ServerObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindServer)
}

func (t TemplateContext) ServerVariableObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindServerVariable)
}

func (t TemplateContext) ChannelObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindChannel)
}

func (t TemplateContext) MessageObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindMessage)
}

func (t TemplateContext) ParametersObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindParameter)
}

func (t TemplateContext) CorrelationIDObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindCorrelationID)
}

func (t TemplateContext) AsyncAPIObj() common.Renderable {
	return retrieveObject(t.object, common.ObjectKindAsyncAPI)
}

func retrieveObject(obj common.Renderable, kind common.ObjectKind) common.Renderable {
	if obj.Kind() != kind {
		return nil
	}
	// Unwrap promise(s) until we get the actual object
	for {
		v, ok := obj.(common.ObjectPromise)
		if !ok {
			break
		}
		obj = reflect.ValueOf(v).MethodByName("T").Call(nil)[0].Interface().(common.Renderable)
	}
	return obj
}
