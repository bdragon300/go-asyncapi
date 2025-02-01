package renderer

import (
	"bytes"
	"fmt"
	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"io/fs"
	"path"
)

func RenderImplementations(objects []common.ImplementationObject, mng *manager.TemplateRenderManager) error {
	res := make(map[string]*bytes.Buffer)
	logger := log.GetLogger(log.LoggerPrefixRendering)
	//TODO: logging

	for _, obj := range objects {
		logger.Debug("Render implementation", "name", obj.Manifest.Name, "protocol", obj.Manifest.Protocol)
		if obj.Config.Disable {
			logger.Debug("-> Skip disabled implementation")
			continue
		}

		ctx := tmpl.ImplTemplateContext{Package: obj.Config.Package, Manifest: obj.Manifest}
		directory, err := renderInlineTemplate(obj.Config.Directory, ctx, mng)
		if err != nil {
			return fmt.Errorf("render directory expression: %w", err)
		}
		directory = path.Clean(directory)

		pkgName, _ := lo.Coalesce(obj.Config.Package, utils.GetPackageName(directory))
		ctx = tmpl.ImplTemplateContext{Directory: directory, Package: pkgName, Manifest: obj.Manifest}

		tplFileGlob := path.Join(obj.Manifest.Dir, "*.tmpl")
		templateFiles := lo.Must(fs.Glob(implementations.ImplementationFS, tplFileGlob))
		for _, templateFile := range templateFiles {
			logger.Debug("-> Render file", "file", templateFile)
			tpl := tmpl.ParseTemplate(implementations.ImplementationFS, templateFile, mng)
			var buf bytes.Buffer
			if err := tpl.ExecuteTemplate(&buf, path.Base(templateFile), ctx); err != nil {
				return fmt.Errorf("execute template %q: %w", templateFile, err)
			}
			if _, ok := res[templateFile]; ok {
				return fmt.Errorf("duplicate file %q", path.Base(templateFile))
			}

			fileName := utils.NormalizePath(path.Join(directory, path.Base(templateFile)))
			res[fileName] = &buf
		}

		mng.AddImplementation(obj, directory)
	}

	return nil
}
