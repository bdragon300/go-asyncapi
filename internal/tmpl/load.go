package tmpl

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/templates"
	"text/template"
)

const mainTemplateName = "main.tmpl"

var mainTemplate *template.Template
var ErrTemplateNotFound = errors.New("template not found")

func LoadTemplate(name string) (*template.Template, error) {
	// TODO: user template dir
	if mainTemplate == nil {
		mainTemplate = template.Must(
			template.New(mainTemplateName).Funcs(GetTemplateFunctions()).ParseFS(templates.TemplateFS, "*/*.tmpl","*.tmpl"),
		)
	}
	t := mainTemplate.Lookup(name)
	if t == nil {
		return nil, ErrTemplateNotFound
	}
	return t, nil
}