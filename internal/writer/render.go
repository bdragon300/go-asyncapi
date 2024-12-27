package writer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"go/format"
	"os"
	"path"
	"slices"
	"text/template"
)

type renderSource interface {
	AllObjects() []common.CompileObject
}

type fileRenderState struct {
	imports     context.ImportsList
	packageName string
	fileName   string
	buf        bytes.Buffer
}

type renderQueueItem struct {
	selection common.RenderSelectionConfig
	object         common.CompileObject
	err            error
}

func RenderObjects(objects []common.CompileObject, opts common.RenderOpts) (map[string]*bytes.Buffer, error) {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	filesState := make(map[string]fileRenderState)
	ns := context.RenderNamespace{}

	logger.Info("Run rendering")
	queue := buildRenderQueue(objects, opts.Selections)
	var postponed []renderQueueItem

	logger.Info("Objects selected", "objects", len(queue))
	for len(queue) > 0 {
		for _, item := range queue {
			logger.Debug("Render", "object", item.object.String())

			logger.Trace("-> Render file name", "object", item.object.String(), "template", item.selection.Render.File)
			fileName, err := renderInlineTemplate(item, opts, item.selection.Render.File)
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
				return nil, err
			}
			fileName = utils.NormalizePath(fileName)
			logger.Trace("-> File", "name", fileName)

			if _, ok := filesState[fileName]; !ok {
				filesState[fileName] = fileRenderState{packageName: item.selection.Render.Package, fileName: fileName}
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
				return nil, err
			}
			logger.Trace("--> Updated file state", "imports", newState.imports.String(), "namespace", newNs.String())

			ns = newNs
			filesState[fileName] = newState
		}
		if len(postponed) == len(queue) {
			return nil, fmt.Errorf(
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
		return res, err
	}

	return res, nil
}

func buildRenderQueue(objects []common.CompileObject, selections []common.RenderSelectionConfig) (res []renderQueueItem) {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	for _, selection := range selections {
		logger.Debug("Select objects", "selection", selection)
		objects := selector.SelectObjects(objects, selection)
		for _, obj := range objects {
			logger.Debug("-> Selected", "object", obj)
			res = append(res, renderQueueItem{selection: selection, object: obj})
		}
	}
	return
}

func renderInlineTemplate(item renderQueueItem, opts common.RenderOpts, text string) (string, error) {
	ctx := &context.RenderContextImpl{
		RenderOpts:             opts,
		CurrentSelectionConfig: item.selection,
		PackageName:            item.selection.Render.Package,
		Imports:                &context.ImportsList{},
		Object:                 item.object,
	}
	common.SetContext(ctx)

	tplCtx := tmpl.NewTemplateContext(ctx, item.object.Renderable, ctx.Imports)

	var res bytes.Buffer
	tpl, err := template.New("").Funcs(tmpl.GetTemplateFunctions()).Parse(text)
	if err != nil {
		return "", err
	}
	if err = tpl.Execute(&res, tplCtx); err != nil {
		return "", err
	}
	return res.String(), nil
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
	common.SetContext(ctx)

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
	// If item is marked reused from other place, do not update the file state and content, just update namespace
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
		if state.buf.Len() == 0 {
			logger.Debug("-> Skip empty file", "file", fileName)
			continue
		}
		b, err := renderFile(tpl, opts, state)
		if err != nil {
			return nil, err
		}
		res[fileName] = b
	}

	return res, nil
}

func renderFile(preambleTpl *template.Template, opts common.RenderOpts, renderState fileRenderState) (*bytes.Buffer, error) {
	var res bytes.Buffer

	ctx := &context.RenderContextImpl{
		RenderOpts:  opts,
		PackageName: renderState.packageName,
		Imports:     &renderState.imports,
	}
	common.SetContext(ctx)
	tplCtx := tmpl.NewTemplateContext(ctx, nil, &renderState.imports)

	if err := preambleTpl.Execute(&res, tplCtx); err != nil {
		return nil, err
	}
	res.WriteRune('\n')
	if _, err := res.Write(renderState.buf.Bytes()); err != nil {
		return nil, err
	}

	return &res, nil
}

// FormatFiles formats the files in-place in the map using gofmt
func FormatFiles(files map[string]*bytes.Buffer) error {
	logger := log.GetLogger(log.LoggerPrefixFormatting)
	logger.Info("Run formatting")

	keys := lo.Keys(files)
	slices.Sort(keys)
	for _, fileName := range keys {
		buf := files[fileName]
		logger.Debug("File", "name", fileName, "bytes", buf.Len())
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return types.ErrorWithContent{err, buf.Bytes()}
		}
		buf.Reset()
		buf.Write(formatted)
		logger.Debug("-> File formatted", "name", fileName, "bytes", buf.Len())
	}

	logger.Info("Formatting completed", "files", len(files))
	return nil
}

func WriteToFiles(files map[string]*bytes.Buffer, baseDir string) error {
	logger := log.GetLogger(log.LoggerPrefixWriting)
	logger.Info("Run writing")

	if err := ensureDir(baseDir); err != nil {
		return err
	}
	totalBytes := 0
	for fileName, buf := range files {
		logger.Debug("File", "name", fileName)
		fullPath := path.Join(baseDir, fileName)
		if err := ensureDir(path.Dir(fullPath)); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, buf.Bytes(), 0o644); err != nil {
			return err
		}
		logger.Debug("-> File wrote", "name", fullPath, "bytes", buf.Len())
		totalBytes += buf.Len()
	}
	logger.Debugf("Writer stats: files: %d, total bytes: %d", len(files), totalBytes)

	logger.Info("Writing completed", "files", len(files))
	return nil
}

func ensureDir(path string) error {
	if info, err := os.Stat(path); os.IsNotExist(err) {
		if err2 := os.MkdirAll(path, 0o755); err2 != nil {
			return err2
		}
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	return nil
}
