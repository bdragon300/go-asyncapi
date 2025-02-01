package renderer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"slices"
	"unicode"
)

type renderQueueItem struct {
	selection common.ConfigSelectionItem
	object         common.CompileObject
	err            error
}

func RenderObjects(objects []common.CompileObject, mng *manager.TemplateRenderManager) error {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	logger.Info("Run objects rendering")
	queue := buildRenderQueue(objects, mng.RenderOpts.Selections)
	var postponed []renderQueueItem

	logger.Info("Objects selected", "objects", len(queue))
	for len(queue) > 0 {
		for _, item := range queue {
			logger.Debug("Render", "object", item.object.String())

			logger.Trace("-> Render file name expression", "object", item.object.String(), "template", item.selection.Render.File)
			fileName, err := renderObjectInlineTemplate(item, item.selection.Render.File, mng)
			switch {
			case errors.Is(err, tmpl.ErrNotDefined):
				// Template can't be rendered right now due to unknown object definition, postpone it
				postponed = append(postponed, item)
				logger.Trace(
					"--> Postpone the file name rendering because some definitions it uses are not known yet",
					"object", item.object.String(),
				)
				continue
			case err != nil:
				return fmt.Errorf("render file name expression: %w", err)
			}
			fileName = utils.NormalizePath(fileName)
			logger.Trace("-> File", "name", fileName)

			logger.Debug("-> Render", "object", item.object.String(), "file", fileName, "template", item.selection.Render.Template)
			mng.BeginObject(item.object.Renderable, item.selection, fileName)
			err = renderObject(item, item.selection.Render.Template, mng)
			switch {
			case errors.Is(err, tmpl.ErrNotDefined):
				// Some objects needed by template code have not been defined and therefore, not in namespace yet.
				// Postpone this run to the end in hope that next runs will define these objects.
				item.err = fmt.Errorf("%s: %w", item.object.String(), err)
				logger.Trace(
					"--> Postpone the object because some the definitions of the object it uses are not known yet",
					"object", item.object.String(),
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
				"missed object definitions, please ensure they are defined by `godef` or `def` functions prior using: \n%w",
				errors.Join(lo.Map(postponed, func(item renderQueueItem, _ int) error { return item.err })...),
			)
		}
		if len(postponed) > 0 {
			logger.Trace("Process postponed objects", "objects", len(postponed))
		}
		queue, postponed = postponed, nil
	}

	return nil
}

func buildRenderQueue(allObjects []common.CompileObject, selections []common.ConfigSelectionItem) (res []renderQueueItem) {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	for _, selection := range selections {
		logger.Debug("Select objects", "selection", selection)
		selectedObjects := selector.SelectObjects(allObjects, selection)
		for _, obj := range selectedObjects {
			logger.Debug("-> Selected", "object", obj)
			res = append(res, renderQueueItem{selection: selection, object: obj})
		}
	}
	return
}

func renderObject(item renderQueueItem, templateName string, mng *manager.TemplateRenderManager) error {
	importManager := mng.ImportsManager.Clone()
	tplCtx := &tmpl.CodeTemplateContext{
		RenderOpts:       mng.RenderOpts,
		CurrentSelection: item.selection,
		PackageName:      mng.PackageName,
		Object:           item.object.Renderable,
		ImportsManager:   &importManager,
	}

	// Execute the main template first to accumulate imports and other data, that will be rendered in preamble
	tpl, err := tmpl.LoadTemplate(templateName)
	if err != nil {
		return fmt.Errorf("template %q: %w", templateName, err)
	}

	var res bytes.Buffer
	if err := tpl.Execute(&res, tplCtx); err != nil {
		return err
	}

	// Update the file state if rendering was successful
	// If item is marked reused from other place, do not update the file state and content, just update the namespace
	if item.selection.ReusePackagePath == "" {
		mng.Buffer.Write(res.Bytes())
		mng.ImportsManager = importManager
	}

	return nil
}

func FinishFiles(mng *manager.TemplateRenderManager) (map[string]*bytes.Buffer, error) {
	states := mng.AllStates()

	var res = make(map[string]*bytes.Buffer, len(states))
	logger := log.GetLogger(log.LoggerPrefixRendering)
	logger.Debug("Render files", "files", len(states))

	tpl, err := tmpl.LoadTemplate(mng.RenderOpts.PreambleTemplate)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", mng.RenderOpts.PreambleTemplate, err)
	}

	keys := lo.Keys(states)
	slices.Sort(keys)
	for _, fileName := range keys {
		state := states[fileName]
		logger.Debug("Render file", "file", fileName, "package", state.PackageName, "imports", state.Imports.String())
		if !bytes.ContainsFunc(state.Buffer.Bytes(), unicode.IsLetter) {
			logger.Debug("-> Skip empty file", "file", fileName)
			continue
		}
		b, err := renderPreambleTemplate(tpl, mng)
		if err != nil {
			return nil, err
		}
		res[fileName] = b
	}

	return res, nil
}
