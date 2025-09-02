package main

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/bdragon300/go-asyncapi/assets"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"

	stdLog "log"

	chlog "github.com/charmbracelet/log"

	"github.com/alexflint/go-arg"
)

const (
	defaultConfigFileName                   = "default_config.yaml"
	defaultMainTemplateName                 = "main.tmpl"
	defaultSubprocessLocatorShutdownTimeout = 3 * time.Second
)

var ErrWrongCliArgs = errors.New("cli args")

type cli struct {
	CodeCmd             *CodeCmd    `arg:"subcommand:code" help:"Generate the code"`
	ClientCmd           *ClientCmd  `arg:"subcommand:client" help:"Build the client executable (requires Go toolchain installed)"`
	InfraCmd            *InfraCmd   `arg:"subcommand:infra" help:"Generate the infrastructure setup files"`
	DiagramCmd          *DiagramCmd `arg:"subcommand:diagram" help:"Generate the architecture diagram"`
	ListImplementations *struct{}   `arg:"subcommand:list-implementations" help:"Show all available protocol implementations"`
	Verbose             int         `arg:"-v" help:"Verbose output: 1 (debug), 2 (trace)" placeholder:"LEVEL"`
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
		chlog.SetLevel(chlog.InfoLevel)
		slogOpts.Level = slog.LevelInfo
	case 1:
		chlog.SetLevel(chlog.DebugLevel)
		slogOpts.Level = slog.LevelDebug
	case 2:
		chlog.SetLevel(log.TraceLevel)
		slogOpts.Level = slog.LevelDebug
	default:
		cliParser.Fail("Invalid verbosity level, use 0, 1 or 2")
	}
	chlog.SetReportTimestamp(false)
	chlog.SetOutput(os.Stderr)
	stdLog.SetOutput(os.Stderr)
	slogHandler := slog.NewTextHandler(os.Stderr, slogOpts)
	if cliArgs.Quiet {
		chlog.SetOutput(io.Discard)
		stdLog.SetOutput(io.Discard)
		slogHandler = slog.NewTextHandler(io.Discard, slogOpts)
	}
	slog.SetDefault(slog.New(slogHandler))

	logger := log.GetLogger("")
	logger.Info("Logging to stderr", "level", chlog.GetLevel())
	builtinConfig, err := loadConfig(assets.AssetFS, defaultConfigFileName)
	if err != nil {
		logger.Error("Cannot load built-in config", "error", err)
		os.Exit(1)
	}
	var userConfig toolConfig
	if cliArgs.ConfigFile != "" {
		log.GetLogger("").Debug("Load config", "file", cliArgs.ConfigFile)
		userConfig, err = loadConfig(os.DirFS(path.Dir(cliArgs.ConfigFile)), path.Base(cliArgs.ConfigFile))
		if err != nil {
			logger.Error("Cannot load user config", "error", err)
			os.Exit(1)
		}
	}
	mergedConfig := mergeConfig(builtinConfig, userConfig)

	switch {
	case cliArgs.CodeCmd != nil:
		err = cliCode(cliArgs.CodeCmd, mergedConfig)
	case cliArgs.ClientCmd != nil:
		err = cliClient(cliArgs.ClientCmd, mergedConfig)
	case cliArgs.InfraCmd != nil:
		err = cliInfra(cliArgs.InfraCmd, mergedConfig)
	case cliArgs.DiagramCmd != nil:
		err = cliDiagram(cliArgs.DiagramCmd, mergedConfig)
	default:
		cliParser.Fail("No command specified. Try --help for more information")
		os.Exit(1)
	}

	if err != nil {
		var me types.MultilineError
		switch {
		case errors.Is(err, ErrWrongCliArgs):
			cliParser.WriteHelp(os.Stderr)
		case chlog.GetLevel() <= chlog.DebugLevel && errors.As(err, &me):
			chlog.Error(err.Error(), "details", me.ContentLines())
		}
		chlog.Error(err.Error())
		chlog.Fatal("Cannot finish the command. Use -v=1 flag to enable debug output")
		os.Exit(1)
	}

	chlog.Info("Done")
}
