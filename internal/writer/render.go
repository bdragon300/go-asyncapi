package writer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"go/format"
	"os"
	"path"
	"text/template"
)

//type MultilineError struct {
//	error
//}
//
//func (e MultilineError) Error() string {
//	s := e.error.Error()
//	i := strings.IndexRune(s, '\n')
//	if i < 0 {
//		return s
//	}
//	return s[:i]
//}
//
//func (e MultilineError) RestLines() string {
//	lineno := 1
//	bld := strings.Builder{}
//	rd := bufio.NewReader(strings.NewReader(e.error.Error()))
//	_, _ = rd.ReadString('\n') // Skip the first line
//
//	for {
//		s, err := rd.ReadString('\n')
//		if err != nil {
//			break // Suppose that the only error here can appear is io.EOF
//		}
//		bld.WriteString(fmt.Sprintf("%-3d| ", lineno))
//		bld.WriteString(s)
//		lineno++
//	}
//
//	return bld.String()
//}

const defaultPreambleTemplateName = "preamble.tmpl"

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

func RenderFiles(source renderSource, opts common.RenderOpts) (map[string]*bytes.Buffer, error) {
	filesState := make(map[string]fileRenderState)
	ns := context.RenderNamespace{}
	// TODO: logging
	queue := buildRenderQueue(source, opts.Selections)
	var postponed []renderQueueItem

	for len(queue) > 0 {
		for _, item := range queue {
			fileName, err := renderInlineTemplate(item, opts, item.selection.File)
			switch {
			case errors.Is(err, context.ErrDefinitionLocationUnknown):
				// Template can't be rendered right now due to unknown object definition, postpone it
				postponed = append(postponed, item)
				continue
			case err != nil:
				return nil, err
			}
			fileName = utils.NormalizePath(fileName)

			if _, ok := filesState[fileName]; !ok {
				filesState[fileName] = fileRenderState{packageName: item.selection.Package, fileName: fileName}
			}
			newState, newNs, err := renderObject(item, opts, item.selection.Template, filesState[fileName], ns)
			switch {
			case errors.Is(err, context.ErrDefinitionLocationUnknown):
				// Some objects needed by template code have not been defined and therefore, not in namespace yet.
				// Postpone this run to the end in hope that next runs will define these objects.
				item.err = err
				postponed = append(postponed, item)
				continue
			case err != nil:
				return nil, err
			}

			ns = newNs
			filesState[fileName] = newState
		}
		if len(postponed) == len(queue) {
			return nil, fmt.Errorf(
				"missed object definitions, please ensure they are rendered by godef: %w",
				errors.Join(lo.Map(postponed, func(item renderQueueItem, _ int) error { return item.err })...),
			)
		}
		queue, postponed = postponed, nil
	}

	res, err := renderFiles(filesState, opts)
	if err != nil {
		return res, err
	}

	return res, nil
}

func buildRenderQueue(source renderSource, selections []common.RenderSelectionConfig) (res []renderQueueItem) {
	for _, selection := range selections {
		objects := selector.SelectObjects(source.AllObjects(), selection)
		for _, obj := range objects {
			res = append(res, renderQueueItem{selection: selection, object: obj})
		}
	}
	return
}

