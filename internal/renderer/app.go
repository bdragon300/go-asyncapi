package renderer

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/samber/lo"
)

const (
	DefaultClientAppFileName = "client.go"
)

func RenderClientApp(queue []RenderQueueItem, activeProtocols []string, goModTemplate string, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	objects := lo.Map(queue, func(item RenderQueueItem, _ int) common.Renderable {
		return item.Object.Renderable
	})
	logger.Debug("Objects selected", "objects", len(objects))
	ctx := tmpl.AppTemplateContext{
		RenderOpts:       mng.RenderOpts,
		Objects:         objects,
		ActiveProtocols: activeProtocols,
	}

	logger.Trace("Loading root template")
	tpl, err := mng.TemplateLoader.LoadRootTemplate()
	if err != nil {
		return fmt.Errorf("load root template: %w", err)
	}
	mng.BeginFile(DefaultClientAppFileName, "main")
	logger.Debug("Render file", "name", DefaultClientAppFileName)
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

	logger.Info("Client app rendered", "file", DefaultClientAppFileName)

	return nil
}