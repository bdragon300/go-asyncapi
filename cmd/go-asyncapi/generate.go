package main

import (
	"encoding/json"
	"fmt"
	"github.com/bdragon300/go-asyncapi/assets"
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
	"github.com/bdragon300/go-asyncapi/internal/types"
	stdHTTP "net/http"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/linker"
	"github.com/bdragon300/go-asyncapi/internal/writer"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
)

const defaultConfigFileName = "default_config.yaml"

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
	TemplateDir string `arg:"--template-dir" help:"Directory with custom templates" placeholder:"DIR"`
	ConfigFile string `arg:"--config-file" help:"YAML configuration file path" placeholder:"PATH"`
	generateObjectSelectionOpts
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

type generateObjectSelectionOpts struct {
	//SelectChannelsAll   bool   `arg:"--select-channels-all" help:"Select all channels to be generated"`
	//SelectChannelsRe    string `arg:"--select-channels-re" help:"Select channels whose name in document matches the regex" placeholder:"REGEX"`
	//IgnoreChannelsAll   bool   `arg:"--ignore-channels-all" help:"Ignore all channels to be generated"`
	//IgnoreChannelsRe    string `arg:"--ignore-channels-re" help:"Ignore channels whose name in document matches the regex" placeholder:"REGEX"`
	//ReuseChannelsModule string `arg:"--reuse-channels-module" help:"Reuse the module with channels code" placeholder:"MODULE"`
	//
	//SelectMessagesAll   bool   `arg:"--select-messages-all" help:"Select all messages to be generated"`
	//SelectMessagesRe    string `arg:"--select-messages-re" help:"Select messages whose name in document matches the regex" placeholder:"REGEX"`
	//IgnoreMessagesAll   bool   `arg:"--ignore-messages-all" help:"Ignore all messages to be generated"`
	//IgnoreMessagesRe    string `arg:"--ignore-messages-re" help:"Ignore messages whose name in document matches the regex" placeholder:"REGEX"`
	//ReuseMessagesModule string `arg:"--reuse-messages-module" help:"Reuse the module with messages code" placeholder:"MODULE"`
	//
	//SelectModelsAll   bool   `arg:"--select-models-all" help:"Select all models to be generated"`
	//SelectModelsRe    string `arg:"--select-models-re" help:"Select models whose name in document matches the regex" placeholder:"REGEX"`
	//IgnoreModelsAll   bool   `arg:"--ignore-models-all" help:"Ignore all models to be generated"`
	//IgnoreModelsRe    string `arg:"--ignore-models-re" help:"Ignore models whose name in document matches the regex" placeholder:"REGEX"`
	//ReuseModelsModule string `arg:"--reuse-models-module" help:"Reuse the module with models code" placeholder:"MODULE"`
	//
	//SelectServersAll   bool   `arg:"--select-servers=all" help:"Select all servers to be generated"`
	//SelectServersRe    string `arg:"--select-servers-re" help:"Select servers whose name in document matches the regex" placeholder:"REGEX"`
	//IgnoreServersAll   bool   `arg:"--ignore-servers-all" help:"Ignore all servers to be generated"`
	//IgnoreServersRe    string `arg:"--ignore-servers-re" help:"Ignore servers whose name in document matches the regex" placeholder:"REGEX"`
	//ReuseServersModule string `arg:"--reuse-servers-module" help:"Reuse the module with servers code" placeholder:"MODULE"`

	NoImplementations bool `arg:"--no-implementations" help:"Do not generate any protocol implementation"`
	//NoEncoding        bool `arg:"--no-encoding" help:"Do not generate encoders/decoders code"`
}

