package tmpl

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"strings"
	"text/template"
	"unicode"
)

func GetTemplateFunctions() template.FuncMap {
	type golangTypeWrapperType interface {
		UnwrapGolangType() (common.GolangType, bool)
	}

	extraFuncs := template.FuncMap{
		"golit": func(val any) (string, error) { return templateGoLit(val) },
		"goptr": func(val common.GolangType) (string, error) { return templateGoPtr(val) },
		"unwrapgoptr": func(val common.GolangType) common.GolangType {
			if v, ok := any(val).(golangTypeWrapperType); ok {
				if wt, ok := v.UnwrapGolangType(); ok {
					return wt
				}
			}
			return nil
		},
		"goid": func(val any) string { return templateGoID(val) },
		"gocomment": func(text string) (string, error) { return templateGoComment(text) },
		"qual": func(parts ...string) string { return common.GetContext().QualifiedName(parts...) },
		"qualgenpkg": func(obj common.GolangType) (string, error) {
			pkg, err := common.GetContext().QualifiedGeneratedPackage(obj)
			if pkg == "" {
				return "", err
			}
			return pkg + ".", err
		},
		"qualrun": func(parts ...string) string { return common.GetContext().QualifiedRuntimeName(parts...) }, // TODO: check if .Import and qual is enough
		"execTemplate": func(templateName string, ctx any) (string, error) {
			return templateExecTemplate(templateName, ctx)
		},
		"localobj": func(obj ...common.GolangType) string {
			for _, o := range obj {
				o.SetDefinitionInfo(common.GetContext().CurrentDefinitionInfo())
			}
			return ""
		},
		"godef": func(r common.GolangType) (string, error) {
			tplName := path.Join(r.GoTemplate(), "definition")
			r.SetDefinitionInfo(common.GetContext().CurrentDefinitionInfo())
			return templateExecTemplate(tplName, unwrapGolangPromise(r))
		},
		"gousage": func(r common.GolangType) (string, error) { return TemplateGoUsage(r) },
		// TODO: function to run external command
	}

	return lo.Assign(sproutFunctions, extraFuncs)
}

func TemplateGoUsage(r common.GolangType) (string, error) {
	tplName := path.Join(r.GoTemplate(), "usage")
	return templateExecTemplate(tplName, unwrapGolangPromise(r))
}

func templateExecTemplate(templateName string, ctx any) (string, error) {
	var bld strings.Builder

	tpl := LoadTemplate(templateName)
	if err := tpl.Execute(&bld, ctx); err != nil {
		return "", err
	}
	return bld.String(), nil
}

func templateGoLit(val any) (string, error) {
	if v, ok := val.(common.GolangType); ok {
		return TemplateGoUsage(v)
	}
	return utils.ToGoLiteral(val), nil
}

func templateGoPtr(val common.GolangType) (string, error) {
	if val == nil {
		return "", fmt.Errorf("cannot get a pointer to nil")
	}
	s, err := TemplateGoUsage(val)
	if err != nil {
		return "", err
	}
	return lo.Ternary(val.IsPointer(), s, "*"+s), nil
}



func templateGoID(val any) string {
	var res string

	switch v := val.(type) {
	case common.Renderable:
		r := unwrapRenderablePromise(common.GetContext().CurrentObject().Renderable)
		// If an object passed as an argument is the same as the current object, return name taken from context
		// This name can be either original name or alias taken from promise
		if r == v {
			return common.GetContext().CurrentObject().GetOriginalName()
		}
		return v.GetOriginalName()  // This is some another object, just return its original name
	case string:
		res = v
	default:
		panic(fmt.Sprintf("goid doesn't support the type %[1]T: %[1]v", val))
	}

	if res == "" {
		return ""
	}
	return utils.ToGolangName(res, unicode.IsUpper(rune(res[0])))
}

type golangTypePromise interface {
	GolangTypeT() common.GolangType
}

func unwrapGolangPromise(val common.GolangType) common.GolangType {
	o, ok := val.(golangTypePromise)
	for ok {
		val = o.GolangTypeT()
		o, ok = val.(golangTypePromise)
	}
	return val
}

type renderablePromise interface {
	RenderableT() common.Renderable
}

func unwrapRenderablePromise(val common.Renderable) common.Renderable {
	o, ok := val.(renderablePromise)
	for ok {
		val = o.RenderableT()
		o, ok = val.(renderablePromise)
	}
	return val
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
