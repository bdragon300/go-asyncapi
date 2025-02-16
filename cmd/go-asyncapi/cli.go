package main

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/assets"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"io"
	"os"
	"path"

	chlog "github.com/charmbracelet/log"

	"github.com/alexflint/go-arg"
)

const defaultConfigFileName = "default_config.yaml"

var ErrWrongCliArgs = errors.New("cli args")

type cli struct {
	GenerateCmd         *GenerateCmd `arg:"subcommand:generate" help:"Generate the code based on AsyncAPI document"`
	ClientCmd		   *ClientCmd   `arg:"subcommand:client" help:"Build the client executable based on AsyncAPI document (requires Go toolchain installed)"`
	InfraCmd			*InfraCmd    `arg:"subcommand:infra" help:"Generate the infrastructure code based on AsyncAPI document"`
	ListImplementations *struct{}    `arg:"subcommand:list-implementations" help:"Show all available protocol implementations"`
	Verbose             int          `arg:"-v" help:"Logging verbosity: 0 default, 1 debug output, 2 more debug output" placeholder:"LEVEL"`
	Quiet               bool         `help:"Suppress the logging output"`

	ConfigFile string `arg:"-c,--config-file" help:"YAML configuration file path" placeholder:"PATH"`
}

func main() {
	cliArgs := cli{}
	cliParser := arg.MustParse(&cliArgs)

	if cliArgs.ListImplementations != nil {
		listImplementations()
		return
	}

	// Setting up the logger
	switch cliArgs.Verbose {
	case 0:
		chlog.SetLevel(chlog.InfoLevel)
	case 1:
		chlog.SetLevel(chlog.DebugLevel)
	default:
		chlog.SetLevel(log.TraceLevel)
	}
	if cliArgs.Quiet {
		chlog.SetOutput(io.Discard)
	}
	chlog.SetReportTimestamp(false)

	logger := log.GetLogger("")
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
	case cliArgs.GenerateCmd != nil:
		err = cliGenerate(cliArgs.GenerateCmd, mergedConfig)
	case cliArgs.ClientCmd != nil:
		err = cliClient(cliArgs.ClientCmd, mergedConfig)
	case cliArgs.InfraCmd != nil:
		err = cliInfra(cliArgs.InfraCmd, mergedConfig)
	default:
		cliParser.Fail("No command specified. Try --help for more information")
		os.Exit(1)
	}

	if err != nil {
		var multilineErr types.ErrorWithContent
		switch {
		case errors.Is(err, ErrWrongCliArgs):
			cliParser.WriteHelp(os.Stderr)
		case chlog.GetLevel() <= chlog.DebugLevel && errors.As(err, &multilineErr):
			chlog.Error(err.Error(), "details", multilineErr.ContentLines())
		}
		chlog.Error(err.Error())
		chlog.Fatal("Cannot finish the command. Use -v=1 flag to enable debug output")
		os.Exit(1)
	} else {
		chlog.Info("Done")
	}
}
