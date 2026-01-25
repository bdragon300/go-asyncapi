package renderer

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
)

func RenderUI(u *jsonpointer.JSONPointer, outFileName string, docContents map[string]any, resources []common.UIHTMLResourceOpts, opts common.UIRenderOpts, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	logger.Trace("Loading root template")
	tpl, err := mng.TemplateLoader.LoadRootTemplate()
	if err != nil {
		return fmt.Errorf("load root template: %w", err)
	}

	renderCtx := tmpl.UITemplateContext{
		DocumentContents: docContents,
		DocumentURL:      *u, // supposed to be non-nil here
		Config:           opts,
		Resources:        resources,
	}

	mng.BeginFile(outFileName, "main")
	logger.Debug("Render file", "name", outFileName)
	if err := tpl.Execute(mng.Buffer, renderCtx); err != nil {
		return fmt.Errorf("root template: %w", err)
	}
	mng.Commit()
	logger.Info("File rendered", "file", outFileName)

	return nil
}
