package tmpl

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/templates"
	"os"
	"text/template"
)

const mainTemplateName = "main.tmpl"

var (
	builtinTemplate *template.Template
	customTemplate  *template.Template
)

var ErrTemplateNotFound = errors.New("template not found")

func LoadTemplate(name string) (*template.Template, error) {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	if builtinTemplate == nil {
		builtinTemplate = template.Must(
			template.New(mainTemplateName).Funcs(GetTemplateFunctions()).ParseFS(templates.TemplateFS, "*/*.tmpl","*.tmpl"),
		)
	}

	var t *template.Template
	if customTemplate != nil {
		logger.Debug("Try loading custom template", "name", name)
		if ct := customTemplate.Lookup(name); ct != nil {
			t = ct
		}
	}
	if t == nil {
		logger.Debug("Loading builtin template", "name", name)
		t = builtinTemplate.Lookup(name)
	}

	if t == nil {
		return nil, ErrTemplateNotFound
	}
	return t, nil
}

func SetCustomTemplateDirectory(dir string) {
	if dir == "" {
		return
	}
	dirFS := os.DirFS(dir)
	customTemplate = template.Must(
		template.New(mainTemplateName).Funcs(GetTemplateFunctions()).ParseFS(dirFS, "*/*.tmpl","*.tmpl"),
	)
}
