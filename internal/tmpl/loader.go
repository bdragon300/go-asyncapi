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

var ErrTemplateNotFound = errors.New("template not found")

func NewTemplateLoader(rootTemplateName string, dir ...fs.FS) *TemplateLoader {
	return &TemplateLoader{dirs: dir, rootName: rootTemplateName}
}

type TemplateLoader struct {
	rootName string
	dirs     []fs.FS
	tpl      *template.Template
}

func (l *TemplateLoader) ParseRecursive(renderManager *manager.TemplateRenderManager) (err error) {
	l.tpl, err = parseTemplateFiles(l.dirs, l.rootName, renderManager)
	return
}

func (l *TemplateLoader) ParseDir(subDir string, renderManager *manager.TemplateRenderManager) ([]string, error) {
	logger := log.GetLogger("")

	l.tpl = template.New(l.rootName).Funcs(GetTemplateFunctions(renderManager))

	tplFileGlob := path.Join(subDir, "*.tmpl")
	var files []string
	for _, dir := range l.dirs {
		f, err := fs.Glob(dir, tplFileGlob)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", tplFileGlob, err)
		}
		files = append(files, f...)
		if _, err = l.tpl.ParseFS(dir, f...); err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		logger.Warn("-> No *.tmpl files found in the directory", "location", subDir)
		return nil, nil
	}

	fileNames := lo.Map(files, func(fileName string, _ int) string {
		return path.Base(fileName)
	})
	return fileNames, nil
}

func (l *TemplateLoader) LoadRootTemplate() (*template.Template, error) {
	return l.LoadTemplate(l.rootName)
}

func (l *TemplateLoader) LoadTemplate(name string) (*template.Template, error) {
	log.GetLogger("").Trace("Loading template", "name", name)
	if t := l.tpl.Lookup(name); t != nil {
		return t, nil
	}
	return nil, ErrTemplateNotFound
}

func parseTemplateFiles(dirs []fs.FS, name string, renderManager *manager.TemplateRenderManager) (*template.Template, error) {
	logger := log.GetLogger("")

	res := template.New(name).Funcs(GetTemplateFunctions(renderManager))

	for _, dir := range dirs {
		var fileNames []string
		// Recursively find all *.tmpl files in the location
		err := fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
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
			logger.Warn("-> No *.tmpl files found in the directory")
			continue
		}
		if _, err = res.ParseFS(dir, fileNames...); err != nil {
			return nil, err
		}
	}

	return res, nil
}
