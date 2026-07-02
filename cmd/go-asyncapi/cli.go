package main

import (
	"errors"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
	"os"
	"path"

	"github.com/bdragon300/go-asyncapi/assets"
	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/doc"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	chlog "github.com/charmbracelet/log"

	"github.com/alexflint/go-arg"
)

type cli struct {
	CodeCmd             *CodeCmd    `arg:"subcommand:code" help:"Generate the Go boilerplate code"`
	ClientCmd           *ClientCmd  `arg:"subcommand:client" help:"Build the client executable (requires Go toolchain installed)"`
	InfraCmd            *InfraCmd   `arg:"subcommand:infra" help:"Generate the infrastructure setup files"`
	DiagramCmd          *DiagramCmd `arg:"subcommand:diagram" help:"Generate the architecture diagram"`
	UICmd               *UICmd      `arg:"subcommand:ui" help:"Generate or serve the documentation UI"`
	DocCmd              *doc.Cmd    `arg:"subcommand:doc" help:"Manipulate AsyncAPI documents: merge, validate, etc."`
	ListImplementations *struct{}   `arg:"subcommand:list-implementations" help:"Show all available protocol implementations"`
	Verbose             int         `arg:"-v" help:"Verbose level: 1 or 2" placeholder:"LEVEL"`
	Quiet               bool        `help:"Suppress the logging output"`

	ConfigFile string `arg:"-c,--config-file" help:"YAML configuration file path" placeholder:"FILE"`
}

func main() {
	cliArgs := cli{}
	cliParser := arg.MustParse(&cliArgs)

	if cliArgs.ListImplementations != nil {
		listImplementations()
		return
	}

	// Setting up the logger
	// Initialize the stdlib logging as well to properly capture logs from other libraries
	slogOpts := &slog.HandlerOptions{AddSource: false, Level: slog.LevelInfo}
	switch cliArgs.Verbose {
	case 0:
		slogOpts.Level = slog.LevelInfo
	case 1:
		slogOpts.Level = slog.LevelDebug
	case 2:
		slogOpts.Level = slog.LevelDebug
	default:
		cliParser.Fail("Invalid verbosity level, use 0, 1 or 2")
	}
	log.SetVerbose(cliArgs.Verbose)
	stdLog.SetOutput(os.Stderr)
	slogHandler := slog.NewTextHandler(os.Stderr, slogOpts)
	if cliArgs.Quiet {
		stdLog.SetOutput(io.Discard)
		slogHandler = slog.NewTextHandler(io.Discard, slogOpts)
	}
	slog.SetDefault(slog.New(slogHandler))
	log.SetQuietMode(cliArgs.Quiet)

	logger := log.GetLogger("")
	logger.Info("Logging to stderr", "level", chlog.GetLevel())
	mergedConfig, err := loadFullConfig(cliArgs)
	if err != nil {
		logger.Error("Cannot load configuration", "error", err)
		os.Exit(1)
	}
	if !cliArgs.Quiet && mergedConfig.Quiet {
		log.SetQuietMode(true)
		stdLog.SetOutput(io.Discard)
		slogHandler = slog.NewTextHandler(io.Discard, slogOpts)
		slog.SetDefault(slog.New(slogHandler))
	}

	switch {
	case cliArgs.CodeCmd != nil:
		err = cliCode(cliArgs.CodeCmd, mergedConfig)
	case cliArgs.ClientCmd != nil:
		err = cliClient(cliArgs.ClientCmd, mergedConfig)
	case cliArgs.InfraCmd != nil:
		err = cliInfra(cliArgs.InfraCmd, mergedConfig)
	case cliArgs.DiagramCmd != nil:
		err = cliDiagram(cliArgs.DiagramCmd, mergedConfig)
	case cliArgs.UICmd != nil:
		err = cliUI(cliArgs.UICmd, mergedConfig)
	case cliArgs.DocCmd != nil:
		err = doc.CliDoc(cliArgs.DocCmd, mergedConfig)
	default:
		cliParser.Fail("No subcommand specified. Try --help for more information")
		os.Exit(1)
	}

	if err != nil {
		var me types.MultilineError
		switch {
		case errors.Is(err, common2.ErrInterruptedByUser):
			logger.Debug("Interrupted by user")
			os.Exit(0)
		case errors.Is(err, common2.ErrInvalidCLIArgument):
			cliParser.WriteHelp(os.Stderr)
		case errors.Is(err, common2.ErrBadResult):
			logger.Error(err.Error())
			os.Exit(2)
		case logger.GetLevel() <= chlog.DebugLevel && errors.As(err, &me):
			logger.Error(err.Error(), "details", me.ContentLines())
		}
		// Command unexpectedly failed and not finished
		logger.Error(err.Error())
		logger.Fatal("Cannot finish the command. Use -v=1 flag to enable debug output")
		os.Exit(1)
	}

	logger.Info("Done")
}

func loadFullConfig(cliArgs cli) (common2.ToolConfig, error) {
	logger := log.GetLogger("")
	builtinConfig, err := common2.LoadConfig(assets.AssetFS, common2.DefaultConfigFileName)
	if err != nil {
		return common2.ToolConfig{}, fmt.Errorf("load built-in config, this is a bug: %w", err)
	}

	fileName := cliArgs.ConfigFile
	if fileName == "" {
		if s, err := os.Stat("go-asyncapi.yaml"); err == nil && !s.IsDir() {
			fileName = "go-asyncapi.yaml"
		} else if s, err := os.Stat("go-asyncapi.yml"); err == nil && !s.IsDir() {
			fileName = "go-asyncapi.yml"
		}
	}

	var userConfig common2.ToolConfig
	if fileName != "" {
		logger.Debug("Loading user config", "file", fileName)
		if userConfig, err = common2.LoadConfig(os.DirFS(path.Dir(fileName)), path.Base(fileName)); err != nil {
			return common2.ToolConfig{}, fmt.Errorf("load config file %q: %w", fileName, err)
		}
	} else {
		logger.Debug("No user config, using only built-in defaults")
	}

	res := common2.MergeConfig(builtinConfig, userConfig)
	res.Quiet = common2.Coalesce(cliArgs.Quiet, res.Quiet)
	return res, err
}
