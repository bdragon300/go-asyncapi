package tpl

import (
	"github.com/bdragon300/go-asyncapi/templates"
	"text/template"
)

const mainTemplateName = "main.tmpl"

var mainTemplate *template.Template

func LoadTemplate(name string) *template.Template {
	// TODO: user template dir
	if mainTemplate == nil {
		mainTemplate = template.Must(
			template.New(mainTemplateName).Funcs(GetTemplateFunctions()).ParseFS(templates.TemplateFS, "*/*.tmpl","*.tmpl"),
		)
	}
	t := mainTemplate.Lookup(name)
	if t == nil {
		panic("template not found: " + name)
	}
	return t
}