func generate(cmd *GenerateCmd) error {
	if cmd.Implementation != nil {
		return generateImplementation(cmd)
	}

	isPub, isSub, pubSubOpts := getPubSubVariant(cmd)
	if !isSub && !isPub {
		return fmt.Errorf("%w: no publisher or subscriber set to generate", ErrWrongCliArgs)
	}
	compileOpts, err := getCompileOpts(*pubSubOpts, isPub, isSub)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	renderOpts, err := getRenderOpts(*pubSubOpts, cmd.TargetDir)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	implDir := path.Join(cmd.TargetDir, cmd.ImplDir)
	mainLogger.Debugf("Target implementations directory is %s", implDir)

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
	files, err := writer.RenderPackages(mainModule, renderOpts)
	if err != nil {
		return fmt.Errorf("schema render: %w", err)
	}

	// Formatting
	if err = writer.FormatFiles(files); err != nil {
		return fmt.Errorf("formatting code: %w", err)
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

func generationCompile(specURL *specurl.URL, compileOpts common.CompileOpts, resolver compiler.SpecFileResolver) (map[string]*compiler.Module, error) {
	logger := types.NewLogger("Compilation 🔨")
	asyncapi.ProtocolBuilders = protocolBuilders()
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
		mqtt.Builder.ProtocolName():  opts.MQTT,
		ws.Builder.ProtocolName():    opts.WS,
		redis.Builder.ProtocolName(): opts.Redis,
		ip.Builder.ProtocolName():    opts.IP,
		tcp.Builder.ProtocolName():   opts.TCP,
		udp.Builder.ProtocolName():   opts.UDP,
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

func getCompileOpts(opts generatePubSubArgs, isPub, isSub bool) (common.CompileOpts, error) {
	//var err error
	res := common.CompileOpts{
		//NoEncodingPackage:   opts.NoEncoding,
		AllowRemoteRefs:     opts.AllowRemoteRefs,
		RuntimeModule:       opts.RuntimeModule,
		GeneratePublishers:  isPub,
		GenerateSubscribers: isSub,
	}

	//includeAll := !opts.SelectChannelsAll && !opts.SelectMessagesAll && !opts.SelectModelsAll && !opts.SelectServersAll
	//f := func(all, ignoreAll bool, re, ignoreRe string) (r common.ObjectCompileOpts, e error) {
	//	r.Enable = (includeAll || all) && !ignoreAll
	//	if re != "" {
	//		if r.IncludeRegex, e = regexp.Compile(re); e != nil {
	//			return
	//		}
	//	}
	//	if ignoreRe != "" {
	//		if r.ExcludeRegex, e = regexp.Compile(ignoreRe); e != nil {
	//			return
	//		}
	//	}
	//	return
	//}

	//if res.ChannelOpts, err = f(opts.SelectChannelsAll, opts.IgnoreChannelsAll, opts.SelectChannelsRe, opts.IgnoreChannelsRe); err != nil {
	//	return res, err
	//}
	//if opts.ReuseChannelsModule != "" {
	//	res.ReusePackages[asyncapi.PackageScopeChannels] = opts.ReuseChannelsModule
	//}
	//if res.MessageOpts, err = f(opts.SelectMessagesAll, opts.IgnoreMessagesAll, opts.SelectMessagesRe, opts.IgnoreMessagesRe); err != nil {
	//	return res, err
	//}
	//if opts.ReuseMessagesModule != "" {
	//	res.ReusePackages[asyncapi.PackageScopeMessages] = opts.ReuseMessagesModule
	//}
	//if res.ModelOpts, err = f(opts.SelectModelsAll, opts.IgnoreModelsAll, opts.SelectModelsRe, opts.IgnoreModelsRe); err != nil {
	//	return res, err
	//}
	//if opts.ReuseModelsModule != "" {
	//	res.ReusePackages[asyncapi.PackageScopeModels] = opts.ReuseModelsModule
	//}
	//if res.ServerOpts, err = f(opts.SelectServersAll, opts.IgnoreServersAll, opts.SelectServersRe, opts.IgnoreServersRe); err != nil {
	//	return res, err
	//}
	//if opts.ReuseServersModule != "" {
	//	res.ReusePackages[asyncapi.PackageScopeServers] = opts.ReuseServersModule
	//}

	return res, nil
}

func getResolver(opts generatePubSubArgs) compiler.SpecFileResolver {
	logger := types.NewLogger("Resolving 📡")
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

func getRenderOpts(opts generatePubSubArgs, targetDir string) (common.RenderOpts, error) {
	res := common.RenderOpts{
		RuntimeModule: opts.RuntimeModule,
		TargetDir:     targetDir,
		TemplateDir: opts.TemplateDir,
	}

	// TODO: logging
	// Selections
	if opts.ConfigFile == "" {
		conf, err := loadConfig(opts.ConfigFile)
		if err != nil {
			return res, err
		}
		for _, item := range conf.Render.Selections {
			pkg, _ := lo.Coalesce(item.Package, lo.Ternary(targetDir != "", path.Base(targetDir), "main"))
			templateName, _ := lo.Coalesce(item.Template, "main")
			sel := common.RenderSelectionConfig{
				Template:     templateName,
				File:         item.File,
				Package:      pkg,
				TemplateArgs: item.TemplateArgs,
				ObjectKindRe: item.ObjectKindRe,
				ModuleURLRe:  item.ModuleURLRe,
				PathRe:       item.PathRe,
			}
			res.Selections = append(res.Selections, sel)
		}
	}

	// ImportBase
	importBase := opts.ProjectModule
	if importBase == "" {
		b, err := getImportBase()
		if err != nil {
			return res, fmt.Errorf("determine project's module (use -M arg to override): %w", err)
		}
		importBase = b
	}
	mainLogger.Debugf("Target import base is %s", importBase)
	res.ImportBase = importBase

	return res, nil
}

func loadConfig(fileName string) (toolConfig, error) {
	var conf toolConfig

	var f io.ReadCloser
	var err error
	if fileName == "" {
		f, err = assets.AssetsFS.Open(defaultConfigFileName)
		if err != nil {
			return conf, fmt.Errorf("cannot open default config file in assets, this is a programming error: %w", err)
		}
	} else {
		f, err = os.Open(fileName)
		if err != nil {
			return conf, fmt.Errorf("cannot open config file: %w", err)
		}
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return conf, fmt.Errorf("cannot read config file: %w", err)
	}

	if err = yaml.Unmarshal(buf, &conf); err != nil {
		return conf, fmt.Errorf("cannot parse YAML config file: %w", err)
	}
	return conf, nil
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

func getImplementationsManifest() (implementations.ImplManifest, error) {
	f, err := implementations.ImplementationsFS.Open("manifest.json")
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

