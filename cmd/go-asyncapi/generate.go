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

	TargetDir string `arg:"-t,--target-dir" default:"./asyncapi" help:"Directory to save the generated code" placeholder:"DIR"`
	ImplDir   string `arg:"--impl-dir" default:"impl" help:"Directory to save implementations inside the target dir" placeholder:"DIR"`
}

type generatePubSubArgs struct {
	Spec string `arg:"required,positional" help:"AsyncAPI specification file path or url" placeholder:"PATH"`

	ProjectModule string `arg:"-M,--module" help:"Module path to use [default: extracted from go.mod file in the current working directory concatenated with target dir]" placeholder:"MODULE"`
	ConfigFile string `arg:"-c,--config-file" help:"YAML configuration file path" placeholder:"PATH"`
	TemplateDir string `arg:"-T,--template-dir" help:"Directory with custom templates" placeholder:"DIR"`
	PreambleTemplate string `arg:"--preamble-template" default:"preamble.tmpl" help:"Custom preamble template name" placeholder:"NAME"`
	DisableFormatting bool `arg:"--disable-formatting" help:"Disable code formatting"`
	ImplementationsOpts
	AllowRemoteRefs bool `arg:"--allow-remote-refs" help:"Allow fetching spec files from remote $ref URLs"`

	RuntimeModule         string        `arg:"--runtime-module" default:"github.com/bdragon300/go-asyncapi/run" help:"Runtime module name" placeholder:"MODULE"`
	FileResolverSearchDir string        `arg:"--file-resolver-search-dir" help:"Directory to search the local spec files for [default: current working directory]" placeholder:"PATH"`
	FileResolverTimeout   time.Duration `arg:"--file-resolver-timeout" default:"30s" help:"Timeout for file resolver to resolve a spec file" placeholder:"DURATION"`
	FileResolverCommand   string        `arg:"--file-resolver-command" help:"Custom file resolver executable to use instead of built-in resolver" placeholder:"PATH"`
}

type generateImplementationArgs struct {
	Protocol string `arg:"required,positional" help:"Protocol name to generate"`
	Name     string `arg:"required,positional" help:"Implementation name to generate"`
}

type ImplementationsOpts struct {
	NoImplementations bool `arg:"--no-implementations" help:"Do not generate any protocol implementation"`

	Kafka string `arg:"--kafka-impl" default:"franz-go" help:"Implementation for Kafka ('none' to disable)" placeholder:"NAME"`
	AMQP  string `arg:"--amqp-impl" default:"amqp091-go" help:"Implementation for AMQP ('none' to disable)" placeholder:"NAME"`
	HTTP  string `arg:"--http-impl" default:"std" help:"Implementation for HTTP ('none' to disable)" placeholder:"NAME"`
	MQTT  string `arg:"--mqtt-impl" default:"paho-mqtt" help:"Implementation for MQTT ('none' to disable)" placeholder:"NAME"`
	WS    string `arg:"--ws-impl" default:"gobwas-ws" help:"Implementation for WebSocket ('none' to disable)" placeholder:"NAME"`
	Redis string `arg:"--redis-impl" default:"go-redis" help:"Implementation for Redis ('none' to disable)" placeholder:"NAME"`
	IP    string `arg:"--ip-impl" default:"std" help:"Implementation for IP raw sockets ('none' to disable)" placeholder:"NAME"`
	TCP   string `arg:"--tcp-impl" default:"std" help:"Implementation for TCP ('none' to disable)" placeholder:"NAME"`
	UDP   string `arg:"--udp-impl" default:"std" help:"Implementation for UDP ('none' to disable)" placeholder:"NAME"`
}

func generate(cmd *GenerateCmd) error {
	if cmd.Implementation != nil {
		return generateImplementation(cmd)
	}

	asyncapi.ProtocolBuilders = protocolBuilders()
	isPub, isSub, pubSubOpts := getPubSubVariant(cmd)
	if !isSub && !isPub {
		return fmt.Errorf("%w: publisher, subscriber or both are required in args", ErrWrongCliArgs)
	}
	compileOpts := getCompileOpts(*pubSubOpts, isPub, isSub)
	renderOpts, err := getRenderOpts(*pubSubOpts, cmd.TargetDir)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	implDir := path.Join(cmd.TargetDir, cmd.ImplDir)
	log.GetLogger("").Debugf("Target implementations directory is %s", implDir)
	tmpl.ParseTemplates(pubSubOpts.TemplateDir)

	// Compilation
	resolver := getResolver(*pubSubOpts)
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

	log.GetLogger("").Info("Finished")
	return nil
}

func generateImplementation(cmd *GenerateCmd) error {
	implDir := path.Join(cmd.TargetDir, cmd.ImplDir)
	log.GetLogger("").Debugf("Target implementations directory is %s", implDir)
	proto := cmd.Implementation.Protocol
	name := cmd.Implementation.Name
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

func getCompileOpts(opts generatePubSubArgs, isPub, isSub bool) common.CompileOpts {
	return common.CompileOpts{
		AllowRemoteRefs:     opts.AllowRemoteRefs,
		RuntimeModule:       opts.RuntimeModule,
		GeneratePublishers:  isPub,
		GenerateSubscribers: isSub,
	}
}

func getRenderOpts(opts generatePubSubArgs, targetDir string) (common.RenderOpts, error) {
	logger := log.GetLogger("")
	res := common.RenderOpts{
		RuntimeModule: opts.RuntimeModule,
		TargetDir:     targetDir,
		DisableFormatting: opts.DisableFormatting,
		PreambleTemplate: opts.PreambleTemplate,
	}

	// Selections
	logger.Debug("Load config", "file", opts.ConfigFile)
	conf, err := loadConfig(opts.ConfigFile)
	if err != nil {
		return res, err
	}
	if logger.GetLevel() == log.TraceLevel {
		buf := lo.Must(yaml.Marshal(conf))
		logger.Trace("Loaded config", "value", string(buf))
	}
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
	res.ImportBase = opts.ProjectModule
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

func getResolver(opts generatePubSubArgs) compiler.SpecFileResolver {
	logger := log.GetLogger(log.LoggerPrefixResolving)
	if opts.FileResolverCommand != "" {
		return compiler.SubprocessSpecFileResolver{
			CommandLine: opts.FileResolverCommand,
			RunTimeout:  opts.FileResolverTimeout,
			Logger:      logger,
		}
	}
	return compiler.DefaultSpecFileResolver{
		Client:  stdHTTP.DefaultClient,
		Timeout: opts.FileResolverTimeout,
		BaseDir: opts.FileResolverSearchDir,
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

func getImplementationsOpts(opts ImplementationsOpts) map[string]string {
	return map[string]string{
		amqp.Builder.ProtocolName():  opts.AMQP,
		http.Builder.ProtocolName():  opts.HTTP,
		kafka.Builder.ProtocolName(): opts.Kafka,
		mqtt.Builder.ProtocolName():  opts.MQTT,
		ws.Builder.ProtocolName():    opts.WS,
		redis.Builder.ProtocolName(): opts.Redis,
		ip.Builder.ProtocolName():    opts.IP,
		tcp.Builder.ProtocolName():   opts.TCP,
		udp.Builder.ProtocolName():   opts.UDP,
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
