package tpl

import (
	"github.com/bdragon300/go-asyncapi/templates"
	"text/template"
)

var inlineTemplates = template.Must(template.ParseFS(templates.Templates))

func LoadTemplate(name string) *template.Template {
	// TODO: template dir
	// TODO: cache
	return inlineTemplates.Lookup(name)
}