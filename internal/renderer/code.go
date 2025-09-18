package renderer

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

type RenderQueueItem struct {
	LayoutItem common.ConfigLayoutItem
	Object     common.Artifact
	Err        error
}

func RenderArtifacts(queue []RenderQueueItem, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	var postponed []RenderQueueItem

	objectsCount := len(queue)
	logger.Debug("Objects selected", "count", objectsCount)
	for len(queue) > 0 {
		for _, item := range queue {
			logger.Debug("Render", "object", item.Object.String())
			mng.SetCodeObject(item.Object, item.LayoutItem)

			logger.Trace("-> Render file name expression", "object", item.Object.String(), "template", item.LayoutItem.Render.File)
			fileName, err := renderObjectInlineTemplate(item, item.LayoutItem.Render.File, mng)
			switch {
			case errors.Is(err, tmpl.ErrNotPinned):
				// Template can't be rendered right now due to unknown object definition, postpone it
				postponed = append(postponed, item)
				logger.Trace(
					"--> Postpone the file name rendering because some definitions it uses are not known yet",
					"object", item.Object.String(),
				)
				continue
			case err != nil:
				return fmt.Errorf("render file name expression: %w", err)
			}
			fileName = utils.ToGoFilePath(fileName)
			logger.Trace("-> File", "name", fileName)

			logger.Debug("-> Render", "object", item.Object.String(), "file", fileName, "template", item.LayoutItem.Render.Template)
			mng.BeginFile(fileName, item.LayoutItem.Render.Package)
			err = renderObject(item, item.LayoutItem.Render.Template, mng)
			switch {
			case errors.Is(err, tmpl.ErrNotPinned):
				// Some objects needed by template code have not been defined and therefore, not in namespace yet.
				// Postpone this run to the end in hope that next runs will define these objects.
				item.Err = fmt.Errorf("%s: %w", item.Object.String(), err)
				logger.Trace(
					"--> Postpone the object because some the definitions of the object it uses are not known yet",
					"object", item.Object.String(),
				)
				postponed = append(postponed, item)
				continue
			case err != nil:
				return fmt.Errorf("render object: %w", err)
			}

			mng.Commit()
			logger.Trace("--> Updated file state", "imports", mng.ImportsManager.String(), "namespace", mng.NamespaceManager.String())
		}
		if len(postponed) == len(queue) {
			return fmt.Errorf(
				"objects are not pinned to any file: \n%w",
				errors.Join(lo.Map(postponed, func(item RenderQueueItem, _ int) error { return item.Err })...),
			)
		}
		if len(postponed) > 0 {
			logger.Trace("Process postponed objects", "objects", len(postponed))
		}
		queue, postponed = postponed, nil
	}
	logger.Info("Objects rendered", "count", objectsCount)

	return nil
}

func renderObject(item RenderQueueItem, templateName string, mng *manager.TemplateRenderManager) error {
	tplCtx := &tmpl.CodeTemplateContext{
		RenderOpts:        mng.RenderOpts,
		CurrentLayoutItem: item.LayoutItem,
		PackageName:       mng.PackageName,
		Object:            item.Object,
		ImportsManager:    mng.ImportsManager,
	}

	var tpl *template.Template
	var err error
	if templateName == "" {
		tpl, err = mng.TemplateLoader.LoadRootTemplate()
	} else {
		tpl, err = mng.TemplateLoader.LoadTemplate(templateName)
	}
	if err != nil {
		return fmt.Errorf("template %q: %w", lo.Ternary(templateName != "", templateName, "<root>"), err)
	}

	var res bytes.Buffer
	importsSnapshot := mng.ImportsManager.Clone()
	if err := tpl.Execute(&res, tplCtx); err != nil {
		return err
	}

	// If item is marked reused from other place, do not render the object and new imports, just update the namespace
	if item.LayoutItem.ReusePackagePath != "" {
		mng.ImportsManager = importsSnapshot
	} else {
		mng.Buffer.Write(res.Bytes())
	}
	return nil
}
