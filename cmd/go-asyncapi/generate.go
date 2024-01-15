package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/amqp"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/http"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/kafka"
	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/linker"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/bdragon300/go-asyncapi/internal/writer"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
)

type GenerateCmd struct {
	Pub            *generatePubSubArgs         `arg:"subcommand:pub" help:"Generate only the publisher code"`
	Sub            *generatePubSubArgs         `arg:"subcommand:sub" help:"Generate only the subscriber code"`
	PubSub         *generatePubSubArgs         `arg:"subcommand:pubsub" help:"Generate both publisher and subscriber code"`
	Implementation *generateImplementationArgs `arg:"subcommand:implementation" help:"Generate the implementation code only"`

	TargetDir string `arg:"-t,--target-dir" default:"./asyncapi" help:"Directory to save the generated code" placeholder:"DIR"`
	ImplDir   string `arg:"--impl-dir" default:"impl" help:"Directory to save implementations inside the target dir" placeholder:"DIR"`
}

type generatePubSubArgs struct {
	Spec string `arg:"required,positional" help:"AsyncAPI specification file path or url" placeholder:"PATH"`

	ProjectModule string `arg:"-M,--project-module" help:"Project module name to use [default: extracted from go.mod file in the current working directory]" placeholder:"MODULE"`
	TargetPackage string `arg:"-T,--target-package" help:"Package for generated code [default: {target-dir-name}]" placeholder:"PACKAGE"`
	PackageScope  string `arg:"--package-scope" default:"type" help:"How to split up the generated code on packages. Possible values: type, all" placeholder:"SCOPE"`
	FileScope     string `arg:"--file-scope" default:"name" help:"How to split up the generated code on files inside packages. Possible values: name, type" placeholder:"SCOPE"`
	generateObjectSelectionOpts
	ImplementationsOpts
	ExternalRefs bool `arg:"--external-refs" help:"Allow fetching specs from external $ref URLs"`

	RuntimeModule string `arg:"--runtime-module" default:"github.com/bdragon300/go-asyncapi/run" help:"Runtime module name" placeholder:"MODULE"`
}

// TODO: below there are new args to implement
type generateImplementationArgs struct {
	Protocol string `arg:"required,positional" help:"Protocol name to generate"`
	Name     string `arg:"required,positional" help:"Implementation name to generate"`
}

type ImplementationsOpts struct {
	Kafka string `arg:"--kafka-impl" default:"franz-go" help:"Implementation for Kafka ('no' to disable)" placeholder:"NAME"`
	AMQP  string `arg:"--amqp-impl" default:"amqp091-go" help:"Implementation for AMQP ('no' to disable)" placeholder:"NAME"`
	HTTP  string `arg:"--http-impl" default:"nethttp" help:"Implementation for HTTP ('no' to disable)" placeholder:"NAME"`
}

