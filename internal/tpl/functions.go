package tpl

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/go-sprout/sprout"
	"github.com/samber/lo"
	"strings"
	"text/template"
	"unicode"
)

var sproutFunctions sprout.FunctionMap

func init() {
	handler := sprout.New()
	sproutFunctions = handler.Build()
}

func GetTemplateFunctions() template.FuncMap {
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
