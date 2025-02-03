package renderer

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"text/template"
)

type renderImplTemplateLoader interface {
	ParseDir(subDir string, renderManager *manager.TemplateRenderManager) ([]string, error)
	LoadTemplate(name string) (*template.Template, error)
}

func RenderImplementations(objects []common.ImplementationObject, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	//TODO: logging

	tplLoader := mng.TemplateLoader.(renderImplTemplateLoader)

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

		pkgName, _ := lo.Coalesce(obj.Config.Package, utils.GetPackageName(directory))
		ctx = tmpl.ImplTemplateContext{Directory: directory, Package: pkgName, Manifest: obj.Manifest}

		fileNames, err := tplLoader.ParseDir(obj.Manifest.Dir, mng)
		if err != nil {
			return fmt.Errorf("parse directory %q: %w", obj.Manifest.Dir, err)
		}
		for _, fileName := range fileNames {
			normFileName := utils.ToGoFilePath(path.Join(directory, fileName))
			mng.BeginFile(normFileName, pkgName)

			logger.Debug("-> Render file", "name", normFileName)
			tpl, err := tplLoader.LoadTemplate(fileName)
			if err != nil {
				return fmt.Errorf("load template %q: %w", fileName, err)
			}
			if err := tpl.Execute(mng.Buffer, ctx); err != nil {
				return fmt.Errorf("execute template %q: %w", fileName, err)
			}

			mng.Commit()
		}

		mng.AddImplementation(obj, directory)
		mng.Commit()
	}


	return nil
}
