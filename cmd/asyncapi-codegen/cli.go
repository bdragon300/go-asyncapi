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

	"github.com/bdragon300/asyncapi-codegen/implementations"

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
	Spec          string `arg:"required,positional" help:"AsyncAPI specification file, yaml or json" placeholder:"FILE"`
	TargetDir     string `arg:"-t,--target-dir" default:"./asyncapi" help:"Directory to save the generated code" placeholder:"DIR"`
	ImplDir       string `arg:"--impl-dir" help:"Directory where protocol implementations will be placed. By default it is {target-dir}/impl" placeholder:"DIR"`
	ProjectModule string `arg:"-M,--project-module" help:"Project module name to use. By default it is extracted from go.mod file in the current working directory" placeholder:"MODULE"`
	TargetPackage string `arg:"-T,--target-package" help:"Package for generated code. By default it is equal to the target directory" placeholder:"NAME"`
	ImplementationsOpts
}

type ImplementationsOpts struct {
	Kafka string `arg:"--kafka-impl" default:"franz-go" help:"Implementation for kafka protocol" placeholder:"NAME"`
}

type cli struct {
	GenerateCmd         *GenerateCmd `arg:"subcommand:generate" help:"Generate the code based on AsyncAPI specification"`
	ListImplementations *struct{}    `arg:"subcommand:list-implementations" help:"Show all available protocol implementations"`
}

func main() {
	cliArgs := cli{}
	cliParser := arg.MustParse(&cliArgs)

	if cliArgs.ListImplementations != nil {
		listImplementations()
		return
	}

	if cliArgs.GenerateCmd == nil {
		cliParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	registerProtocols()

	cmd := cliArgs.GenerateCmd
	if err := generate(cmd); err != nil {
		cliParser.Fail(err.Error())
	}
}

func registerProtocols() {
	kafka.Register()
	amqp.Register()
}

func listImplementations() {
	manifest, err := getImplementationsManifest()
	if err != nil {
		panic(err.Error())
	}
	for proto, implInfo := range manifest {
		os.Stdout.WriteString(proto + ":\n")
		for implName, info := range implInfo {
			os.Stdout.WriteString(fmt.Sprintf("* %s -- %s\n", implName, info.URL))
		}
	}
}

func generate(cmd *GenerateCmd) error {
	var err error

	importBase := cmd.ProjectModule
	if importBase == "" {
		if importBase, err = getImportBase(); err != nil {
			return fmt.Errorf("cannot extract module name from go.mod (you can specify it by -M argument): %w", err)
		}
	}
	targetPkg, _ := lo.Coalesce(cmd.TargetPackage, cmd.TargetDir)
	importBase = path.Join(importBase, targetPkg)
	implDir, _ := lo.Coalesce(cmd.ImplDir, path.Join(cmd.TargetDir, "impl"))

	spec, err := unmarshalSpecFile(cmd.Spec)
	if err != nil {
		return fmt.Errorf("error while reading spec file: %v", err)
	}

	localLinker := &linker.LocalLinker{}
	compileCtx := common.NewCompileContext(localLinker)

	// Compilation
	if err = scan.CompileSchema(compileCtx, reflect.ValueOf(spec)); err != nil {
		return fmt.Errorf("schema compile error: %v", err)
	}

	// Linking
	if err = localLinker.Process(compileCtx); err != nil {
		return fmt.Errorf("schema linking error: %v", err)
	}

	// Assembling
	files, err := render.AssemblePackages(compileCtx.Packages, importBase, cmd.TargetDir)
	if err != nil {
		return fmt.Errorf("schema assemble/render error: %v", err)
	}

	// Rendering
	if err = render.WriteAssembled(files, cmd.TargetDir); err != nil {
		return fmt.Errorf("error while writing code to files: %v", err)
	}

	// Rendering implementations
	implManifest, err := getImplementationsManifest()
	if err != nil {
		panic(err.Error())
	}
	selectedImpls := getSelectedImplementations(cmd.ImplementationsOpts)
	for p := range compileCtx.Protocols {
		if _, ok := implManifest[p][selectedImpls[p]]; !ok {
			return fmt.Errorf("unknown implementation %s for %s protocol, use list-implementations command to see possible values", selectedImpls[p], p)
		}
		d := path.Join(implDir, p)
		if err = os.MkdirAll(d, 0o750); err != nil {
			return fmt.Errorf("cannot create directory %q: %w", d, err)
		}
		if err = render.WriteImplementation(implManifest[p][selectedImpls[p]].Dir, d); err != nil {
			return fmt.Errorf("cannot render implementation for protocol %q: %w", p, err)
		}
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

func getSelectedImplementations(opts ImplementationsOpts) map[string]string {
	return map[string]string{
		"kafka": opts.Kafka,
	}
}

func getImplementationsManifest() (implementations.ImplManifest, error) {
	f, err := implementations.Implementations.Open("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("cannot open manifest.json: %w", err)
	}
	dec := json.NewDecoder(f)
	var meta implementations.ImplManifest
	if err = dec.Decode(&meta); err != nil {
		return nil, fmt.Errorf("cannot parse manifest.json: %w", err)
	}

	return meta, nil
}
