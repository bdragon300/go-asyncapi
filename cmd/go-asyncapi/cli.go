package main

import (
	"errors"
	"io"
	"os"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/charmbracelet/log"

	"github.com/alexflint/go-arg"
	"github.com/bdragon300/go-asyncapi/internal/writer"
)

var ErrWrongCliArgs = errors.New("cli args")

type cli struct {
	GenerateCmd         *GenerateCmd `arg:"subcommand:generate" help:"Generate the code based on AsyncAPI specification"`
	ListImplementations *struct{}    `arg:"subcommand:list-implementations" help:"Show all available protocol implementations"`
	Verbose             int          `arg:"-v" help:"Logging verbosity: 0 default, 1 debug output, 2 more debug output" placeholder:"LEVEL"`
	Quiet               bool         `help:"Suppress the output"`
}

var mainLogger *types.Logger

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

	// Setting up the logger
	switch cliArgs.Verbose {
	case 0:
		log.SetLevel(log.InfoLevel)
	case 1:
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(types.TraceLevel)
	}
	if cliArgs.Quiet {
		log.SetOutput(io.Discard)
	}
	log.SetReportTimestamp(false)
	mainLogger = types.NewLogger("") // FIXME: configure logger in single place

	cmd := cliArgs.GenerateCmd
	if err := generate(cmd); err != nil {
		var multilineErr writer.MultilineError
		switch {
		case errors.Is(err, ErrWrongCliArgs):
			cliParser.WriteHelp(os.Stderr)
		case log.GetLevel() <= log.DebugLevel && errors.As(err, &multilineErr):
			log.Error(err.Error(), "details", multilineErr.RestLines())
		}

		log.Error(err.Error())
		log.Fatal("Cannot finish the generation. Use -v=1 flag to enable debug output")
	}
}
