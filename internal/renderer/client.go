package renderer

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/samber/lo"
)

func RenderClientApp(queue []RenderQueueItem, activeProtocols []string, goModTemplate, sourceCodeFile string, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	objects := lo.Map(queue, func(item RenderQueueItem, _ int) common.Renderable {
		return item.Object.Renderable
	})
	logger.Debug("Objects selected", "count", len(objects))
	ctx := tmpl.ClientAppTemplateContext{
		RenderOpts:       mng.RenderOpts,
		Objects:         objects,
		ActiveProtocols: activeProtocols,
	}

	logger.Trace("Loading root template")
	tpl, err := mng.TemplateLoader.LoadRootTemplate()
	if err != nil {
		return fmt.Errorf("load root template: %w", err)
	}
	mng.BeginFile(sourceCodeFile, "main")
	logger.Debug("Render file", "name", sourceCodeFile)
	if err = tpl.Execute(mng.Buffer, ctx); err != nil {
		return fmt.Errorf("root template: %w", err)
	}
	mng.Commit()

	logger.Trace("Loading template", goModTemplate)
	tpl, err = mng.TemplateLoader.LoadTemplate(goModTemplate)
	if err != nil {
		return fmt.Errorf("load template %s: %w", goModTemplate, err)
	}
	mng.BeginFile("go.mod", "main")
	logger.Debug("Render file", "name", "go.mod")
	if err = tpl.Execute(mng.Buffer, ctx); err != nil {
		return fmt.Errorf("template %s: %w", goModTemplate, err)
	}
	mng.Commit()

	logger.Info("Client app rendered", "file", sourceCodeFile)

	return nil
}