package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/tpl"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/go-sprout/sprout"
	"github.com/samber/lo"
	"strings"
	"text/template"
	"unicode"
)

var templateFunctions sprout.FunctionMap

func init() {
	handler := sprout.New()
	templateFunctions = handler.Build()
}

type TemplateSelections struct {
	Objects []common.Renderable
}

func (r TemplateSelections) SelectLangs() []common.Renderable {
	return lo.Filter(r.Objects, func(item common.Renderable, _ int) bool {
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
	renderContext    *context.RenderContextImpl
	objectSelections TemplateSelections
}

func (t TemplateContext) Selections() TemplateSelections {
	return t.objectSelections
}


func GetTemplateFunctions(ctx common.RenderContext) template.FuncMap {
	extraFuncs := template.FuncMap{
		"golit": func(val any) string { return templateGoLit(val) },
		"goptr": func(val common.GolangType) (*lang.GoPointer, error) { return templateGoPtr(val) },
		"unwrapgoptr": func(val common.GolangType) common.GolangType {
			if v, ok := any(val).(lang.GolangTypeWrapperType); ok {
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
		"runTemplate": func(templateName string, ctx any) (string, error) {
			tmpl := tpl.LoadTemplate(templateName)
			var bld strings.Builder
			if err := tmpl.Execute(&bld, ctx); err != nil {
				return "", err
			}
			return bld.String(), nil
		},
	}

	return lo.Assign(templateFunctions, extraFuncs)
}


func templateGoLit(val any) string {
	type usageDrawable interface {
		U() string
	}

	var res string
	switch val.(type) {
	case usageDrawable:
		return val.(usageDrawable).U()
	case bool, string, int, complex128:
		// default constant types can be left bare
		return fmt.Sprintf("%#v", val)
	case float64:
		res = fmt.Sprintf("%#v", val)
		if !strings.Contains(res, ".") && !strings.Contains(res, "e") {
			// If the formatted value is not in scientific notation, and does not have a dot, then
			// we add ".0". Otherwise, it will be interpreted as an int.
			// See:
			// https://github.com/dave/jennifer/issues/39
			// https://github.com/golang/go/issues/26363
			res += ".0"
		}
		return res
	case float32, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		// other built-in types need specific type info
		return fmt.Sprintf("%T(%#v)", val, val)
	case complex64:
		// fmt package already renders parenthesis for complex64
		return fmt.Sprintf("%T%#v", val, val)
	}

	panic(fmt.Sprintf("unsupported type for literal: %T", val))
}

func templateGoPtr(val common.GolangType) (*lang.GoPointer, error) {
	if val == nil {
		return nil, fmt.Errorf("cannot get a pointer to nil")
	}
	return &lang.GoPointer{Type: val}, nil
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

func selectObjects[T common.Renderable](selections []common.Renderable, kind common.ObjectKind) []T {
	return lo.FilterMap(selections, func(item common.Renderable, _ int) (T, bool) {
		if item.Kind() == kind {
			return item.(T), true
		}
		return lo.Empty[T](), false
	})
}
