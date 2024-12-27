package tmpl

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/templates"
	"github.com/samber/lo"
	"io/fs"
	"os"
	"path"
	"text/template"
)

const mainTemplateName = "main.tmpl"

var (
	builtinTemplate *template.Template
)

var ErrTemplateNotFound = errors.New("template not found")

func LoadTemplate(name string) (*template.Template, error) {
	log.GetLogger("").Trace("Loading template", "name", name)
	if t := builtinTemplate.Lookup(name); t != nil {
		return t, nil
	}
	return nil, ErrTemplateNotFound
}

func ParseTemplates(customDirectory string) {
	logger := log.GetLogger("")

	builtinTemplate = template.Must(
		template.New(mainTemplateName).Funcs(GetTemplateFunctions()).ParseFS(templates.TemplateFS, "*/*.tmpl","*.tmpl"),
	)
	if customDirectory == "" {
		return
	}

	logger.Debug("Use custom templates", "dir", customDirectory)
	dirFS := os.DirFS(customDirectory)
	fileNames := lo.Must(fs.Glob(dirFS, "*.tmpl"))
	fileNames = append(fileNames, lo.Must(fs.Glob(dirFS, "*/*.tmpl"))...)
	if len(fileNames) == 0 {
		logger.Warn("-> No *.tmpl files found in the directory", "dir", customDirectory)
		return
	}
	files := lo.Map(fileNames, func(fileName string, _ int) string {
		logger.Debug("-> Found custom template file", "name", fileName)
		return path.Join(customDirectory, fileName)
	})
	builtinTemplate = template.Must(builtinTemplate.ParseFiles(files...))
}
