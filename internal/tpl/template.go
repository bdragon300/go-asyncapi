package tpl

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"reflect"
	"strings"
	"text/template"
	"unicode"
)

type renderContext interface {
	Imports() []common.ImportItem
	CurrentSelection() common.RenderSelectionConfig
}

func NewTemplateContext(renderContext renderContext, object common.Renderable, index int) TemplateContext {
	return TemplateContext{
		renderContext:    renderContext,
		object: object,
		index: index,
	}
}

// TemplateContext is passed as value to the root template on selections processing.
type TemplateContext struct {
	renderContext    renderContext
	object common.Renderable
	index int
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
	if  t.object.Kind() == common.ObjectKindOther {
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

func GetTemplateFunctions(ctx common.RenderContext) template.FuncMap {
	type golangTypeWrapperType interface {
		UnwrapGolangType() (common.GolangType, bool)
	}

	extraFuncs := template.FuncMap{
		"golit": func(val any) string { return templateGoLit(val) },
		"goptr": func(val common.GolangType) (string, error) { return templateGoPtr(val) },
		"unwrapgoptr": func(val common.GolangType) common.GolangType {
			if v, ok := any(val).(golangTypeWrapperType); ok {
				if wt, ok := v.UnwrapGolangType(); ok {
					return wt
				}
			}
			return nil
		},
		"goid": func(name string) string { return templateGoID(name) },
		"gocomment": func(text string) (string, error) { return templateGoComment(text) },
		"qual": func(parts ...string) string { return ctx.QualifiedName(parts...) },
		"qualgenpkg": func(obj common.GolangType) (string, error) {
			pkg, err := ctx.QualifiedGeneratedPackage(obj)
			if pkg == "" {
				return "", err
			}
			return pkg + ".", err
		},
		"qualrun": func(parts ...string) string { return ctx.QualifiedRuntimeName(parts...) }, // TODO: check if .Import and qual is enough
		"execTemplate": func(templateName string, ctx any) (string, error) {
			tmpl := LoadTemplate(templateName)
			var bld strings.Builder
			if err := tmpl.Execute(&bld, ctx); err != nil {
				return "", err
			}
			return bld.String(), nil
		},
		// TODO: function to run external command
	}

	return lo.Assign(sproutFunctions, extraFuncs)
}

func templateGoLit(val any) string {
	type usageDrawable interface {
		U() string
	}

	if v, ok := val.(usageDrawable); ok {
		return v.U()
	}
	return utils.ToGoLiteral(val)
}

func templateGoPtr(val common.GolangType) (string, error) {
	type golangPointerType interface {
		IsPointer() bool
	}

	if val == nil {
		return "", fmt.Errorf("cannot get a pointer to nil")
	}
	var isPtr bool
	if v, ok := val.(golangPointerType); ok && v.IsPointer() {
		isPtr = true
	}
	if isPtr {
		return "*" + val.U(), nil
	}
	return val.U(), nil
}

func templateGoID(val string) string {
	if val == "" {
		return ""
	}
	return utils.ToGolangName(val, unicode.IsUpper(rune(val[0])))
}

func templateGoComment(text string) (string, error) {
	if strings.HasPrefix(text, "//") || strings.HasPrefix(text, "/*") {
		// automatic formatting disabled.
		return text, nil
	}

	var b strings.Builder
	if strings.Contains(text, "\n") {
		if _, err := b.WriteString("/*\n"); err != nil {
			return "", err
		}
	} else {
		if _, err := b.WriteString("// "); err != nil {
			return "", err
		}
	}
	if _, err := b.WriteString(text); err != nil {
		return "", err
	}
	if strings.Contains(text, "\n") {
		if !strings.HasSuffix(text, "\n") {
			if _, err := b.WriteString("\n"); err != nil {
				return "", err
			}
		}
		if _, err := b.WriteString("*/"); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

//func selectObjects[T common.Renderable](selections []common.Renderable, kind common.ObjectKind) []T {
//	return lo.FilterMap(selections, func(item common.Renderable, _ int) (T, bool) {
//		if item.Kind() == kind {
//			return item.(T), true
//		}
//		return lo.Empty[T](), false
//	})
//}