type generateObjectSelectionOpts struct {
	OnlyChannels        bool   `arg:"--only-channels" help:"Generate channels only"`
	ChannelsRe          string `arg:"--channels-re" help:"Generate only the channels that component key matches the regex" placeholder:"REGEX"`
	IgnoreChannelsRe    string `arg:"--ignore-channels-re" help:"Ignore the channels that component key matches the regex" placeholder:"REGEX"`
	IgnoreChannels      bool   `arg:"--ignore-channels" help:"Ignore all channels to be generated"`
	ReuseChannelsModule string `arg:"--reuse-channels-module" help:"Module name with channels code generated before to reuse now" placeholder:"MODULE"`

	OnlyMessages        bool   `arg:"--only-messages" help:"Generate messages only"`
	MessagesRe          string `arg:"--messages-re" help:"Generate only the messages that component key matches the regex" placeholder:"REGEX"`
	IgnoreMessagesRe    string `arg:"--ignore-messages-re" help:"Ignore the messages that component key matches the regex" placeholder:"REGEX"`
	IgnoreMessages      bool   `arg:"--ignore-messages" help:"Ignore all messages to be generated"`
	ReuseMessagesModule string `arg:"--reuse-messages-module" help:"Module name with messages code generated before to reuse now" placeholder:"MODULE"`

	OnlyModels        bool   `arg:"--only-models" help:"Generate models only"`
	ModelsRe          string `arg:"--models-re" help:"Generate only the models that component key matches the regex" placeholder:"REGEX"`
	IgnoreModelsRe    string `arg:"--ignore-models-re" help:"Ignore the models that component key matches the regex" placeholder:"REGEX"`
	IgnoreModels      bool   `arg:"--ignore-models" help:"Ignore all models to be generated"`
	ReuseModelsModule string `arg:"--reuse-models-module" help:"Module name with models code generated before to reuse now" placeholder:"MODULE"`

	OnlyServers        bool   `arg:"--only-servers" help:"Generate servers only"`
	ServersRe          string `arg:"--servers-re" help:"Generate only the servers that component key matches the regex" placeholder:"REGEX"`
	IgnoreServersRe    string `arg:"--ignore-servers-re" help:"Ignore the servers that component key matches the regex" placeholder:"REGEX"`
	IgnoreServers      bool   `arg:"--ignore-servers" help:"Ignore all servers to be generated"`
	ReuseServersModule string `arg:"--reuse-servers-module" help:"Module name with servers code generated before to reuse now" placeholder:"MODULE"`

	NoImplementations bool `arg:"--no-implementations" help:"Do not generate any protocol implementation"`
	NoEncoding        bool `arg:"--no-encoding" help:"Do not generate encoders/decoders code"`
}

