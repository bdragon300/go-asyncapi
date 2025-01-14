package tmpl

import (
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
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
		"goid": func(val any) string { return templateGoID(val, true) },
		"goidorig": func(val any) string { return templateGoID(val, false) },
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
		"gopkg": func(obj any) (pkg string, err error) {
			switch v := obj.(type) {
			case common.GolangType:
				pkg, err = common.GetContext().QualifiedTypeGeneratedPackage(v)
			case *common.ImplementationObject:
				if lo.IsNil(v) {
					return "", errors.New("argument is nil")
				}
				pkg, err = common.GetContext().QualifiedImplementationGeneratedPackage(*v)
			default:
				return "", fmt.Errorf("type is not supported %[1]T: %[1]v", obj)
			}

			if err != nil {
				return "", fmt.Errorf("%s: %w", obj, err)
			}
			return lo.Ternary(pkg != "", pkg + ".", ""), nil
		},
		"gousage": func(r common.GolangType) (string, error) { return TemplateGoUsage(r) },

		// Type helpers
		"deref": func(r common.Renderable) common.Renderable {
			if r == nil {
				return nil
			}
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
		"ptr": func(val common.GolangType) (common.GolangType, error) {
			if lo.IsNil(val) {
				return nil, fmt.Errorf("cannot get a pointer to nil")
			}
			return &lang.GoPointer{Type: val}, nil
		},
		"impl": func(protocol string) *common.ImplementationObject {
			impl, found := common.GetContext().FindImplementationInNamespace(protocol)
			if !found {
				return nil
			}
			return &impl
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

func templateGoID(val any, forceCapitalize bool) string {
	var res string

	switch v := val.(type) {
	case common.Renderable:
		// Prefer the name of the topObject over the name of the val if the topObject is a Ref and points to val.
		// Otherwise, use the name of the val.
		//
		// For example, the topObject is a lang.Ref defined in `servers.myServer`. val contains the render.Server
		// defined in `components.servers.reusableServer` that this Ref is points to. Then we'll use the "myServer"
		// as the server name in generated code: functions, structs, etc.
		topObject := common.GetContext().GetObject()
		res = lo.Ternary(common.CheckSameRenderables(topObject.Renderable, v), topObject.Name(), v.Name())
	case string:
		res = v
	default:
		panic(fmt.Sprintf("type is not supported %[1]T: %[1]v", val))
	}

	if res == "" {
		return ""
	}
	exported := true
	if !forceCapitalize {
		exported = unicode.IsUpper(rune(res[0]))
	}
	return utils.ToGolangName(res, exported)
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
