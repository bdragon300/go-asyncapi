package tmpl

import (
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/samber/lo"
	"io/fs"
	"path"
	"strings"
	"text/template"
)

const templateExtension = ".tmpl"

var ErrTemplateNotFound = errors.New("template not found")

func NewTemplateLoader(rootTemplateName string, locations ...fs.FS) *TemplateLoader {
	return &TemplateLoader{locations: locations, rootName: rootTemplateName}
}

// TemplateLoader parses and manages the Go templates from one or more given locations. All found templates are kept in
// a root [text/template.Template] object, that is available by LoadRootTemplate method. Other templates are returned
// by LoadTemplate.
type TemplateLoader struct {
	locations []fs.FS
	rootName  string
	tpl       *template.Template
}

// ParseRecursive parses all templates from the locations recursively, initializing them with template functions
// using the passed renderManager.
//
// After parsing, templates are available by LoadRootTemplate and LoadTemplate methods.
func (l *TemplateLoader) ParseRecursive(renderManager *manager.TemplateRenderManager) (err error) {
	l.tpl, err = parseTemplateFiles(l.locations, l.rootName, renderManager)
	return
}

// ParseDir parses all templates in the directory *not-recursively* in all locations, initializing them with template
// functions using the passed renderManager.
//
// Returns the list of parsed template file names without paths.
func (l *TemplateLoader) ParseDir(dir string, renderManager *manager.TemplateRenderManager) ([]string, error) {
	logger := log.GetLogger("")

	l.tpl = template.New(l.rootName).Funcs(GetTemplateFunctions(renderManager))

	fileGlob := path.Join(dir, "*" + templateExtension)
	var files []string
	for _, loc := range l.locations {
		f, err := fs.Glob(loc, fileGlob)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", fileGlob, err)
		}
		files = append(files, f...)
		if _, err = l.tpl.ParseFS(loc, f...); err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		logger.Warn("-> No template files found", "dir", dir, "extension", templateExtension)
		return nil, nil
	}

	fileNames := lo.Map(files, func(fileName string, _ int) string {
		return path.Base(fileName)
	})
	return fileNames, nil
}

// LoadRootTemplate returns the template with the root name, that was set on creation.
func (l *TemplateLoader) LoadRootTemplate() (*template.Template, error) {
	return l.LoadTemplate(l.rootName)
}

// LoadTemplate returns the template by the given name.
func (l *TemplateLoader) LoadTemplate(name string) (*template.Template, error) {
	log.GetLogger("").Trace("Loading template", "name", name)
	if t := l.tpl.Lookup(name); t != nil {
		return t, nil
	}
	return nil, ErrTemplateNotFound
}

func parseTemplateFiles(locations []fs.FS, name string, renderManager *manager.TemplateRenderManager) (*template.Template, error) {
	logger := log.GetLogger("")

	res := template.New(name).Funcs(GetTemplateFunctions(renderManager))

	var fileCount int
	for _, loc := range locations {
		var fileNames []string
		// Recursively find all template files in the location
		err := fs.WalkDir(loc, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, templateExtension) {
				return nil
			}
			fileNames = append(fileNames, path)
			logger.Debug("-> Found template file", "name", path)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk directory: %w", err)
		}
		if len(fileNames) == 0 {
			logger.Debug("-> No template files found in the directory", "extension", templateExtension)
			continue
		}
		if _, err = res.ParseFS(loc, fileNames...); err != nil {
			return nil, err
		}
		fileCount += len(fileNames)
	}

	if fileCount == 0 {
		logger.Warn("-> No template files found", "locations", locations, "extension", templateExtension)
		return res, nil
	}
	logger.Debug("-> Parsed templates", "files", fileCount, "templates", len(res.Templates()))

	return res, nil
}
