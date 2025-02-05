package main

import (
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/renderer"
	"github.com/samber/lo"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	toolchainCommand = "go"
	goGetSubcommand     = "get"
	goBuildSubcommand   = "build"
)

type ClientCmd struct {
	Spec string `arg:"required,positional" help:"AsyncAPI specification file path or url" placeholder:"PATH"`

	OutputFile string `arg:"-o,--output" help:"Executable output file path" placeholder:"PATH"`
	KeepSource bool   `arg:"--keep-source" help:"Do not automatically remove the generated code on exit"`
	ConfigFile string `arg:"-c,--config-file" help:"YAML configuration file path" placeholder:"PATH"`

	TemplateDir string `arg:"-T,--template-dir" help:"Directory with custom templates" placeholder:"DIR"`
	PreambleTemplate string `arg:"--preamble-template" help:"Custom preamble template name" placeholder:"NAME"`
	GoModTemplate string `arg:"--go-mod-template" help:"Custom go.mod template name" placeholder:"NAME"`

	RuntimeModule       string        `arg:"--runtime-module" help:"Runtime module name" placeholder:"MODULE"`
	AllowRemoteRefs bool `arg:"--allow-remote-refs" help:"Allow resolver to fetch the files from remote $ref URLs"`
	ResolverSearchDir   string        `arg:"--resolver-search-dir" help:"Directory to search the local spec files for [default: current working directory]" placeholder:"PATH"`
	ResolverTimeout time.Duration `arg:"--resolver-timeout" help:"Timeout for resolver to resolve a spec file" placeholder:"DURATION"`
	ResolverCommand string        `arg:"--resolver-command" help:"Custom resolver executable to use instead of built-in resolver" placeholder:"PATH"`
}

func cliClient(cmd *ClientCmd) error {
	logger := log.GetLogger("")
	// TODO: config file

	projectModule := lo.RandomString(10, lo.LowerCaseLettersCharset)
	targetDir, err := os.MkdirTemp("", "go-asyncapi-client-")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer func() {
		if !cmd.KeepSource {
			if err := os.RemoveAll(targetDir); err != nil {
				logger.Warn("remove directory", "error", err)
			}
		} else {
			logger.Info("Generated code location", "directory", targetDir)
		}
	}()

	logger.Debug("Generate the client code", "targetDir", targetDir, "module", projectModule)
	generateCmd := &GenerateCmd{
		TargetDir: targetDir,
		ConfigFile: cmd.ConfigFile,
		PubSub: &generatePubSubArgs{
			Spec:              cmd.Spec,
			ProjectModule:     projectModule,
			RuntimeModule:     cmd.RuntimeModule,
			TemplateDir:       cmd.TemplateDir,
			PreambleTemplate:  cmd.PreambleTemplate,
			AllowRemoteRefs:   cmd.AllowRemoteRefs,
			ResolverSearchDir: cmd.ResolverSearchDir,
			ResolverTimeout:   cmd.ResolverTimeout,
			ResolverCommand:   cmd.ResolverCommand,
			ClientApp:         true,
		},
	}
	if err = cliGenerate(generateCmd); err != nil {
		return fmt.Errorf("generate client code: %w", err)
	}

	outputFile := cmd.OutputFile
	if outputFile == "" {
		outputFile = "client"
		if runtime.GOOS == "windows" {
			outputFile += ".exe"
		}
	}
	outputFile, err = filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("resolve output file path: %w", err)
	}
	sourceFile := path.Join(targetDir, renderer.DefaultClientAppFileName)
	logger.Debug("Compile the executable", "sourceFile", sourceFile, "outputFile", outputFile)
	err = compileClientApp(sourceFile, outputFile)
	switch {
	case errors.Is(err, exec.ErrNotFound):
		logger.Error(
			"Go toolchain not found! Please install it by running `apt install golang-go` (Debian/Ubuntu), `brew install go` (Mac) or follow the instructions at https://golang.org/doc/install",
			"error", err,
		)
		return err
	case err != nil:
		return fmt.Errorf("run generated code: %w", err)
	}

	logger.Info("Client executable compiled", "file", outputFile)

	return nil
}

func compileClientApp(sourceFile, outputFile string) error {
	logger := log.GetLogger("")
	toolchainPath, err := exec.LookPath(toolchainCommand)
	if err != nil {
		return fmt.Errorf("find toolchain %q: %w", toolchainCommand, err)
	}
	logger.Debug("Go toolchain found", "path", toolchainPath)

	subcommands := [][]string{
		{goGetSubcommand, "./..."},
		{goBuildSubcommand, "-v", "-o", outputFile, sourceFile},
	}
	for _, subcommand := range subcommands {
		cmdLine := toolchainPath + " " + strings.Join(subcommand, " ")
		logger.Infof("Run %s", cmdLine)
		cmd := exec.Command(toolchainPath, subcommand...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = path.Dir(sourceFile)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q: %w", cmdLine, err)
		}
	}

	return nil
}