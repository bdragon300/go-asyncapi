package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/log"
	chlog "github.com/charmbracelet/log"
	"github.com/samber/lo"
)

// Go toolchain command and its subcommands called to build the client executable
const (
	toolchainCommand  = "go"
	goGetSubcommand   = "get"
	goBuildSubcommand = "build"
)

type ClientCmd struct {
	Document string `arg:"required,positional" help:"AsyncAPI document file or url" placeholder:"FILE"`

	ConfigFile       string `arg:"-c,--config-file" help:"YAML configuration file path" placeholder:"FILE"`
	OutputExecFile   string `arg:"-o,--output" help:"Executable output file name" placeholder:"FILE"`
	OutputSourceFile string `arg:"--output-source" help:"Source code output file path" placeholder:"FILE"`
	KeepSource       bool   `arg:"--keep-source" help:"Do not automatically remove the generated code on exit"`

	TemplateDir      string `arg:"-T,--template-dir" help:"User templates directory" placeholder:"DIR"`
	TempDir          string `arg:"--temp-dir" help:"Temporary directory to store the generated code. Implies --keep-source as well" placeholder:"DIR"`
	PreambleTemplate string `arg:"--preamble-template" help:"Preamble template name" placeholder:"NAME"`
	GoModTemplate    string `arg:"--go-mod-template" help:"Custom go.mod template name" placeholder:"NAME"`

	RuntimeModule   string        `arg:"--runtime-module" help:"Runtime module name" placeholder:"MODULE"`
	AllowRemoteRefs bool          `arg:"--allow-remote-refs" help:"Allow locator to fetch the files from remote $ref URLs"`
	LocatorRootDir  string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout  time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand  string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`
}

func cliClient(cmd *ClientCmd, globalConfig toolConfig) error {
	logger := log.GetLogger("")
	cmdConfig := cliClientMergeConfig(globalConfig, cmd)

	projectModule := lo.RandomString(10, lo.LowerCaseLettersCharset)
	targetDir := cmd.TempDir
	if targetDir == "" {
		var err error
		targetDir, err = os.MkdirTemp("", "go-asyncapi-client-")
		if err != nil {
			return fmt.Errorf("create temporary directory: %w", err)
		}
		defer func() {
			if !cmdConfig.Client.KeepSource {
				if err := os.RemoveAll(targetDir); err != nil {
					logger.Warn("remove directory", "error", err)
				}
			}
		}()
	}
	logger.Debug("Generated code location", "directory", targetDir)

	// Generate the client code
	logger.Debug("Generate the client code", "targetDir", targetDir, "module", projectModule)
	generateCmd := &CodeCmd{
		TargetDir:        targetDir,
		Document:         cmd.Document,
		ProjectModule:    projectModule,
		RuntimeModule:    cmdConfig.RuntimeModule,
		TemplateDir:      cmdConfig.Code.TemplatesDir,
		PreambleTemplate: cmdConfig.Code.PreambleTemplate,
		AllowRemoteRefs:  cmdConfig.Locator.AllowRemoteReferences,
		LocatorRootDir:   cmdConfig.Locator.RootDirectory,
		LocatorTimeout:   cmdConfig.Locator.Timeout,
		LocatorCommand:   cmdConfig.Locator.Command,
		ClientApp:        true,
		goModTemplate:    cmdConfig.Client.GoModTemplate,
	}
	if err := cliCode(generateCmd, cmdConfig); err != nil {
		return fmt.Errorf("generate client code: %w", err)
	}

	outputFile := cmdConfig.Client.OutputFile
	if outputFile == "" {
		outputFile = "client"
		if runtime.GOOS == "windows" {
			outputFile += ".exe"
		}
	}
	absoluteOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("output file path: %w", err)
	}

	// Run Go build
	sourceFile := path.Join(targetDir, cmdConfig.Client.OutputSourceFile)
	logger.Debug("Compile the executable", "sourceFile", sourceFile, "outputFile", absoluteOutputFile)
	err = runGoBuild(sourceFile, absoluteOutputFile)
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

	logger.Infof("Client executable saved to %q", outputFile)

	return nil
}

func runGoBuild(sourceFile, outputFile string) error {
	logger := log.GetLogger("")
	toolchainPath, err := exec.LookPath(toolchainCommand)
	if err != nil {
		return fmt.Errorf("find toolchain %q: %w", toolchainCommand, err)
	}
	logger.Debug("Go toolchain found", "path", toolchainPath)

	subcommands := [][]string{
		{goGetSubcommand, "./..."},
		{goBuildSubcommand, "-o", outputFile, sourceFile},
	}
	if chlog.GetLevel() <= chlog.DebugLevel {
		subcommands = [][]string{
			{goGetSubcommand, "./..."},
			{goBuildSubcommand, "-v", "-o", outputFile, sourceFile},
		}
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

func cliClientMergeConfig(globalConfig toolConfig, cmd *ClientCmd) toolConfig {
	res := globalConfig

	res.Client.OutputFile = coalesce(cmd.OutputExecFile, globalConfig.Client.OutputFile)
	res.Client.OutputSourceFile = coalesce(cmd.OutputSourceFile, globalConfig.Client.OutputSourceFile)
	res.Client.KeepSource = coalesce(cmd.KeepSource, globalConfig.Client.KeepSource)
	res.Client.GoModTemplate = coalesce(cmd.GoModTemplate, globalConfig.Client.GoModTemplate)
	res.Client.TempDir = coalesce(cmd.TempDir, globalConfig.Client.TempDir)

	res.Code.TemplatesDir = coalesce(cmd.TemplateDir, globalConfig.Code.TemplatesDir)
	res.Code.PreambleTemplate = coalesce(cmd.PreambleTemplate, globalConfig.Code.PreambleTemplate)

	res.RuntimeModule = coalesce(cmd.RuntimeModule, globalConfig.RuntimeModule)
	res.Locator.AllowRemoteReferences = coalesce(cmd.AllowRemoteRefs, globalConfig.Locator.AllowRemoteReferences)
	res.Locator.RootDirectory = coalesce(cmd.LocatorRootDir, globalConfig.Locator.RootDirectory)
	res.Locator.Timeout = coalesce(cmd.LocatorTimeout, globalConfig.Locator.Timeout)
	res.Locator.Command = coalesce(cmd.LocatorCommand, globalConfig.Locator.Command)

	return res
}
