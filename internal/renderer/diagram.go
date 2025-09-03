package renderer

import (
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

func RenderDiagramOneFile(objects []common.Artifact, fileName string, config common.ConfigDiagram, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	logger.Trace("Loading root template")
	tpl, err := mng.TemplateLoader.LoadRootTemplate()
	if err != nil {
		return fmt.Errorf("load root template: %w", err)
	}
	ctx := tmpl.DiagramTemplateContext{
		Objects: objects,
		Config:  config,
	}

	logger.Debug("Render diagram file", "name", fileName)
	if err := renderDiagram(tpl, ctx, fileName, mng); err != nil {
		return err
	}
	logger.Info("Diagram file rendered", "file", fileName)
	return nil
}

func RenderDiagramMultipleFiles(documents map[string][]common.Artifact, fileExtension string, config common.ConfigDiagram, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	logger.Trace("Loading root template")
	tpl, err := mng.TemplateLoader.LoadRootTemplate()
	if err != nil {
		return fmt.Errorf("load root template: %w", err)
	}

	for location, group := range utils.OrderedKeysIter(documents) {
		ctx := tmpl.DiagramTemplateContext{
			Objects: group,
			Config:  config,
		}
		fileName := strings.TrimSuffix(path.Base(location), path.Ext(location))

		// Ensure unique file names
		newFileName := fileName
		for i := 1; ; i++ {
			if _, ok := mng.CommittedStates()[newFileName]; !ok {
				break
			}
			newFileName = fmt.Sprintf("%s_%d", fileName, i)
		}
		newFileName += fileExtension

		logger.Debug("Rendering file", "name", newFileName)
		if err := renderDiagram(tpl, ctx, newFileName, mng); err != nil {
			return err
		}
	}
	logger.Info("Diagram files rendered", "count", len(mng.CommittedStates()))
	return nil
}

func renderDiagram(tpl *template.Template, ctx tmpl.DiagramTemplateContext, outputFileName string, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	mng.BeginFile(outputFileName, "main")
	logger.Debug("Render file", "name", outputFileName)
	if err := tpl.Execute(mng.Buffer, ctx); err != nil {
		return fmt.Errorf("root template: %w", err)
	}
	mng.Commit()
	return nil
}
