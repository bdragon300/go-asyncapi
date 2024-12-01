package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/templates"
	"github.com/go-sprout/sprout"
	"github.com/samber/lo"
	"strings"
	"text/template"
)

var templateFunctions sprout.FunctionMap

func init() {
	handler := sprout.New()
	templateFunctions = handler.Build()
}

type TemplateSelections struct {
	Objects []common.Renderer
}

func (r TemplateSelections) SelectLangs() []common.Renderer {
	return lo.Filter(r.Objects, func(item common.Renderer, _ int) bool {
		return item.Kind() == common.ObjectKindLang
	})
}

func (r TemplateSelections) SelectSchemas() []*lang.GoStruct {
	return selectObjects[*lang.GoStruct](r.Objects, common.ObjectKindSchema)
}

func (r TemplateSelections) SelectServers() []*ProtoServer {
	return selectObjects[*ProtoServer](r.Objects, common.ObjectKindServer)
}

func (r TemplateSelections) SelectServerVariables() []*ServerVariable {
	return selectObjects[*ServerVariable](r.Objects, common.ObjectKindServerVariable)
}

func (r TemplateSelections) SelectChannels() []*ProtoChannel {
	return selectObjects[*ProtoChannel](r.Objects, common.ObjectKindChannel)
}

func (r TemplateSelections) SelectMessages() []*ProtoMessage {
	return selectObjects[*ProtoMessage](r.Objects, common.ObjectKindMessage)
}

func (r TemplateSelections) SelectParameters() []*Parameter {
	return selectObjects[*Parameter](r.Objects, common.ObjectKindParameter)
}

func (r TemplateSelections) SelectCorrelationIDs() []*CorrelationID {
	return selectObjects[*CorrelationID](r.Objects, common.ObjectKindCorrelationID)
}

func (r TemplateSelections) SelectServerBindings() []*Bindings {
	return selectObjects[*Bindings](r.Objects, common.ObjectKindServerBindings)
}

func (r TemplateSelections) SelectChannelBindings() []*Bindings {
	return selectObjects[*Bindings](r.Objects, common.ObjectKindChannelBindings)
}

func (r TemplateSelections) SelectOperationBindings() []*Bindings {
	return selectObjects[*Bindings](r.Objects, common.ObjectKindOperationBindings)
}

func (r TemplateSelections) SelectMessageBindings() []*Bindings {
	return selectObjects[*Bindings](r.Objects, common.ObjectKindMessageBindings)
}

func (r TemplateSelections) SelectAsyncAPI() *AsyncAPI {
	res := selectObjects[*AsyncAPI](r.Objects, common.ObjectKindAsyncAPI)
	if len(res) == 0 {
		return nil
	}
	return res[0]
}

func NewTemplateContext(renderContext *context.RenderContextImpl, selections TemplateSelections) TemplateContext {
	return TemplateContext{
		renderContext: renderContext,
		objectSelections: selections,
	}
}

type TemplateContext struct {
	renderContext *context.RenderContextImpl
	objectSelections TemplateSelections
}

func (t TemplateContext) Selections() TemplateSelections {
	return t.objectSelections
}

func (t TemplateContext) RenderContext() common.RenderContext {
	return t.renderContext
}

// TODO: make just functions?
func (t TemplateContext) templateGoLit(val any) string {
	panic("not implemented")
}

func (t TemplateContext) templateGoPtr(val any) string {
	panic("not implemented")
}

func (t TemplateContext) templateGoID(val any) string {
	panic("not implemented")
}

func (t TemplateContext) templateGoComment(text string) string {
	panic("not implemented")
}

func GetTemplateFunctions(t *TemplateContext) template.FuncMap {
	extraFuncs := template.FuncMap{
		"golit": func(val any) string { return t.templateGoLit(val) },
		"goptr": func(val any) string { return t.templateGoPtr(val) },
		"goid": func(val any) string { return t.templateGoID(val) },
		"gocomment": func(text string) string { return t.templateGoComment(text) },
		"qual": func(pkgExpr string) string { return t.renderContext.QualifiedName(pkgExpr) },
		"qualg": func(subPkg, name string) string { return t.renderContext.QualifiedGeneratedName(subPkg, name) },
		"qualr": func(subPkg, name string) string { return t.renderContext.QualifiedRuntimeName(subPkg, name) },
		"runTemplate": func(templateName string, ctx any) (string, error) {
			// TODO: template dir
			tpl := template.Must(template.ParseFS(templates.Templates))
			var bld strings.Builder
			if err := tpl.Execute(&bld, ctx); err != nil {
				return "", err
			}
			return bld.String(), nil
		},
		"contentTypeID": func(contentType string) string {
			// TODO: add other formats: protobuf, avro, etc.
			switch {
			case strings.HasSuffix(contentType, "json"):
				return "json"
			case strings.HasSuffix(contentType, "yaml"):
				return "yaml"
			}
			return ""
		},
	}

	return lo.Assign(templateFunctions, extraFuncs)
}

func selectObjects[T common.Renderer](selections []common.Renderer, kind common.ObjectKind) []T {
	return lo.FilterMap(selections, func(item common.Renderer, _ int) (T, bool) {
		if item.Kind() == kind {
			return item.(T), true
		}
		return lo.Empty[T](), false
	})
}
