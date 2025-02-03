package renderer

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/samber/lo"
)

const defaultClientAppFileName = "client.go"

func RenderClientApp(queue []RenderQueueItem, activeProtocols []string, mng *manager.TemplateRenderManager) error {
	// TODO: logging

	objects := lo.Map(queue, func(item RenderQueueItem, _ int) common.Renderable {
		return item.Object.Renderable
	})
	ctx := tmpl.AppTemplateContext{
		Objects:         objects,
		ActiveProtocols: activeProtocols,
	}
	tpl, err := mng.TemplateLoader.LoadRootTemplate()
	if err != nil {
		return fmt.Errorf("root template: %w", err)
	}

	mng.BeginFile(defaultClientAppFileName, "main")
	if err := tpl.Execute(mng.Buffer, ctx); err != nil {
		return err
	}
	mng.Commit()

	return nil
}