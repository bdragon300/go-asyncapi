package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/factory"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols/http"

	"github.com/charmbracelet/log"

	"github.com/bdragon300/asyncapi-codegen-go/implementations"

	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols/amqp"

	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols/kafka"
	"github.com/bdragon300/asyncapi-codegen-go/internal/writer"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"

	"github.com/bdragon300/asyncapi-codegen-go/internal/linker"

	"github.com/alexflint/go-arg"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
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
	Verbose             bool         `arg:"-v,--verbose" help:"Verbose output"`
	Trace               bool         `arg:"--trace" help:"Trace output"` // TODO: --quiet
}

var logger = common.NewLogger("")

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
	log.SetLevel(log.InfoLevel)
	if cliArgs.Verbose {
		log.SetLevel(log.DebugLevel)
	}
	if cliArgs.Trace {
		log.SetLevel(common.TraceLevel)
	}
	log.SetReportTimestamp(false)
	logger.SetReportTimestamp(false)

	registerProtocols()

	cmd := cliArgs.GenerateCmd
	if err := generate(cmd); err != nil {
		var multilineErr writer.MultilineError
		if log.GetLevel() == common.TraceLevel && errors.As(err, &multilineErr) {
			log.Error(err.Error(), "details", multilineErr.RestLines())
		}

		log.Error(err.Error())
		log.Fatal("Cannot finish the generation. Use --verbose or --trace flag to get more info")
	}
}

func registerProtocols() {
	kafka.Register()
	amqp.Register()
	http.Register()
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

func generate(cmd *GenerateCmd) error {
	var err error

	importBase := cmd.ProjectModule
	if importBase == "" {
		if importBase, err = getImportBase(); err != nil {
			return fmt.Errorf("extraction module name from go.mod (you can specify it by -M argument): %w", err)
		}
	}
	targetPkg, _ := lo.Coalesce(cmd.TargetPackage, cmd.TargetDir)
	logger.Debugf("Target package name is %s", targetPkg)
	importBase = path.Join(importBase, targetPkg)
	logger.Debugf("Target import base is %s", importBase)
	implDir, _ := lo.Coalesce(cmd.ImplDir, path.Join(cmd.TargetDir, "impl"))
	logger.Debugf("Target implementations directory is %s", implDir)

	// Compilation queue
	// Compilation spec -> linking -> push external refs to queue -> compilation next spec -> etc.
	specPath := cmd.Spec
	specLinker := linker.NewSpecLinker()
	var compilers []factory.Compiler
	for {
		logger.Info("Run compilation", "path", specPath)
		comp, err := factory.MakeCompiler(specPath)
		if err != nil {
			return fmt.Errorf("make a compiler: %w", err)
		}
		compilers = append(compilers, comp)
		logger.Debug("Loading spec path", "path", specPath)
		if err := comp.Load(); err != nil {
			return fmt.Errorf("load the spec: %w", err)
		}
		logger.Debug("Compilation a loaded file", "path", specPath)
		if err := comp.Compile(specLinker); err != nil {
			return fmt.Errorf("compilation the spec: %w", err)
		}
		logger.Debugf("Compiler stats: %s", comp.Stats())

		// Linking
		logger.Info("Run linking", "path", specPath)
		if err = specLinker.Process(comp); err != nil {
			return fmt.Errorf("schema linking error: %w", err)
		}
		ref, ok := specLinker.PopExternalQuery()
		if !ok {
			break // Queue is empty
		}
		specPath = ref.Ref()
	}
	logger.Debugf("Linker stats: %s", specLinker.Stats())
	danglingQueries := specLinker.DanglingQueries()
	if len(danglingQueries) > 0 {
		logger.Error("Cannot assign some refs", "links", danglingQueries)
	}
	logger.Info("Compilation/linking completed", "files", len(compilers))

	firstComp := compilers[0]

	// Rendering
	logger.Info("Run rendering")
	files, err := writer.RenderPackages(firstComp, importBase, cmd.TargetDir)
	if err != nil {
		return fmt.Errorf("schema render: %w", err)
	}

	// Writing
	logger.Info("Run writing")
	if err = writer.WriteToFiles(files, cmd.TargetDir); err != nil {
		return fmt.Errorf("writing code to files: %w", err)
	}

	// Rendering implementations
	logger.Info("Run writing selected implementations")
	implManifest, err := getImplementationsManifest()
	if err != nil {
		panic(err.Error())
	}
	selectedImpls := getSelectedImplementations(cmd.ImplementationsOpts)
	var total int
	for _, p := range firstComp.Protocols() {
		implName := selectedImpls[p]
		if implName == "no" || implName == "" {
			logger.Debug("Implementation has been unselected", "protocol", p)
			continue
		}
		if _, ok := implManifest[p][implName]; !ok {
			return fmt.Errorf("unknown implementation %q for %q protocol, use list-implementations command to see possible values", implName, p)
		}
		logger.Debug("Writing implementation", "protocol", p, "name", implName)
		n, err := writer.WriteImplementation(implManifest[p][implName].Dir, path.Join(implDir, p))
		if err != nil {
			return fmt.Errorf("implementation rendering for protocol %q: %w", p, err)
		}
		total += n
	}
	logger.WithPrefix("Writing 📝").Debugf(
		"Implementations writer stats: total bytes: %d, protocols: %q",
		total, strings.Join(firstComp.Protocols(), ","),
	)

	logger.Info("Finished")
	return nil
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
		kafka.ProtoName: opts.Kafka,
		amqp.ProtoName:  opts.AMQP,
		http.ProtoName:  opts.HTTP,
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
