package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi/amqp"
	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi/http"
	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi/kafka"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"github.com/charmbracelet/log"

	"github.com/bdragon300/asyncapi-codegen-go/implementations"

	"github.com/alexflint/go-arg"
	"github.com/bdragon300/asyncapi-codegen-go/internal/writer"
)

type GenerateCmd struct {
	Spec          string `arg:"required,positional" help:"AsyncAPI specification file path or url" placeholder:"PATH"`
	TargetDir     string `arg:"-t,--target-dir" default:"./asyncapi" help:"Directory to save the generated code" placeholder:"DIR"`
	ImplDir       string `arg:"--impl-dir" help:"Directory where protocol implementations will be placed. By default it is {target-dir}/impl" placeholder:"DIR"`
	ProjectModule string `arg:"-M,--project-module" help:"Project module name to use. By default it is extracted from go.mod file in the current working directory" placeholder:"MODULE"`
	TargetPackage string `arg:"-T,--target-package" help:"Package for generated code. By default it is equal to the target directory" placeholder:"NAME"`
	ImplementationsOpts
}

type ImplementationsOpts struct {
	Kafka string `arg:"--kafka-impl" default:"franz-go" help:"Implementation for Kafka protocol or 'no' to disable implementation" placeholder:"NAME"`
	AMQP  string `arg:"--amqp-impl" default:"amqp091-go" help:"Implementation for AMQP protocol or 'no' to disable implementation" placeholder:"NAME"`
	HTTP  string `arg:"--http-impl" default:"stdhttp" help:"Implementation for HTTP protocol or 'no' to disable implementation" placeholder:"NAME"`
}

type cli struct {
	GenerateCmd         *GenerateCmd `arg:"subcommand:generate" help:"Generate the code based on AsyncAPI specification"`
	ListImplementations *struct{}    `arg:"subcommand:list-implementations" help:"Show all available protocol implementations"`
	Verbose             int          `arg:"-v" help:"Logging verbosity: 0 default, 1 debug output, 2 more debug output" placeholder:"LEVEL"`
	Quiet               bool         `help:"Suppress the output"`
}

var logger *types.Logger

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
	logger = types.NewLogger("")

	cmd := cliArgs.GenerateCmd
	if err := generate(cmd); err != nil {
		var multilineErr writer.MultilineError
		if log.GetLevel() <= log.DebugLevel && errors.As(err, &multilineErr) {
			log.Error(err.Error(), "details", multilineErr.RestLines())
		}

		log.Error(err.Error())
		log.Fatal("Cannot finish the generation. Use -v=1 flag to enable debug output")
	}
}

func listImplementations() {
	manifest, err := getImplementationsManifest()
	if err != nil {
		panic(err.Error())
	}
	for proto, implInfo := range manifest {
		_, _ = os.Stdout.WriteString(proto + ":\n")
		for implName, info := range implInfo {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("* %s (%s)\n", implName, info.URL))
		}
		_, _ = os.Stdout.WriteString("\n")
	}
}

func getSelectedImplementations(opts ImplementationsOpts) map[string]string {
	return map[string]string{
		amqp.Builder.ProtocolName():  opts.AMQP,
		http.Builder.ProtocolName():  opts.HTTP,
		kafka.Builder.ProtocolName(): opts.Kafka,
	}
}

func protocolBuilders() map[string]asyncapi.ProtocolBuilder {
	return map[string]asyncapi.ProtocolBuilder{
		amqp.Builder.ProtocolName():  amqp.Builder,
		http.Builder.ProtocolName():  http.Builder,
		kafka.Builder.ProtocolName(): kafka.Builder,
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
