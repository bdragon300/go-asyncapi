package main

import (
	"encoding/json"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/specurl"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/tcp"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/udp"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/ip"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/redis"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/ws"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/mqtt"

	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/amqp"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/http"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/kafka"
	stdHTTP "net/http"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/linker"
	"github.com/bdragon300/go-asyncapi/internal/writer"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
)

const (
	defaultConfigFileName = "default_config.yaml"
	defaultPackage 	  = "main"
	defaultTemplate 	  = "main.tmpl"
)

type GenerateCmd struct {
	Pub            *generatePubSubArgs         `arg:"subcommand:pub" help:"Generate only the publisher code"`
	Sub            *generatePubSubArgs         `arg:"subcommand:sub" help:"Generate only the subscriber code"`
	PubSub         *generatePubSubArgs         `arg:"subcommand:pubsub" help:"Generate both publisher and subscriber code"`
	Implementation *generateImplementationArgs `arg:"subcommand:implementation" help:"Generate the implementation code only"`

	TargetDir string `arg:"-t,--target-dir" help:"Directory to save the generated code" placeholder:"DIR"`
	ImplDir   string `arg:"--impl-dir" help:"Directory to save implementations inside the target dir" placeholder:"DIR"`
	ConfigFile string `arg:"-c,--config-file" help:"YAML configuration file path" placeholder:"PATH"`
}

type generatePubSubArgs struct {
	Spec string `arg:"required,positional" help:"AsyncAPI specification file path or url" placeholder:"PATH"`

	ProjectModule string `arg:"-M,--module" help:"Module path to use [default: extracted from go.mod file in the current working directory concatenated with target dir]" placeholder:"MODULE"`
	TemplateDir string `arg:"-T,--template-dir" help:"Directory with custom templates" placeholder:"DIR"`
	PreambleTemplate string `arg:"--preamble-template" help:"Custom preamble template name" placeholder:"NAME"`
	DisableFormatting bool `arg:"--disable-formatting" help:"Disable code formatting"`
	ImplementationsOpts
	AllowRemoteRefs bool `arg:"--allow-remote-refs" help:"Allow fetching spec files from remote $ref URLs"`

	RuntimeModule       string        `arg:"--runtime-module" help:"Runtime module name" placeholder:"MODULE"`
	ResolverSearchDir   string        `arg:"--resolver-search-dir" help:"Directory to search the local spec files for [default: current working directory]" placeholder:"PATH"`
	ResolverTimeout time.Duration `arg:"--resolver-timeout" help:"Timeout for resolver to resolve a spec file" placeholder:"DURATION"`
	ResolverCommand string        `arg:"--resolver-command" help:"Custom resolver executable to use instead of built-in resolver" placeholder:"PATH"`
}

type generateImplementationArgs struct {
	Protocol string `arg:"required,positional" help:"Protocol name to generate"`
	Name     string `arg:"required,positional" help:"Implementation name to generate"`
}

type ImplementationsOpts struct {
	NoImplementations bool `arg:"--no-implementations" help:"Do not generate any protocol implementation"`

	Kafka string `arg:"--kafka-impl" help:"Implementation for Kafka ('none' to disable)" placeholder:"NAME"`
	AMQP  string `arg:"--amqp-impl" help:"Implementation for AMQP ('none' to disable)" placeholder:"NAME"`
	HTTP  string `arg:"--http-impl" help:"Implementation for HTTP ('none' to disable)" placeholder:"NAME"`
	MQTT  string `arg:"--mqtt-impl" help:"Implementation for MQTT ('none' to disable)" placeholder:"NAME"`
	WS    string `arg:"--ws-impl" help:"Implementation for WebSocket ('none' to disable)" placeholder:"NAME"`
	Redis string `arg:"--redis-impl" help:"Implementation for Redis ('none' to disable)" placeholder:"NAME"`
	IP    string `arg:"--ip-impl" help:"Implementation for IP raw sockets ('none' to disable)" placeholder:"NAME"`
	TCP   string `arg:"--tcp-impl" help:"Implementation for TCP ('none' to disable)" placeholder:"NAME"`
	UDP   string `arg:"--udp-impl" help:"Implementation for UDP ('none' to disable)" placeholder:"NAME"`
}

