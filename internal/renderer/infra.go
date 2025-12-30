package renderer

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
)

func RenderInfra(objects []common.Artifact, activeProtocols []string, outputFileName string, serverConfig []common.InfraServerOpts, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	logger.Debug("Objects selected", "count", len(objects))
	ctx := tmpl.InfraTemplateContext{
		ServerConfig:    serverConfig,
		Objects:         objects,
		ActiveProtocols: activeProtocols,
	}

	logger.Trace("Loading root template")
	tpl, err := mng.TemplateLoader.LoadRootTemplate()
	if err != nil {
		return fmt.Errorf("load root template: %w", err)
	}
	mng.BeginFile(outputFileName, "main")
	logger.Debug("Render file", "name", outputFileName)
	if err = tpl.Execute(mng.Buffer, ctx); err != nil {
		return fmt.Errorf("root template: %w", err)
	}
	mng.Commit()

	logger.Info("Infra file rendered", "file", outputFileName)

	return nil
}
