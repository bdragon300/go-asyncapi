package tmpl

import (
	"errors"
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

	// TODO: review the functions names
	extraFuncs := template.FuncMap{
		"golit": func(val any) (string, error) { return templateGoLit(val) },
		"goptr": func(val common.GolangType) (string, error) { return templateGoPtr(val) },
		"unwrapgoptr": func(val common.GolangType) common.GolangType {  // TODO: remove?
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
			if err != nil {
				return "", fmt.Errorf("%s: %w", obj, err)
			}
			if pkg == "" {
				return "", err
			}
			return pkg + ".", err
		},
		"qualrun": func(parts ...string) string { return common.GetContext().QualifiedRuntimeName(parts...) },
		"tmpl": func(templateName string, ctx any) (string, error) {
			return templateExecTemplate(templateName, ctx)
		},
		"trytmpl": func(templateName string, ctx any) (string, error) {
			res, err := templateExecTemplate(templateName, ctx)
			switch {
			case errors.Is(err, ErrTemplateNotFound):
				return "", nil
			case err != nil:
				return "", err
			}

			return res, nil
		},
		"localobj": func(obj ...common.GolangType) string {
			for _, o := range obj {
				if !lo.IsNil(o) {
					common.GetContext().DefineTypeInNamespace(o, common.GetContext().CurrentSelection(), false)
				}
			}
			return ""
		},
		"godef": func(r common.GolangType) (string, error) {
			tplName := path.Join(r.GoTemplate(), "definition")
			common.GetContext().DefineTypeInNamespace(r, common.GetContext().CurrentSelection(), true)
			if v, ok := r.(golangTypeUnwrapper); ok {
				r = v.UnwrapGolangType()
			}
			res, err := templateExecTemplate(tplName, r)
			if err != nil {
				return "", fmt.Errorf("%s: %w", r, err)
			}
			return res, nil
		},
		"def": func(name string) string {
			common.GetContext().DefineNameInNamespace(name)
			return ""
		},
		"gousage": func(r common.GolangType) (string, error) { return TemplateGoUsage(r) },
		"godefined": func(r any) bool {
			return templateGoDefined(r)
		},
		"gondefined": func(r any) bool {
			return !templateGoDefined(r)
		},
		"deref": func(r common.Renderable) common.Renderable {
			if v, ok := r.(renderableUnwrapper); ok {
				return v.UnwrapRenderable()
			}
			return r
		},
		"debug": func(args ...any) string { // TODO: remove or replace with log
			fmt.Printf("DEBUG: %[1]v %[1]p\n", args...)
			return ""
		},
		// TODO: function to run external command
	}

	return lo.Assign(sproutFunctions, extraFuncs)
}

func templateGoDefined(r any) bool {
	if lo.IsNil(r) {
		return false
	}
	switch v := r.(type) {
	case common.GolangType:
		return common.GetContext().TypeDefinedInNamespace(v)
	case string:
		return common.GetContext().NameDefinedInNamespace(v)
	}

	panic(fmt.Sprintf("unsupported type %[1]T: %[1]v", r))
}

type renderableUnwrapper interface {
	UnwrapRenderable() common.Renderable
}

type golangTypeUnwrapper interface {
	UnwrapGolangType() common.GolangType
}

func TemplateGoUsage(r common.GolangType) (string, error) {
	tplName := path.Join(r.GoTemplate(), "usage")
	if v, ok := r.(golangTypeUnwrapper); ok {
		r = v.UnwrapGolangType()
	}
	res, err := templateExecTemplate(tplName, r)
	if err != nil {
		return "", fmt.Errorf("%s: %w", r, err)
	}
	return res, nil
}

func templateExecTemplate(templateName string, ctx any) (string, error) {
	var bld strings.Builder

	tpl, err := LoadTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("template %q: %w", templateName, err)
	}
	if err := tpl.Execute(&bld, ctx); err != nil {
		return "", fmt.Errorf("template %q: %w", templateName, err)
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
		return "", fmt.Errorf("%s: %w", val, err)
	}
	return lo.Ternary(val.IsPointer(), s, "*"+s), nil
}

func templateGoID(val any) string {
	var res string

	switch v := val.(type) {
	case common.Renderable:
		res = common.GetContext().GetObjectName(v)
	case string:
		res = v
	default:
		panic(fmt.Sprintf("type is not supported %[1]T: %[1]v", val))
	}

	if res == "" {
		return ""
	}
	return utils.ToGolangName(res, unicode.IsUpper(rune(res[0])))
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
