package renderer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"slices"
	"unicode"
)

type fileRenderState struct {
	imports     context.ImportsList
	packageName string
	fileName   string
	buf        bytes.Buffer
}

type renderQueueItem struct {
	selection common.ConfigSelectionItem
	object         common.CompileObject
	err            error
}

func RenderObjects(objects []common.CompileObject, ns context.RenderNamespace, opts common.RenderOpts) (map[string]*bytes.Buffer, context.RenderNamespace, error) {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	filesState := make(map[string]fileRenderState)

	logger.Info("Run objects rendering")
	queue := buildRenderQueue(objects, opts.Selections)
	var postponed []renderQueueItem

	logger.Info("Objects selected", "objects", len(queue))
	for len(queue) > 0 {
		for _, item := range queue {
			logger.Debug("Render", "object", item.object.String())

			logger.Trace("-> Render file name expression", "object", item.object.String(), "template", item.selection.Render.File)
			fileName, err := renderObjectInlineTemplate(item, opts, item.selection.Render.File)
			switch {
			case errors.Is(err, context.ErrNotDefined):
				// Template can't be rendered right now due to unknown object definition, postpone it
				postponed = append(postponed, item)
				logger.Trace(
					"--> Postpone the file name rendering because some definitions it uses are not known yet",
					"object", item.object.String(),
				)
				continue
			case err != nil:
				return nil, ns, err
			}
			fileName = utils.NormalizePath(fileName)
			logger.Trace("-> File", "name", fileName)

			if _, ok := filesState[fileName]; !ok {
				pkgName, _ := lo.Coalesce(item.selection.Render.Package, utils.GetPackageName(path.Dir(fileName)))
				filesState[fileName] = fileRenderState{packageName: pkgName, fileName: fileName}
			}

			logger.Debug("-> Render", "object", item.object.String(), "file", fileName, "template", item.selection.Render.Template)
			newState, newNs, err := renderObject(item, opts, item.selection.Render.Template, filesState[fileName], ns)
			switch {
			case errors.Is(err, context.ErrNotDefined):
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
				return nil, ns, err
			}

			logger.Trace("--> Updated file state", "imports", newState.imports.String(), "namespace", newNs.String())
			ns = newNs
			filesState[fileName] = newState
		}
		if len(postponed) == len(queue) {
			return nil, ns, fmt.Errorf(
				"missed object definitions, please ensure they are defined by `godef` or `def` functions prior using: \n%w",
				errors.Join(lo.Map(postponed, func(item renderQueueItem, _ int) error { return item.err })...),
			)
		}
		if len(postponed) > 0 {
			logger.Trace("Process postponed objects", "objects", len(postponed))
		}
		queue, postponed = postponed, nil
	}

	logger.Debug("Render files", "files", len(filesState))
	res, err := renderFiles(filesState, opts)
	if err != nil {
		return res, ns, err
	}

	return res, ns, nil
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

func renderObject(
	item renderQueueItem,
	opts common.RenderOpts,
	templateName string,
	fileState fileRenderState,
	ns context.RenderNamespace,
) (fileRenderState, context.RenderNamespace, error) {
	importsCopy := fileState.imports.Clone()
	nsCopy := ns.Clone()

	ctx := &context.RenderContextImpl{
		RenderOpts:             opts,
		CurrentSelectionConfig: item.selection,
		PackageName:            fileState.packageName,
		PackageNamespace:       &nsCopy,
		Imports:                &importsCopy,
		Object:                 item.object,
	}
	tmpl.SetContext(ctx)

	tplCtx := tmpl.NewTemplateContext(ctx, item.object.Renderable, &importsCopy)

	// Execute the main template first to accumulate imports and other data, that will be rendered in preamble
	tpl, err := tmpl.LoadTemplate(templateName)
	if err != nil {
		return fileState, nsCopy, fmt.Errorf("template %q: %w", templateName, err)
	}

	var res bytes.Buffer
	if err := tpl.Execute(&res, tplCtx); err != nil {
		return fileState, nsCopy, err
	}

	// Update the file state if rendering was successful
	// If item is marked reused from other place, do not update the file state and content, just update the namespace
	if item.selection.ReusePackagePath == "" {
		fileState.buf.Write(res.Bytes())
		fileState.buf.WriteRune('\n') // Separate writes following each other (if any)
		fileState.imports = importsCopy
	}

	return fileState, nsCopy, nil
}

func renderFiles(files map[string]fileRenderState, opts common.RenderOpts) (map[string]*bytes.Buffer, error) {
	var res = make(map[string]*bytes.Buffer, len(files))
	logger := log.GetLogger(log.LoggerPrefixRendering)

	tpl, err := tmpl.LoadTemplate(opts.PreambleTemplate)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", opts.PreambleTemplate, err)
	}

	keys := lo.Keys(files)
	slices.Sort(keys)
	for _, fileName := range keys {
		state := files[fileName]
		logger.Debug("Render file", "file", fileName, "package", state.packageName, "imports", state.imports.String())
		if !bytes.ContainsFunc(state.buf.Bytes(), unicode.IsLetter) {
			logger.Debug("-> Skip empty file", "file", fileName)
			continue
		}
		b, err := renderObjectFileTemplate(tpl, opts, state)
		if err != nil {
			return nil, err
		}
		res[fileName] = b
	}

	return res, nil
}