func generate(cmd *GenerateCmd) error {
	if cmd.Implementation != nil {
		return generateImplementation(cmd)
	}

	isPub, isSub, pubSubOpts := getPubSubVariant(cmd) // TODO: pass pub/sub bool flags to the generation functions
	targetPkg, _ := lo.Coalesce(pubSubOpts.TargetPackage, path.Base(cmd.TargetDir))
	mainLogger.Debugf("Target package name is %s", targetPkg)
	if !isSub && !isPub {
		return fmt.Errorf("%w: no publisher or subscriber set to generate", ErrWrongCliArgs)
	}
	compileOpts, err := getCompileOpts(*pubSubOpts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	renderOpts, err := getRenderOpts(*pubSubOpts, cmd.TargetDir, targetPkg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	implDir := path.Join(cmd.TargetDir, cmd.ImplDir)
	mainLogger.Debugf("Target implementations directory is %s", implDir)

	// Compilation
	specID, _ := utils.SplitSpecPath(pubSubOpts.Spec)
	modules, err := generationCompile(specID, compileOpts)
	if err != nil {
		return err
	}
	objSources := lo.MapValues(modules, func(value *compiler.Module, _ string) linker.ObjectSource { return value })
	mainModule := modules[specID]

	// Linking
	if err = generationLinking(objSources); err != nil {
		return err
	}

	// Rendering
	protoRenderers := lo.MapValues(protocolBuilders(), func(value asyncapi.ProtocolBuilder, _ string) common.ProtocolRenderer { return value })
	files, err := writer.RenderPackages(mainModule, protoRenderers, renderOpts)
	if err != nil {
		return fmt.Errorf("schema render: %w", err)
	}

	// Writing
	if err = writer.WriteToFiles(files, cmd.TargetDir); err != nil {
		return fmt.Errorf("writing code to files: %w", err)
	}

	// Rendering the selected implementations
	if !pubSubOpts.NoImplementations {
		selectedImpls := getImplementationsOpts(pubSubOpts.ImplementationsOpts)
		if err = generationWriteImplementations(selectedImpls, mainModule.Protocols(), implDir); err != nil {
			return err
		}
	}

	mainLogger.Info("Finished")
	return nil
}

func generateImplementation(cmd *GenerateCmd) error {
	implDir := path.Join(cmd.TargetDir, cmd.ImplDir)
	mainLogger.Debugf("Target implementations directory is %s", implDir)
	proto := cmd.Implementation.Protocol
	name := cmd.Implementation.Name
	if err := generationWriteImplementations(map[string]string{proto: name}, []string{proto}, implDir); err != nil {
		return err
	}

	mainLogger.Info("Finished")
	return nil
}

func generationCompile(specID string, compileOpts common.CompileOpts) (map[string]*compiler.Module, error) {
	logger := types.NewLogger("Compilation 🔨")
	asyncapi.ProtocolBuilders = protocolBuilders()
	compileQueue := []string{specID}             // Queue of specIDs to compile
	modules := make(map[string]*compiler.Module) // Compilers by spec id
	for len(compileQueue) > 0 {
		specID = compileQueue[0]          // Pop from the queue
		compileQueue = compileQueue[1:]   //
		if _, ok := modules[specID]; ok { // Skip if specID has been already modules
			continue
		}

		logger.Info("Run compilation", "specID", specID)
		module := compiler.NewModule(specID)
		modules[specID] = module

		if !compileOpts.EnableExternalRefs && utils.IsRemoteSpecID(specID) {
			return nil, fmt.Errorf("external refs are forbidden by default for security reasons, use --external-refs flag to allow fetching specs from external resources")
		}
		logger.Debug("Loading a spec", "specID", specID)
		if err := module.Load(); err != nil {
			return nil, fmt.Errorf("load the spec: %w", err)
		}
		logger.Debug("Compilation a spec", "specID", specID)
		if err := module.Compile(common.NewCompileContext(specID, compileOpts)); err != nil {
			return nil, fmt.Errorf("compilation the spec: %w", err)
		}
		logger.Debugf("Compiler stats: %s", module.Stats())
		compileQueue = lo.Flatten([][]string{compileQueue, module.RemoteSpecIDs()}) // Extend queue with remote specIDs
	}

	logger.Info("Compilation completed", "files", len(modules))
	return modules, nil
}

func generationLinking(objSources map[string]linker.ObjectSource) error {
	logger := types.NewLogger("Linking 🔗")
	logger.Info("Run linking")
	// Linking refs
	linker.AssignRefs(objSources)
	danglingRefs := linker.DanglingRefs(objSources)
	logger.Debugf("Linker stats: %s", linker.Stats(objSources))
	if len(danglingRefs) > 0 {
		logger.Error("Some refs remain dangling", "refs", danglingRefs)
		return fmt.Errorf("cannot finish linking")
	}

	// Linking list promises
	logger.Debug("Run linking the list promises")
	linker.AssignListPromises(objSources)
	danglingPromises := linker.DanglingPromisesCount(objSources)
	logger.Debugf("Linker stats: %s", linker.Stats(objSources))
	if danglingPromises > 0 {
		logger.Error("Cannot assign internal list promises", "promises", danglingPromises)
		return fmt.Errorf("cannot finish linking")
	}

	refsCount := lo.SumBy(lo.Values(objSources), func(item linker.ObjectSource) int {
		return lo.CountBy(item.Promises(), func(p common.ObjectPromise) bool {
			return p.Origin() == common.PromiseOriginUser
		})
	})
	logger.Info("Linking completed", "refs", refsCount)
	return nil
}

func generationWriteImplementations(selectedImpls map[string]string, protocols []string, implDir string) error {
	logger := types.NewLogger("Writing 📝")
	logger.Info("Writing implementations")
	implManifest, err := getImplementationsManifest()
	if err != nil {
		panic(err.Error())
	}

	var totalBytes int
	var writtenProtocols []string
	for _, p := range protocols {
		implName := selectedImpls[p]
		if implName == "no" || implName == "" {
			logger.Debug("Implementation has been unselected", "protocol", p)
			continue
		}
		writtenProtocols = append(writtenProtocols, p)
		if _, ok := implManifest[p][implName]; !ok {
			return fmt.Errorf("unknown implementation %q for %q protocol, use list-implementations command to see possible values", implName, p)
		}
		logger.Debug("Writing implementation", "protocol", p, "name", implName)
		n, err := writer.WriteImplementation(implManifest[p][implName].Dir, path.Join(implDir, p))
		if err != nil {
			return fmt.Errorf("implementation rendering for protocol %q: %w", p, err)
		}
		totalBytes += n
	}
	logger.Debugf(
		"Implementations writer stats: total bytes: %d, protocols: %s",
		totalBytes, strings.Join(writtenProtocols, ","),
	)

	logger.Info("Writing implementations completed", "count", len(writtenProtocols))
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

func getImplementationsOpts(opts ImplementationsOpts) map[string]string {
	return map[string]string{
		amqp.Builder.ProtocolName():  opts.AMQP,
		http.Builder.ProtocolName():  opts.HTTP,
		kafka.Builder.ProtocolName(): opts.Kafka,
	}
}

func getPubSubVariant(cmd *GenerateCmd) (pub bool, sub bool, variant *generatePubSubArgs) {
	switch {
	case cmd.PubSub != nil:
		return true, true, cmd.PubSub
	case cmd.Pub != nil:
		return true, false, cmd.Pub
	case cmd.Sub != nil:
		return false, true, cmd.Sub
	}
	return
}

func getCompileOpts(opts generatePubSubArgs) (common.CompileOpts, error) {
	var err error
	res := common.CompileOpts{
		ReusePackages:      nil,
		NoEncodingPackage:  opts.NoEncoding,
		EnableExternalRefs: opts.ExternalRefs,
		RuntimeModule:      opts.RuntimeModule,
	}

	includeAll := !opts.OnlyChannels && !opts.OnlyMessages && !opts.OnlyModels && !opts.OnlyServers
	f := func(only, ignore bool, re, ire string) (r common.ObjectSelectionOpts, e error) {
		r.Enable = (includeAll || only) && !ignore
		if re != "" {
			if r.IncludeRegex, e = regexp.Compile(re); e != nil {
				return
			}
		}
		if ire != "" {
			if r.ExcludeRegex, e = regexp.Compile(ire); e != nil {
				return
			}
		}
		return
	}

	if res.ChannelsSelection, err = f(opts.OnlyChannels, opts.IgnoreChannels, opts.ChannelsRe, opts.IgnoreChannelsRe); err != nil {
		return res, err
	}
	// TODO: make enum for channels
	if opts.ReuseChannelsModule != "" {
		res.ReusePackages["channels"] = opts.ReuseChannelsModule
	}
	if res.MessagesSelection, err = f(opts.OnlyMessages, opts.IgnoreMessages, opts.MessagesRe, opts.IgnoreMessagesRe); err != nil {
		return res, err
	}
	if opts.ReuseMessagesModule != "" {
		res.ReusePackages["messages"] = opts.ReuseMessagesModule
	}
	if res.ModelsSelection, err = f(opts.OnlyModels, opts.IgnoreModels, opts.ModelsRe, opts.IgnoreModelsRe); err != nil {
		return res, err
	}
	if opts.ReuseModelsModule != "" {
		res.ReusePackages["models"] = opts.ReuseModelsModule
	}
	if res.ServersSelection, err = f(opts.OnlyServers, opts.IgnoreServers, opts.ServersRe, opts.IgnoreServersRe); err != nil {
		return res, err
	}
	if opts.ReuseServersModule != "" {
		res.ReusePackages["servers"] = opts.ReuseServersModule
	}

	return res, nil
}

func getRenderOpts(opts generatePubSubArgs, targetDir, targetPkg string) (common.RenderOpts, error) {
	res := common.RenderOpts{
		RuntimeModule: opts.RuntimeModule,
		TargetPackage: targetPkg,
		TargetDir:     targetDir,
	}

	importBase := opts.ProjectModule
	if importBase == "" {
		b, err := getImportBase()
		if err != nil {
			return res, fmt.Errorf("determine project's module (use -M arg to override): %w", err)
		}
		importBase = b
	}
	importBase = path.Join(importBase, targetPkg)
	mainLogger.Debugf("Target import base is %s", importBase)
	res.ImportBase = importBase

	switch opts.PackageScope {
	case "all":
		res.PackageScope = common.PackageScopeAll
	case "type":
		res.PackageScope = common.PackageScopeType
	default:
		return res, fmt.Errorf("%w: unknown package scope: %q", ErrWrongCliArgs, opts.PackageScope)
	}

	switch opts.FileScope {
	case "type":
		res.FileScope = common.FileScopeType
	case "name":
		res.FileScope = common.FileScopeName
	default:
		return res, fmt.Errorf("%w: unknown file scope: %q", ErrWrongCliArgs, opts.FileScope)
	}

	return res, nil
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
