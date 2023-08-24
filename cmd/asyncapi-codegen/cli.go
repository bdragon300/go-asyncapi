package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/packages"
	"github.com/samber/lo"

	"github.com/alexflint/go-arg"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/bdragon300/asyncapi-codegen/internal/schema"
	"gopkg.in/yaml.v3"
)

type GenerateCmd struct{}

type cli struct {
	Spec     string       `arg:"required,--spec" help:"AsyncAPI spec file path"`
	OutDir   string       `arg:"--out-dir" default:"./generated" help:"Directory where the generated code will be placed"`
	Generate *GenerateCmd `arg:"subcommand:generate"`
}

func main() {
	cliArgs := cli{}
	arg.MustParse(&cliArgs)

	if cliArgs.Generate == nil {
		panic("No command given")
	}

	f, err := os.Open(cliArgs.Spec)
	if err != nil {
		panic(err)
	}
	jsonBuf, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	_ = f.Close()

	specBuf := schema.AsyncAPI{}
	err = yaml.Unmarshal(jsonBuf, &specBuf)
	if err != nil {
		panic(err)
	}

	modelsBucket := packages.ModelsPackage{}
	messageBucket := packages.MessagePackage{}
	scanPackages := map[common.PackageKind]scan.Package{
		common.ModelsPackageKind:  &modelsBucket,
		common.MessagePackageKind: &messageBucket,
	}
	scanCtx := scan.Context{Packages: scanPackages, RefMgr: scan.NewRefManager()}
	if err = scan.WalkSchema(&scanCtx, reflect.ValueOf(specBuf)); err != nil {
		panic(err)
	}

	scanCtx.RefMgr.ProcessRefs(&scanCtx)

	if err = ensureDir(cliArgs.OutDir); err != nil {
		panic(err)
	}

	files1, err := packages.RenderModels(&modelsBucket, cliArgs.OutDir)
	if err != nil {
		panic(err)
	}
	files2, err := packages.RenderMessages(&messageBucket, cliArgs.OutDir)
	if err != nil {
		panic(err)
	}
	files := lo.Assign(files1, files2)
	for fileName, fileObj := range files {
		fullPath := path.Join(cliArgs.OutDir, fileName)
		if err = ensureDir(path.Dir(fullPath)); err != nil {
			panic(err)
		}
		if err = fileObj.Save(fullPath); err != nil {
			panic(err)
		}
	}
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