func renderInlineTemplate(item renderQueueItem, opts common.RenderOpts, text string) (string, error) {
	ctx := &context.RenderContextImpl{
		RenderOpts:             opts,
		CurrentSelectionConfig: item.selection,
		PackageName:            item.selection.Package,
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
	if err = tpl.Execute(&res, tplCtx.Object()); err != nil {
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
	var tpl *template.Template
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
	if tpl = tmpl.LoadTemplate(templateName); tpl == nil {
		return fileState, nsCopy, fmt.Errorf("template %q not found", templateName)
	}

	var res bytes.Buffer
	if err := tpl.Execute(&res, tplCtx); err != nil {
		return fileState, nsCopy, err
	}

	// Update the file state if rendering was successful
	fileState.buf.Write(res.Bytes())
	fileState.imports = importsCopy
	fileState.buf.WriteRune('\n') // Separate writes following each other (if any)
	return fileState, nsCopy, nil
}

func renderFiles(files map[string]fileRenderState, opts common.RenderOpts) (map[string]*bytes.Buffer, error) {
	var res = make(map[string]*bytes.Buffer, len(files))

	// TODO: redefinition preamble in config/cli args
	tpl := tmpl.LoadTemplate(defaultPreambleTemplateName)
	if tpl == nil {
		return nil, fmt.Errorf("template %q not found", defaultPreambleTemplateName)
	}

	for fileName, state := range files {
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

//func RenderFiles(source renderSource, opts common.RenderOpts) (fileContents map[string]*bytes.Buffer, err error) {
//	fileContents = make(map[string]*bytes.Buffer)
//	logger := types.NewLogger("Rendering 🎨")
//	rendered := 0
//	totalObjects := 0
//
//	logger.Info("Run rendering")
//
//	files := make(map[string]*jen.File)
//	for _, pkgName := range source.Packages() {
//		ctx := &common.RenderContext{
//			CurrentPackage: pkgName,
//			RenderOpts:     opts,
//		}
//		items := source.PackageObjects(pkgName)
//		targetPkg := pkgName
//		if opts.PackageScope == common.PackageScopeAll {
//			targetPkg = opts.TargetPackage
//		}
//		logger.Debug("Package", "pkg", targetPkg, "items", len(items))
//		totalObjects += len(items)
//		for _, item := range items {
//			if !item.Object.DirectRendering() {
//				continue
//			}
//
//			fileName := pkgName + ".go" // All objects with the same type in one file
//			if opts.FileScope == common.FileScopeName {
//				// Every single object in a separate file
//				fileName = utils.NormalizePath(item.Object.ID()) + ".go"
//			}
//			if opts.PackageScope == common.PackageScopeType {
//				fileName = path.Join(targetPkg, fileName)
//			}
//
//			f, ok := files[fileName]
//			if !ok {
//				f = jen.NewFilePathName(opts.ImportBase, targetPkg)
//				f.HeaderComment(GeneratedCodePreamble)
//			}
//
//			rendered++
//			ctx.Logger.Debug("Render object", "pkg", pkgName, "object", item.Object.String(), "file", fileName)
//			func() {
//				// catch panics produced by rendering
//				defer func() {
//					if r := recover(); r != nil {
//						err = fmt.Errorf("%s: %s\n%v", item.Object.String(), debug.Stack(), r)
//					}
//				}()
//				for _, stmt := range item.Object.RenderDefinition(ctx) {
//					f.Add(stmt)
//				}
//			}()
//			if err != nil {
//				return
//			}
//			files[fileName] = f
//		}
//	}
//	logger.Debugf("Render stats: packages %d, objects: %d (rendered directly: %d)", len(source.Packages()), totalObjects, rendered)
//
//	for fileName, f := range files {
//		logger.Trace("Render file", "file", fileName)
//		buf := &bytes.Buffer{}
//		if b, ok := fileContents[fileName]; ok {
//			buf.WriteRune('\n')
//			buf = b
//		}
//		if err = f.Render(buf); err != nil {
//			if strings.ContainsRune(err.Error(), '\n') {
//				return fileContents, MultilineError{err}
//			}
//			return fileContents, err
//		}
//		logger.Debug("Rendered file", "file", fileName, "size", buf.Len())
//
//		fileContents[fileName] = buf
//	}
//
//	logger.Info("Rendering completed", "objects", rendered)
//	return
//}

// FormatFiles formats the files in-place in the map using gofmt
func FormatFiles(files map[string]*bytes.Buffer) error {
	logger := types.NewLogger("Formatting 📐")
	logger.Info("Run formatting")

	for fileName, buf := range files {
		logger.Debug("File", "name", fileName)
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return err
		}
		buf.Reset()
		buf.Write(formatted)
		logger.Debug("File formatted", "name", fileName, "bytes", buf.Len())
	}

	logger.Info("Formatting completed", "files", len(files))
	return nil
}

func WriteToFiles(files map[string]*bytes.Buffer, baseDir string) error {
	logger := types.NewLogger("Writing 📝")
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
		logger.Debug("File wrote", "name", fullPath, "bytes", buf.Len())
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
