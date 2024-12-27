package main

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"io"
	"os"

	chlog "github.com/charmbracelet/log"

	"github.com/alexflint/go-arg"
)

var ErrWrongCliArgs = errors.New("cli args")

type cli struct {
	GenerateCmd         *GenerateCmd `arg:"subcommand:generate" help:"Generate the code based on AsyncAPI specification"`
	ListImplementations *struct{}    `arg:"subcommand:list-implementations" help:"Show all available protocol implementations"`
	Verbose             int          `arg:"-v" help:"Logging verbosity: 0 default, 1 debug output, 2 more debug output" placeholder:"LEVEL"`
	Quiet               bool         `help:"Suppress the output"`
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

	cmd := cliArgs.GenerateCmd
	if err := generate(cmd); err != nil {
		var multilineErr types.ErrorWithContent
		switch {
		case errors.Is(err, ErrWrongCliArgs):
			cliParser.WriteHelp(os.Stderr)
		case chlog.GetLevel() <= chlog.DebugLevel && errors.As(err, &multilineErr):
			chlog.Error(err.Error(), "details", multilineErr.ContentLines())
		}

		chlog.Error(err.Error())
		chlog.Fatal("Cannot finish the generation. Use -v=1 flag to enable debug output")
	}
}
