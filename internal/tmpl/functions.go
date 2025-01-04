package tmpl

import (
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"strings"
	"text/template"
	"unicode"
)

func GetTemplateFunctions() template.FuncMap {
	type golangTypeExtractor interface {
		InnerGolangType() common.GolangType
	}

	extraFuncs := template.FuncMap{
		// Functions that return go code as string
		"golit": func(val any) (string, error) { return templateGoLit(val) },
		"goptr": func(val common.GolangType) (string, error) { return templateGoPtr(val) },
		"goid": func(val any) string { return templateGoID(val) },
		"gocomment": func(text string) (string, error) { return templateGoComment(text) },
		"goqual": func(parts ...string) string { return common.GetContext().QualifiedName(parts...) },
		"goqualrun": func(parts ...string) string { return common.GetContext().QualifiedRuntimeName(parts...) },
		"godef": func(r common.GolangType) (string, error) {
			tplName := path.Join(r.GoTemplate(), "definition")
			common.GetContext().DefineTypeInNamespace(r, common.GetContext().CurrentSelection(), true)
			if v, ok := r.(golangTypeWrapper); ok {
				r = v.UnwrapGolangType()
			}
			res, err := templateExecTemplate(tplName, r)
			if err != nil {
				return "", fmt.Errorf("%s: %w", r, err)
			}
			return res, nil
		},
		"gopkg": func(obj common.GolangType) (string, error) {
			pkg, err := common.GetContext().QualifiedGeneratedPackage(obj)
			if err != nil {
				return "", fmt.Errorf("%s: %w", obj, err)
			}
			if pkg == "" {
				return "", err
			}
			return pkg + ".", err
		},
		"gousage": func(r common.GolangType) (string, error) { return TemplateGoUsage(r) },

		// Type helpers
		"deref": func(r common.Renderable) common.Renderable {
			return common.DerefRenderable(r)
		},
		"innertype": func(val common.GolangType) common.GolangType {
			if v, ok := any(val).(golangTypeExtractor); ok {
				return v.InnerGolangType()
			}
			return nil
		},
		"visible": func(r common.Renderable) common.Renderable {
			return lo.Ternary(!lo.IsNil(r) && r.Visible(), r, nil)
		},

		// Templates calling
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

		// Working with render namespace
		"def": func(objects ...any) string {
			for _, o := range objects {
				switch v := o.(type) {
				case common.GolangType:
					if !lo.IsNil(o) {
						common.GetContext().DefineTypeInNamespace(v, common.GetContext().CurrentSelection(), false)
					}
				case string:
					if o != "" {
						common.GetContext().DefineNameInNamespace(v)
					}
				}
			}
			return ""
		},
		"defined": func(r any) bool {
			return templateGoDefined(r)
		},
		"ndefined": func(r any) bool {
			return !templateGoDefined(r)
		},

		// Other
		"debug": func(args ...any) string {
			logger := log.GetLogger(log.LoggerPrefixRendering)
			for _, arg := range args {
				logger.Debugf("debug: [%[1]p] %[1]v", arg)
			}
			return ""
		},
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

type golangTypeWrapper interface {
	UnwrapGolangType() common.GolangType
}

func TemplateGoUsage(r common.GolangType) (string, error) {
	tplName := path.Join(r.GoTemplate(), "usage")
	if v, ok := r.(golangTypeWrapper); ok {
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
	return lo.Ternary(val.Addressable(), "*"+s, s), nil
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
