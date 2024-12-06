package tpl

import (
	"github.com/bdragon300/go-asyncapi/templates"
	"github.com/go-sprout/sprout"
	"text/template"
)

var inlineTemplates = template.Must(template.ParseFS(templates.Templates))

var sproutFunctions sprout.FunctionMap

func init() {
	handler := sprout.New()
	sproutFunctions = handler.Build()
}

func LoadTemplate(name string) *template.Template {
	// TODO: template dir
	// TODO: cache
	return inlineTemplates.Lookup(name)
}