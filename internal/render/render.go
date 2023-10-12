package render

import (
	"fmt"
	"os"
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

func AssemblePackages(packages map[string]*common.Package, importBase, baseDir string) (files map[string]*jen.File, err error) {
	files = make(map[string]*jen.File)

	for pkgName, pkg := range packages {
		ctx := &common.AssembleContext{
			CurrentPackage: pkgName,
			ImportBase:     importBase,
		}
		f := jen.NewFilePathName(baseDir, pkgName)
		for _, item := range pkg.Items() {
			if !item.Typ.AllowRender() {
				continue
			}
			for _, stmt := range item.Typ.AssembleDefinition(ctx) {
				f.Add(stmt)
			}
			f.Add(jen.Line())
		}

		files[path.Join(pkgName, pkgName+".go")] = f
	}
	return
}

func WriteFiles(files map[string]*jen.File, baseDir string) error {
	if err := ensureDir(baseDir); err != nil {
		return err
	}
	for fileName, fileObj := range files {
		fullPath := path.Join(baseDir, fileName)
		if err := ensureDir(path.Dir(fullPath)); err != nil {
			return err
		}
		if err := fileObj.Save(fullPath); err != nil {
			return err
		}
	}
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
		return fmt.Errorf("path %q is already exists and is not a directory", path)
	}

	return nil
}