func generate(cmd *GenerateCmd) error {
	logger := log.GetLogger("")
	isPub, isSub, pubSubOpts := getPubSubVariant(cmd)
	mergedConfig, err := mergeConfig(cmd, pubSubOpts)
	if err != nil {
		return err
	}

	if cmd.Implementation != nil {
		return generateImplementation(cmd.Implementation, mergedConfig)
	}

	asyncapi.ProtocolBuilders = protocolBuilders()
	if !isSub && !isPub {
		return fmt.Errorf("%w: publisher, subscriber or both are required in args", ErrWrongCliArgs)
	}

	if logger.GetLevel() == log.TraceLevel {
		buf := lo.Must(yaml.Marshal(mergedConfig))
		logger.Trace("Use the resulting config", "value", string(buf))
	}
	compileOpts := getCompileOpts(mergedConfig, isPub, isSub)
	renderOpts, err := getRenderOpts(mergedConfig, mergedConfig.Directories.Target)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	implDir := path.Join(mergedConfig.Directories.Target, mergedConfig.Directories.Implementations)
	logger.Debugf("Target implementations directory is %s", implDir)
	tmpl.ParseTemplates(mergedConfig.Directories.Templates)

	// Compilation
	resolver := getResolver(mergedConfig)
	specURL := specurl.Parse(pubSubOpts.Spec)
	modules, err := generationCompile(specURL, compileOpts, resolver)
	if err != nil {
		return err
	}
	objSources := lo.MapValues(modules, func(value *compiler.Module, _ string) linker.ObjectSource { return value })
	mainModule := modules[specURL.SpecID]

	// Linking
	if err = generationLinking(objSources); err != nil {
		return err
	}

	// Rendering
	// Render definitions from all modules
	allObjects := lo.FlatMap(lo.Values(modules), func(m *compiler.Module, _ int) []common.CompileObject { return m.AllObjects() })
	files, err := writer.RenderObjects(allObjects, renderOpts)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	// Formatting
	if !renderOpts.DisableFormatting {
		if err = writer.FormatFiles(files); err != nil {
			return fmt.Errorf("formatting code: %w", err)
		}
	}

	// Writing
	if err = writer.WriteToFiles(files, mergedConfig.Directories.Target); err != nil {
		return fmt.Errorf("writing code to files: %w", err)
	}

	// Rendering the selected implementations
	if !mergedConfig.NoImplementations {
		selectedImpls := getImplementationsOpts(mergedConfig.Implementations)
		if err = generationWriteImplementations(selectedImpls, mainModule.Protocols(), implDir); err != nil {
			return err
		}
	}

	logger.Info("Finished")
	return nil
}

func generateImplementation(cmd *generateImplementationArgs, conf toolConfig) error {
	implDir := path.Join(conf.Directories.Target, conf.Directories.Implementations)  // TODO: remove this line
	log.GetLogger("").Debugf("Target implementations directory is %s", implDir)
	proto := cmd.Protocol
	name := cmd.Name
	if err := generationWriteImplementations(map[string]string{proto: name}, []string{proto}, implDir); err != nil {
		return err
	}

	log.GetLogger("").Info("Finished")
	return nil
}

func protocolBuilders() map[string]asyncapi.ProtocolBuilder {
	return map[string]asyncapi.ProtocolBuilder{
		amqp.Builder.ProtocolName():  amqp.Builder,
		http.Builder.ProtocolName():  http.Builder,
		kafka.Builder.ProtocolName(): kafka.Builder,
		mqtt.Builder.ProtocolName():  mqtt.Builder,
		ws.Builder.ProtocolName():    ws.Builder,
		redis.Builder.ProtocolName(): redis.Builder,
		ip.Builder.ProtocolName():    ip.Builder,
		tcp.Builder.ProtocolName():   tcp.Builder,
		udp.Builder.ProtocolName():   udp.Builder,
	}
}

