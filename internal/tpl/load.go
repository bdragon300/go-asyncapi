package tpl

import (
	"github.com/bdragon300/go-asyncapi/templates"
	"text/template"
)

var loadedTemplates map[string]*template.Template

func LoadTemplate(name string) *template.Template {
	// TODO: user template dir
	if t, ok := loadedTemplates[name]; ok {
		return t
	}
	loadedTemplates[name] = template.Must(template.New(name).Funcs(GetTemplateFunctions()).ParseFS(templates.TemplateFS, name + ".tmpl"))
	return loadedTemplates[name]
}