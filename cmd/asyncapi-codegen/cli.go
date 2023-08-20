package main

import (
	"fmt"
	"github.com/bdragon300/asyncapi-codegen/internal/buckets"
	"io"
	"os"
	"path"
	"reflect"

	"github.com/alexflint/go-arg"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/renderer"
	"github.com/bdragon300/asyncapi-codegen/internal/scanner"
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

	typeBucket := buckets.Schema{}
	scanBuckets := map[common.BucketKind]scanner.Bucket{
		common.BucketSchema: &typeBucket,
	}
	scanCtx := scanner.Context{Buckets: scanBuckets, RefMgr: scanner.NewRefManager()}
	if err = scanner.WalkSchema(&scanCtx, reflect.ValueOf(specBuf)); err != nil {
		panic(err)
	}

	scanCtx.RefMgr.ProcessRefs(&scanCtx)

	if err = ensureDir(cliArgs.OutDir); err != nil {
		panic(err)
	}

	files, err := renderer.RenderTypes(&typeBucket, cliArgs.OutDir)
	if err != nil {
		panic(err)
	}
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
