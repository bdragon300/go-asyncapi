package writer

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/dave/jennifer/jen"
)

const GeneratedCodePreamble = "Code generated by asyncapi-codegen-go tool. DO NOT EDIT."

type MultilineError struct {
	error
}

func (e MultilineError) Error() string {
	s := e.error.Error()
	i := strings.IndexRune(s, '\n')
	if i < 0 {
		return s
	}
	return s[:i]
}

func (e MultilineError) RestLines() string {
	lineno := 1
	bld := strings.Builder{}
	rd := bufio.NewReader(strings.NewReader(e.error.Error()))
	_, _ = rd.ReadString('\n') // Skip the first line

	for {
		s, err := rd.ReadString('\n')
		if err != nil {
			break // Suppose that the only error here can appear is io.EOF
		}
		bld.WriteString(fmt.Sprintf("%-3d| ", lineno))
		bld.WriteString(s)
		lineno++
	}

	return bld.String()
}

func RenderPackages(packages map[string]*common.Package, importBase, baseDir string) (files map[string]*bytes.Buffer, err error) {
	files = make(map[string]*bytes.Buffer)
	logger := common.NewLogger("Rendering 🎨")
	counter := 0

	for pkgName, pkg := range packages {
		ctx := &common.RenderContext{
			CurrentPackage: pkgName,
			ImportBase:     importBase,
			Logger:         logger,
		}
		f := jen.NewFilePathName(baseDir, pkgName)
		f.HeaderComment(GeneratedCodePreamble)
		ctx.Logger.Debug("Package", "pkg", pkgName, "items", len(pkg.Items()))
		for _, item := range pkg.Items() {
			counter++
			if !item.Typ.DirectRendering() {
				continue
			}
			for _, stmt := range item.Typ.RenderDefinition(ctx) {
				f.Add(stmt)
			}
			f.Add(jen.Line())
		}

		buf := &bytes.Buffer{}
		if err = f.Render(buf); err != nil {
			if strings.ContainsRune(err.Error(), '\n') {
				return files, MultilineError{err}
			}
			return files, err
		}

		fileName := pkgName + ".go"
		ctx.Logger.Debug("Package rendered as file", "pkg", pkgName, "file", fileName, "bytes", buf.Len())
		files[path.Join(pkgName, fileName)] = buf
	}
	logger.Info("Finished", "packages", len(packages), "objects", counter)
	return
}

func WriteToFiles(files map[string]*bytes.Buffer, baseDir string) error {
	l := common.NewLogger("Writing 📝")

	if err := ensureDir(baseDir); err != nil {
		return err
	}
	totalBytes := 0
	for fileName, buf := range files {
		l.Debug("File", "name", fileName)
		fullPath := path.Join(baseDir, fileName)
		if err := ensureDir(path.Dir(fullPath)); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, buf.Bytes(), 0o644); err != nil {
			return err
		}
		l.Debug("File wrote", "name", fullPath, "bytes", buf.Len())
		totalBytes += buf.Len()
	}
	l.Info("Finished", "files", len(files), "total_bytes", totalBytes)
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
