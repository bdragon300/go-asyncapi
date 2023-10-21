package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/protocols/amqp"

	"github.com/bdragon300/asyncapi-codegen/internal/protocols/kafka"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"

	"github.com/bdragon300/asyncapi-codegen/internal/linker"

	"github.com/alexflint/go-arg"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"gopkg.in/yaml.v3"
)

type GenerateCmd struct {
	SpecFile      string `arg:"required,positional" help:"AsyncAPI specification file, yaml or json"`
	TargetDir     string `arg:"-t,--target-dir" default:"./asyncapi" help:"Directory to save the generated code"`
	ModuleName    string `arg:"-M,--module-path" help:"Go module name to use. By default it is extracted from go.mod file in the current directory"`
	TargetPackage string `arg:"-T,--target-package" help:"Package for generated code. By default it is equal to the target directory"`
}

type cli struct {
	GenerateCmd *GenerateCmd `arg:"subcommand:generate" help:"Generate the code based on AsyncAPI specification"`
}

func main() {
	cliArgs := cli{}
	cliParser := arg.MustParse(&cliArgs)

	if cliArgs.GenerateCmd == nil {
		cliParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	registerProtocols()

	cmd := cliArgs.GenerateCmd
	if err := generate(cmd); err != nil {
		cliParser.Fail(fmt.Sprintf("Error: %v", err))
	}
}

func registerProtocols() {
	kafka.Register()
	amqp.Register()
}

func generate(cmd *GenerateCmd) error {
	var err error

	importBase := cmd.ModuleName
	if importBase == "" {
		if importBase, err = getImportBase(); err != nil {
			return fmt.Errorf("cannot extract module name from go.mod (you can specify it by -M argument): %w", err)
		}
	}
	targetPkg, _ := lo.Coalesce(cmd.TargetPackage, cmd.TargetDir)
	importBase = path.Join(importBase, targetPkg)

	spec, err := unmarshalSpecFile(cmd.SpecFile)
	if err != nil {
		return fmt.Errorf("error while reading spec file: %v", err)
	}

	localLinker := &linker.LocalLinker{}
	compileCtx := common.CompileContext{Packages: make(map[string]*common.Package), Linker: localLinker}
	if err = scan.CompileSchema(&compileCtx, reflect.ValueOf(spec)); err != nil {
		return fmt.Errorf("schema compile error: %v", err)
	}

	if err = localLinker.Process(&compileCtx); err != nil {
		return fmt.Errorf("schema linking error: %v", err)
	}

	files, err := render.AssemblePackages(compileCtx.Packages, importBase, cmd.TargetDir)
	if err != nil {
		return fmt.Errorf("schema assemble/render error: %v", err)
	}
	if err = render.WriteFiles(files, cmd.TargetDir); err != nil {
		return fmt.Errorf("error while writing code to files: %v", err)
	}

	return nil
}

func unmarshalSpecFile(fileName string) (*compile.AsyncAPI, error) {
	res := compile.AsyncAPI{}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	switch {
	case strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml"):
		dec := yaml.NewDecoder(f)
		err = dec.Decode(&res)
	case strings.HasSuffix(fileName, ".json"):
		dec := json.NewDecoder(f)
		err = dec.Decode(&res)
	default:
		err = errors.New("cannot determine format of a spec file: unknown filename extension")
	}
	return &res, err
}

func getImportBase() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current working directory: %w", err)
	}
	fn := path.Join(pwd, "go.mod")
	f, err := os.Open(fn)
	if err != nil {
		return "", fmt.Errorf("unable to open %q: %w", fn, err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("unable read %q file: %w", fn, err)
	}
	modpath := modfile.ModulePath(data)
	if modpath == "" {
		return "", fmt.Errorf("module path not found in %q", fn)
	}
	return modpath, nil
}