func mergeConfig(cmd *GenerateCmd, generateArgs *generatePubSubArgs) (config toolConfig, err error) {
	if config, err = loadDefaultConfig(); err != nil {
		return
	}
	if lo.IsNil(generateArgs) {
		generateArgs = &generatePubSubArgs{}
	}

	var userConfig toolConfig
	if cmd.ConfigFile != "" {
		log.GetLogger("").Debug("Load config", "file", cmd.ConfigFile)
		if userConfig, err = loadConfig(cmd.ConfigFile); err != nil {
			return
		}
	}

	config.ConfigVersion = coalesce(userConfig.ConfigVersion, config.ConfigVersion)
	config.ProjectModule = coalesce(generateArgs.ProjectModule, userConfig.ProjectModule, config.ProjectModule)
	config.RuntimeModule = coalesce(generateArgs.RuntimeModule, userConfig.RuntimeModule, config.RuntimeModule)
	config.Directories.Templates = coalesce(generateArgs.TemplateDir, userConfig.Directories.Templates, config.Directories.Templates)
	config.Directories.Target = coalesce(cmd.TargetDir, userConfig.Directories.Target, config.Directories.Target)
	config.Directories.Implementations = coalesce(cmd.ImplDir, userConfig.Directories.Implementations, config.Directories.Implementations)

	// *Replace* selections
	if len(userConfig.Selections) > 0 {
		config.Selections = userConfig.Selections
	}

	config.Resolver.AllowRemoteReferences = coalesce(generateArgs.AllowRemoteRefs, userConfig.Resolver.AllowRemoteReferences, config.Resolver.AllowRemoteReferences)
	config.Resolver.SearchDirectory = coalesce(generateArgs.ResolverSearchDir, userConfig.Resolver.SearchDirectory, config.Resolver.SearchDirectory)
	config.Resolver.Timeout = coalesce(generateArgs.ResolverTimeout, userConfig.Resolver.Timeout, config.Resolver.Timeout)
	config.Resolver.Command = coalesce(generateArgs.ResolverCommand, userConfig.Resolver.Command, config.Resolver.Command)
	config.NoImplementations = coalesce(userConfig.NoImplementations, config.NoImplementations)
	config.Render.PreambleTemplate = coalesce(generateArgs.PreambleTemplate, userConfig.Render.PreambleTemplate, config.Render.PreambleTemplate)
	config.Render.DisableFormatting = coalesce(generateArgs.DisableFormatting, userConfig.Render.DisableFormatting, config.Render.DisableFormatting)

	// Merge implementations options
	if len(userConfig.Implementations) > 0 {
		for k, v := range userConfig.Implementations {
			config.Implementations[k] = v
		}
	}

	return
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

func getCompileOpts(cfg toolConfig, isPub, isSub bool) common.CompileOpts {
	return common.CompileOpts{
		AllowRemoteRefs:     cfg.Resolver.AllowRemoteReferences,
		RuntimeModule:       cfg.RuntimeModule,
		GeneratePublishers:  isPub,
		GenerateSubscribers: isSub,
	}
}

func getRenderOpts(conf toolConfig, targetDir string) (common.RenderOpts, error) {
	logger := log.GetLogger("")
	res := common.RenderOpts{
		RuntimeModule:     conf.RuntimeModule,
		TargetDir:         targetDir,
		DisableFormatting: conf.Render.DisableFormatting,
		PreambleTemplate:  conf.Render.PreambleTemplate,
	}

	// Selections
	for _, item := range conf.Selections {
		pkg, _ := lo.Coalesce(
			item.Render.Package,
			lo.Ternary(path.Dir(item.Render.File) != ".", path.Dir(item.Render.File), ""),
			lo.Ternary(targetDir != "", path.Base(targetDir), ""),
			defaultPackage,
		)
		templateName, _ := lo.Coalesce(item.Render.Template, defaultTemplate)
		sel := common.RenderSelectionConfig{
			Protocols:        item.Protocols,
			ObjectKinds:     item.ObjectKinds,
			ModuleURLRe:      item.ModuleURLRe,
			PathRe:           item.PathRe,
			NameRe: item.NameRe,
			Render: common.RenderSelectionConfigRender{
				Template:         templateName,
				File:             item.Render.File,
				Package:          pkg,
				Protocols:        item.Render.Protocols,
				ProtoObjectsOnly: item.Render.ProtoObjectsOnly,
			},
			ReusePackagePath: item.ReusePackagePath,
			AllSupportedProtocols: lo.Keys(asyncapi.ProtocolBuilders),
		}
		logger.Debug("Use selection", "value", sel)
		res.Selections = append(res.Selections, sel)
	}

	// ImportBase
	res.ImportBase = conf.ProjectModule
	if res.ImportBase == "" {
		m, err := getProjectModule()
		if err != nil {
			return res, fmt.Errorf("getting project module from ./go.mod (use -M arg to override): %w", err)
		}
		logger.Debug("Determined project module", "value", m)
		// Clean path and remove empty, current and parent directories, leaving only names
		// This is not the best solution, however, it should work for most cases. Moreover, user can always override it.
		parts := lo.Filter(strings.Split(path.Clean(targetDir), string(os.PathSeparator)), func(s string, _ int) bool {
			return !lo.Contains([]string{"", ".", ".."}, s)
		})
		res.ImportBase = path.Join(m, path.Join(parts...))
	}
	logger.Debug("Import base", "value", res.ImportBase)

	return res, nil
}

func getProjectModule() (string, error) {
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

func getResolver(conf toolConfig) compiler.SpecFileResolver {
	logger := log.GetLogger(log.LoggerPrefixResolving)
	if conf.Resolver.Command != "" {
		return compiler.SubprocessSpecFileResolver{
			CommandLine: conf.Resolver.Command,
			RunTimeout:  conf.Resolver.Timeout,
			Logger:      logger,
		}
	}
	return compiler.DefaultSpecFileResolver{
		Client:  stdHTTP.DefaultClient,
		Timeout: conf.Resolver.Timeout,
		BaseDir: conf.Resolver.SearchDirectory,
		Logger:  logger,
	}
}

func generationCompile(specURL *specurl.URL, compileOpts common.CompileOpts, resolver compiler.SpecFileResolver) (map[string]*compiler.Module, error) {
	logger := log.GetLogger(log.LoggerPrefixCompilation)
	compileQueue := []*specurl.URL{specURL}      // Queue of specIDs to compile
	modules := make(map[string]*compiler.Module) // Compilers by spec id
	for len(compileQueue) > 0 {
		specURL, compileQueue = compileQueue[0], compileQueue[1:] // Pop an item from queue
		if _, ok := modules[specURL.SpecID]; ok {
			continue // Skip if a spec file has been already compiled
		}

		logger.Info("Run compilation", "specURL", specURL)
		module := compiler.NewModule(specURL)
		modules[specURL.SpecID] = module

		if !compileOpts.AllowRemoteRefs && specURL.IsRemote() {
			return nil, fmt.Errorf(
				"%s: external requests are forbidden by default for security reasons, use --allow-remote-refs flag to allow them",
				specURL,
			)
		}
		logger.Debug("Loading a spec", "specURL", specURL)
		if err := module.Load(resolver); err != nil {
			return nil, fmt.Errorf("load a spec: %w", err)
		}
		logger.Debug("Compilation a spec", "specURL", specURL)
		if err := module.Compile(common.NewCompileContext(specURL, compileOpts)); err != nil {
			return nil, fmt.Errorf("compilation a spec: %w", err)
		}
		logger.Debugf("Compiler stats: %s", module.Stats())
		compileQueue = lo.Flatten([][]*specurl.URL{compileQueue, module.ExternalSpecs()}) // Extend queue with remote specPaths
	}

	logger.Info("Compilation completed", "files", len(modules))
	return modules, nil
}

func generationLinking(objSources map[string]linker.ObjectSource) error {
	logger := log.GetLogger(log.LoggerPrefixLinking)
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

func getImplementationsOpts(opts map[string]toolConfigImplementation) map[string]string {
	return map[string]string{
		amqp.Builder.ProtocolName():  opts["amqp"].Name,
		http.Builder.ProtocolName():  opts["http"].Name,
		kafka.Builder.ProtocolName(): opts["kafka"].Name,
		mqtt.Builder.ProtocolName():  opts["mqtt"].Name,
		ws.Builder.ProtocolName():    opts["ws"].Name,
		redis.Builder.ProtocolName(): opts["redis"].Name,
		ip.Builder.ProtocolName():    opts["ip"].Name,
		tcp.Builder.ProtocolName():   opts["tcp"].Name,
		udp.Builder.ProtocolName():   opts["udp"].Name,
	}
}

func generationWriteImplementations(selectedImpls map[string]string, protocols []string, implDir string) error {
	logger := log.GetLogger(log.LoggerPrefixWriting)
	logger.Info("Writing implementations")
	implManifest := lo.Must(loadImplementationsManifest())

	var totalBytes int
	var writtenProtocols []string
	for _, p := range protocols {
		implName := selectedImpls[p]
		switch implName {
		case "none":
			logger.Debug("Implementation has been unselected", "protocol", p)
			continue
		case "":
			logger.Info("No implementation for the protocol", "protocol", p)
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

func loadImplementationsManifest() (implementations.ImplManifest, error) {
	f, err := implementations.ImplementationFS.Open("manifest.json")
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

func coalesce[T comparable](vals ...T) T {
	res, _ := lo.Coalesce(vals...)
	return res
}