package writer

import (
	"bytes"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tpl"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	"os"
	"path"
	"text/template"

	"github.com/bdragon300/go-asyncapi/internal/compiler"
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

type renderSource interface {
	AllObjects() []compiler.Object
}

func RenderPackages(source renderSource, opts common.RenderOpts) (fileContents map[string]*bytes.Buffer, err error) {
	fileContents = make(map[string]*bytes.Buffer)
	// TODO: logging
	for _, selection := range opts.Selections {
		objects := selector.SelectObjects(source.AllObjects(), selection.RenderSelectionFilterConfig)
		if len(objects) == 0 {
			continue
		}

		ctx := &context.RenderContextImpl{RenderOpts: opts}
		selectionObjects := lo.Map(objects, func(item compiler.Object, _ int) common.Renderable { return item.Object})
		tplCtx := render.NewTemplateContext(ctx, render.TemplateSelections{Objects: selectionObjects})
		context.Context = ctx
		// TODO: template in file name
		if _, ok := fileContents[selection.File]; !ok {
			fileContents[selection.File] = &bytes.Buffer{}
		}

		// TODO: redefinition preambule in config/cli args
		var tmpl *template.Template
		if tmpl = tpl.LoadTemplate("preamble"); tmpl == nil {
			return nil, fmt.Errorf("template not found: preamble")
		}
		tmpl = tmpl.Funcs(render.GetTemplateFunctions(ctx))
		if err = tmpl.Execute(fileContents[selection.File], tplCtx); err != nil {
			return
		}

		if tmpl = tpl.LoadTemplate(selection.Template); tmpl == nil {
			return nil, fmt.Errorf("template not found: %s", selection.Template)
		}
		tmpl = tmpl.Funcs(render.GetTemplateFunctions(ctx))
		if err = tmpl.Execute(fileContents[selection.File], tplCtx); err != nil {
			return
		}
	}

	return
}

//func RenderPackages(source renderSource, opts common.RenderOpts) (fileContents map[string]*bytes.Buffer, err error) {
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
//				fileName = utils.ToFileName(item.Object.ID()) + ".go"
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
