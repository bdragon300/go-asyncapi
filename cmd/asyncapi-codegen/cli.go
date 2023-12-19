package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/compiler"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"

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

	registerProtocols()

	cmd := cliArgs.GenerateCmd
	if err := generate(cmd); err != nil {
		var multilineErr writer.MultilineError
		if log.GetLevel() == types.TraceLevel && errors.As(err, &multilineErr) {
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
	importBase := cmd.ProjectModule
	if importBase == "" {
		b, err := getImportBase()
		if err != nil {
			return fmt.Errorf("extraction module name from go.mod (you can specify it by -M argument): %w", err)
		}
		importBase = b
	}
	targetPkg, _ := lo.Coalesce(cmd.TargetPackage, cmd.TargetDir)
	logger.Debugf("Target package name is %s", targetPkg)
	importBase = path.Join(importBase, targetPkg)
	logger.Debugf("Target import base is %s", importBase)
	implDir, _ := lo.Coalesce(cmd.ImplDir, path.Join(cmd.TargetDir, "impl"))
	logger.Debugf("Target implementations directory is %s", implDir)

	// Compilation
	specID, _ := utils.SplitSpecPath(cmd.Spec)
	firstSpecID := specID
	specLinker := linker.NewSpecLinker()
	compileQueue := []string{specID}                // Queue of specIDs to compile
	compiled := make(map[string]*compiler.Compiler) // Compilers by spec id
	for len(compileQueue) > 0 {
		specID = compileQueue[0]           // Pop from the queue
		compileQueue = compileQueue[1:]    //
		if _, ok := compiled[specID]; ok { // Skip if specID has been already compiled
			continue
		}

		logger.Info("Run compilation", "path", specID)
		comp := compiler.NewCompiler(specID)
		compiled[specID] = comp

		logger.Debug("Loading spec path", "path", specID)
		if err := comp.Load(); err != nil {
			return fmt.Errorf("load the spec: %w", err)
		}
		logger.Debug("Compilation a loaded file", "path", specID)
		if err := comp.Compile(common.NewCompileContext(specLinker, specID)); err != nil {
			return fmt.Errorf("compilation the spec: %w", err)
		}
		logger.Debugf("Compiler stats: %s", comp.Stats())
		compileQueue = lo.Flatten([][]string{compileQueue, comp.RemoteSpecIDs()}) // Extend queue with remote specIDs
	}
	logger.Info("Compilation completed", "files", len(compiled))

	comps := lo.MapValues(compiled, func(value *compiler.Compiler, _ string) linker.ObjectSource { return value })

	// Linking: refs
	logger.Info("Run linking", "path", specID)
	specLinker.ProcessRefs(comps)
	danglingRefs := specLinker.DanglingRefs()
	logger.Debugf("Linker stats: %s", specLinker.Stats())
	if len(danglingRefs) > 0 {
		logger.Error("Some refs remain dangling", "refs", danglingRefs)
		return fmt.Errorf("cannot finish linking")
	}

	// Linking: list promises
	logger.Debug("Run linking the list promises")
	specLinker.ProcessListPromises(comps)
	danglingPromises := specLinker.DanglingPromisesCount()
	logger.Debugf("Linker stats: %s", specLinker.Stats())
	if danglingPromises > 0 {
		logger.Error("Cannot assign internal list promises", "promises", danglingPromises)
		return fmt.Errorf("cannot finish linking")
	}
	logger.Info("Linking completed", "files", len(compiled))

	firstComp := compiled[firstSpecID]

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
		"Implementations writer stats: total bytes: %d, protocols: %s",
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
