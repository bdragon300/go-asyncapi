package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/resolver"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/writer"
	"github.com/bdragon300/go-asyncapi/templates/client"
	templates "github.com/bdragon300/go-asyncapi/templates/code"
	"gopkg.in/yaml.v3"

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
	"github.com/bdragon300/go-asyncapi/internal/renderer"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
)

const (
	mainTemplateName = "main.tmpl"
)

type GenerateCmd struct {
	Pub    *generatePubSubArgs `arg:"subcommand:pub" help:"Generate only the publisher code"`
	Sub    *generatePubSubArgs `arg:"subcommand:sub" help:"Generate only the subscriber code"`
	PubSub *generatePubSubArgs `arg:"subcommand:pubsub" help:"Generate both publisher and subscriber code"`

	TargetDir string `arg:"-t,--target-dir" help:"Directory to save the generated code" placeholder:"DIR"`
}

type generatePubSubArgs struct {
	Spec string `arg:"required,positional" help:"AsyncAPI specification file path or url" placeholder:"PATH"`

	ProjectModule string `arg:"-M,--module" help:"Project module in the generated code. By default, read get from go.mod in the current working directory" placeholder:"MODULE"`
	RuntimeModule string `arg:"--runtime-module" help:"Runtime module path" placeholder:"MODULE"`

	TemplateDir       string `arg:"-T,--template-dir" help:"Directory with custom templates" placeholder:"DIR"`
	PreambleTemplate  string `arg:"--preamble-template" help:"Custom preamble template name" placeholder:"NAME"`
	DisableFormatting bool   `arg:"--disable-formatting" help:"Disable code formatting"`

	AllowRemoteRefs   bool          `arg:"--allow-remote-refs" help:"Allow resolver to fetch the files from remote $ref URLs"`
	ResolverSearchDir string        `arg:"--resolver-search-dir" help:"Directory to search the local spec files for [default: current working directory]" placeholder:"PATH"`
	ResolverTimeout   time.Duration `arg:"--resolver-timeout" help:"Timeout for resolver to resolve a spec file" placeholder:"DURATION"`
	ResolverCommand   string        `arg:"--resolver-command" help:"Custom resolver executable to use instead of built-in resolver" placeholder:"PATH"`

	ClientApp     bool   `arg:"--client-app" help:"Generate a client application code as well"`
	goModTemplate string `arg:"-"`
}

func cliGenerate(cmd *GenerateCmd, globalConfig toolConfig) error {
	logger := log.GetLogger("")
	isPub, isSub, pubSubOpts := getPubSubVariant(cmd)
	cmdConfig := cliGenerateMergeConfig(globalConfig, cmd, pubSubOpts)

	if logger.GetLevel() == log.TraceLevel {
		buf := lo.Must(yaml.Marshal(cmdConfig))
		logger.Trace("Use the resulting config", "value", string(buf))
	}

	asyncapi.ProtocolBuilders = protocolBuilders()
	if !isSub && !isPub {
		return fmt.Errorf("%w: publisher, subscriber or both are required in args", ErrWrongCliArgs)
	}

	compileOpts := getCompileOpts(cmdConfig, isPub, isSub)
	renderOpts, err := getRenderOpts(cmdConfig, cmdConfig.Directories.Target, true)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	renderManager := manager.NewTemplateRenderManager(renderOpts)

	//
	// Compilation & linking
	//
	fileResolver := getResolver(cmdConfig)
	specURL := specurl.Parse(pubSubOpts.Spec)
	modules, err := runCompilationLinking(fileResolver, specURL, compileOpts)
	if err != nil {
		return fmt.Errorf("compilation: %w", err)
	}

	//
	// Rendering
	//
	mainModule := modules[specURL.SpecID]
	activeProtocols := collectActiveProtocols(mainModule.AllObjects())
	logger.Debug("Renders protocols", "value", activeProtocols)

	// Implementations
	implementationOpts := getImplementationOpts(cmdConfig.Implementations)
	if !implementationOpts.Disable {
		tplLoader := tmpl.NewTemplateLoader(mainTemplateName, implementations.ImplementationFS)
		renderManager.TemplateLoader = tplLoader

		supportedProtocols := lo.Keys(asyncapi.ProtocolBuilders)
		protocols := lo.Intersect(supportedProtocols, activeProtocols)
		if len(protocols) < len(activeProtocols) {
			logger.Warn("Some protocols have no implementations", "protocols", lo.Without(activeProtocols, protocols...))
		}

		// Render only implementations for protocols that are actually used in the spec
		slices.Sort(protocols)
		logger.Debug("Run implementations rendering", "protocols", protocols)
		implObjects, err := getImplementations(implementationOpts, protocols)
		if err != nil {
			return fmt.Errorf("getting implementations: %w", err)
		}
		if err = renderer.RenderImplementations(implObjects, renderManager); err != nil {
			return fmt.Errorf("render implementations: %w", err)
		}
		logger.Debug("Implementations rendering complete")
	}

	// Module objects
	logger.Debug("Run objects rendering")
	templateDirs := []fs.FS{templates.TemplateFS}
	if cmdConfig.Directories.Templates != "" {
		logger.Debug("Custom templates location", "directory", cmdConfig.Directories.Templates)
		templateDirs = append(templateDirs, os.DirFS(cmdConfig.Directories.Templates))
	}
	tplLoader := tmpl.NewTemplateLoader(mainTemplateName, templateDirs...)
	logger.Trace("Parse templates", "dirs", templateDirs)
	renderManager.TemplateLoader = tplLoader
	if err = tplLoader.ParseRecursive(renderManager); err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}
	allObjects := lo.FlatMap(lo.Values(modules), func(m *compiler.Module, _ int) []common.CompileObject { return m.AllObjects() })
	renderQueue := selectObjects(allObjects, renderOpts.Selections)
	if err = renderer.RenderObjects(renderQueue, renderManager); err != nil {
		return fmt.Errorf("render objects: %w", err)
	}
	logger.Debug("Objects rendering complete")

	// Client app
	if pubSubOpts.ClientApp {
		logger.Debug("Run client app rendering")
		templateDirs = []fs.FS{templates.TemplateFS, client.TemplateFS}
		if cmdConfig.Directories.Templates != "" {
			logger.Debug("Custom templates location", "directory", cmdConfig.Directories.Templates)
			templateDirs = append(templateDirs, os.DirFS(cmdConfig.Directories.Templates))
		}
		tplLoader = tmpl.NewTemplateLoader(mainTemplateName, templateDirs...)
		logger.Trace("Parse templates", "dirs", templateDirs)
		renderManager.TemplateLoader = tplLoader
		if err = tplLoader.ParseRecursive(renderManager); err != nil {
			return fmt.Errorf("parse templates: %w", err)
		}

		if err = renderer.RenderClientApp(renderQueue, activeProtocols, cmdConfig.Client.GoModTemplate, cmdConfig.Client.OutputSourceFile, renderManager); err != nil {
			return fmt.Errorf("render client app: %w", err)
		}
		logger.Debug("Client app rendering complete")
	}

	// Render the final result: preamble, etc.
	logger.Debug("Finish the files rendering")
	files, err := renderer.FinishFiles(renderManager)
	if err != nil {
		return fmt.Errorf("finish files: %w", err)
	}
	logger.Debug("Rendering finishing complete")

	//
	// Formatting
	//
	if !renderOpts.DisableFormatting {
		logger.Debug("Run formatting")
		if err = writer.FormatFiles(files); err != nil {
			return fmt.Errorf("formatting: %w", err)
		}
		logger.Debug("Formatting complete")
	}

	//
	// Writing
	//
	logger.Debug("Run writing")
	if err = writer.WriteBuffersToFiles(files, cmdConfig.Directories.Target); err != nil {
		return fmt.Errorf("writing: %w", err)
	}
	logger.Debug("Writing complete")

	logger.Info("Code generation finished")
	return nil
}

func collectActiveProtocols(allObjects []common.CompileObject) []string {
	return lo.Uniq(lo.FilterMap(allObjects, func(obj common.CompileObject, _ int) (string, bool) {
		if obj.Kind() != common.ObjectKindServer && !obj.Selectable() || !obj.Visible() {
			return "", false
		}
		obj2 := common.DerefRenderable(obj.Renderable)
		v, ok := obj2.(*render.Server)
		if !ok {
			return "", false
		}
		return v.Protocol, true
	}))
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

func cliGenerateMergeConfig(globalConfig toolConfig, cmd *GenerateCmd, generateArgs *generatePubSubArgs) toolConfig {
	res := globalConfig

	res.ProjectModule = coalesce(generateArgs.ProjectModule, res.ProjectModule)
	res.RuntimeModule = coalesce(generateArgs.RuntimeModule, res.RuntimeModule)
	res.Directories.Templates = coalesce(generateArgs.TemplateDir, res.Directories.Templates)
	res.Directories.Target = coalesce(cmd.TargetDir, res.Directories.Target)

	res.Resolver.AllowRemoteReferences = coalesce(generateArgs.AllowRemoteRefs, res.Resolver.AllowRemoteReferences)
	res.Resolver.SearchDirectory = coalesce(generateArgs.ResolverSearchDir, res.Resolver.SearchDirectory)
	res.Resolver.Timeout = coalesce(generateArgs.ResolverTimeout, res.Resolver.Timeout)
	res.Resolver.Command = coalesce(generateArgs.ResolverCommand, res.Resolver.Command)
	res.Render.PreambleTemplate = coalesce(generateArgs.PreambleTemplate, res.Render.PreambleTemplate)
	res.Render.DisableFormatting = coalesce(generateArgs.DisableFormatting, res.Render.DisableFormatting)

	res.Client.GoModTemplate = coalesce(generateArgs.goModTemplate, res.Client.GoModTemplate)

	return res
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

func getRenderOpts(conf toolConfig, targetDir string, findProjectModule bool) (common.RenderOpts, error) {
	logger := log.GetLogger("")
	res := common.RenderOpts{
		RuntimeModule:     conf.RuntimeModule,
		TargetDir:         targetDir,
		DisableFormatting: conf.Render.DisableFormatting,
		PreambleTemplate:  conf.Render.PreambleTemplate,
	}

	// Selections
	for _, item := range conf.Selections {
		sel := common.ConfigSelectionItem{
			Protocols:   item.Protocols,
			ObjectKinds: item.ObjectKinds,
			ModuleURLRe: item.ModuleURLRe,
			PathRe:      item.PathRe,
			NameRe:      item.NameRe,
			Render: common.ConfigSelectionItemRender{
				Template:         item.Render.Template,
				File:             item.Render.File,
				Package:          item.Render.Package,
				Protocols:        item.Render.Protocols,
				ProtoObjectsOnly: item.Render.ProtoObjectsOnly,
			},
			ReusePackagePath:      item.ReusePackagePath,
			AllSupportedProtocols: lo.Keys(asyncapi.ProtocolBuilders),
		}
		logger.Debug("Use selection", "value", sel)
		res.Selections = append(res.Selections, sel)
	}

	// ImportBase
	res.ImportBase = conf.ProjectModule
	if res.ImportBase == "" && findProjectModule {
		m, err := getProjectModule()
		if err != nil {
			return res, fmt.Errorf("read go.mod (use -M arg to override): %w", err)
		}
		logger.Debug("Determined project module", "value", m)
		// Clean target directory path, removing empty, current and parent directories, leaving only the names.
		// This is not the best solution, however, it should work for most cases. Moreover, user can always override it.
		parts := lo.Filter(strings.Split(path.Clean(targetDir), string(os.PathSeparator)), func(s string, _ int) bool {
			return !lo.Contains([]string{"", ".", ".."}, s)
		})
		res.ImportBase = path.Join(m, path.Join(parts...))
	}
	logger.Debug("Import base", "value", res.ImportBase)

	return res, nil
}

func getImplementationOpts(conf toolConfigImplementations) common.RenderImplementationsOpts {
	return common.RenderImplementationsOpts{
		Disable:   conf.Disable,
		Directory: conf.Directory,
		Protocols: lo.Map(conf.Protocols, func(item toolConfigImplementationProtocol, _ int) common.ConfigImplementationProtocol {
			return common.ConfigImplementationProtocol{
				Protocol:         item.Protocol,
				Name:             item.Name,
				Disable:          item.Disable,
				Directory:        item.Directory,
				Package:          item.Package,
				ReusePackagePath: item.ReusePackagePath,
			}
		}),
	}
}

func selectObjects(allObjects []common.CompileObject, selections []common.ConfigSelectionItem) (res []renderer.RenderQueueItem) {
	logger := log.GetLogger(log.LoggerPrefixRendering)

	for _, selection := range selections {
		logger.Debug("Select objects", "selection", selection)
		selectedObjects := selector.SelectObjects(allObjects, selection)
		for _, obj := range selectedObjects {
			logger.Debug("-> Selected", "object", obj)
			res = append(res, renderer.RenderQueueItem{Selection: selection, Object: obj})
		}
	}
	return
}

func getProjectModule() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current working directory: %w", err)
	}
	fn := path.Join(pwd, "go.mod")
	f, err := os.Open(fn)
	if err != nil {
		return "", fmt.Errorf("open %q: %w", fn, err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read %q file: %w", fn, err)
	}
	modpath := modfile.ModulePath(data)
	if modpath == "" {
		return "", fmt.Errorf("reading module name from %q", fn)
	}
	return modpath, nil
}

func getResolver(conf toolConfig) resolver.SpecFileResolver {
	logger := log.GetLogger(log.LoggerPrefixResolving)
	if conf.Resolver.Command != "" {
		return resolver.SubprocessSpecFileResolver{
			CommandLine: conf.Resolver.Command,
			RunTimeout:  conf.Resolver.Timeout,
			Logger:      logger,
		}
	}
	return resolver.DefaultSpecFileResolver{
		Client:  stdHTTP.DefaultClient,
		Timeout: conf.Resolver.Timeout,
		BaseDir: conf.Resolver.SearchDirectory,
		Logger:  logger,
	}
}

func runCompilationLinking(fileResolver resolver.SpecFileResolver, specURL *specurl.URL, compileOpts common.CompileOpts) (map[string]*compiler.Module, error) {
	logger := log.GetLogger("")

	logger.Debug("Run compilation")
	modules, err := runCompilation(specURL, compileOpts, fileResolver)
	if err != nil {
		return nil, err
	}
	logger.Debug("Compilation complete", "files", len(modules))
	objSources := lo.MapValues(modules, func(value *compiler.Module, _ string) linker.ObjectSource { return value })

	logger.Debug("Run linking")
	if err = runLinking(objSources); err != nil {
		return nil, fmt.Errorf("linking: %w", err)
	}
	logger.Debug("Linking complete")
	return modules, nil
}

func runCompilation(specURL *specurl.URL, compileOpts common.CompileOpts, fileResolver resolver.SpecFileResolver) (map[string]*compiler.Module, error) {
	logger := log.GetLogger(log.LoggerPrefixCompilation)
	compileQueue := []*specurl.URL{specURL}      // Queue of specIDs to compile
	modules := make(map[string]*compiler.Module) // Compilers by spec id
	for len(compileQueue) > 0 {
		specURL, compileQueue = compileQueue[0], compileQueue[1:] // Pop an item from queue
		if _, ok := modules[specURL.SpecID]; ok {
			continue // Skip if a spec file has been already compiled
		}

		logger.Info("Compile a spec", "specURL", specURL)
		module := compiler.NewModule(specURL)
		modules[specURL.SpecID] = module

		if !compileOpts.AllowRemoteRefs && specURL.IsRemote() {
			return nil, fmt.Errorf(
				"%s: external requests are forbidden by default for security reasons, use --allow-remote-refs flag to allow them",
				specURL,
			)
		}
		logger.Debug("Loading a spec", "specURL", specURL)
		if err := module.Load(fileResolver); err != nil {
			return nil, fmt.Errorf("load a spec: %w", err)
		}
		logger.Debug("Compilation a spec", "specURL", specURL)
		if err := module.Compile(common.NewCompileContext(specURL, compileOpts)); err != nil {
			return nil, fmt.Errorf("compilation a spec: %w", err)
		}
		logger.Debugf("Compiler stats: %s", module.Stats())
		compileQueue = lo.Flatten([][]*specurl.URL{compileQueue, module.ExternalSpecs()}) // Extend queue with remote specPaths
	}

	return modules, nil
}

func getImplementations(conf common.RenderImplementationsOpts, protocols []string) ([]common.ImplementationObject, error) {
	var res []common.ImplementationObject
	logger := log.GetLogger(log.LoggerPrefixCompilation)

	manifest := lo.Must(loadImplementationsManifest())

	for _, protocol := range protocols {
		protoConf := getImplementationConfig(conf, protocol, manifest)
		if protoConf.Disable {
			logger.Debug("Skip disabled implementation", "protocol", protocol, "name", protoConf.Name)
			continue
		}
		logger.Trace("Compile implementation", "protocol", protocol, "name", protoConf.Name)
		protoManifest, found := lo.Find(manifest, func(item implementations.ImplManifestItem) bool {
			return item.Name == protoConf.Name && item.Protocol == protocol
		})
		if !found {
			return res, fmt.Errorf("cannot find implementation %q for protocol %s", protoConf.Name, protocol)
		}

		res = append(res, common.ImplementationObject{Manifest: protoManifest, Config: protoConf})
	}

	return res, nil
}

func runLinking(objSources map[string]linker.ObjectSource) error {
	logger := log.GetLogger(log.LoggerPrefixLinking)

	// Linking refs
	linker.AssignRefs(objSources)
	danglingRefs := linker.DanglingRefs(objSources)
	logger.Debugf("Linker stats: %s", linker.Stats(objSources))
	if len(danglingRefs) > 0 {
		logger.Error("Some refs remain dangling", "refs", danglingRefs)
		return fmt.Errorf("cannot resolve all refs")
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
	logger.Info("Linking complete", "refs", refsCount)
	return nil
}

func getImplementationConfig(conf common.RenderImplementationsOpts, protocol string, manifest implementations.ImplManifest) common.ConfigImplementationProtocol {
	// Get default implementation
	protoManifest, found := lo.Find(manifest, func(item implementations.ImplManifestItem) bool {
		return item.Default && item.Protocol == protocol
	})
	if !found {
		panic(fmt.Sprintf("cannot find default implementation for protocol %s. This is a bug: %v", protocol, manifest))
	}

	protoConf, _ := lo.Find(conf.Protocols, func(item common.ConfigImplementationProtocol) bool { return item.Protocol == protocol })
	return common.ConfigImplementationProtocol{
		Protocol:         protocol,
		Name:             coalesce(protoConf.Name, protoManifest.Name),
		Disable:          coalesce(protoConf.Disable, conf.Disable),
		Directory:        coalesce(protoConf.Directory, conf.Directory),
		Package:          protoConf.Package,
		ReusePackagePath: protoConf.ReusePackagePath,
	}
}

func loadImplementationsManifest() (implementations.ImplManifest, error) {
	f, err := implementations.ImplementationFS.Open("manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("cannot open manifest.yaml: %w", err)
	}
	dec := yaml.NewDecoder(f)
	var meta implementations.ImplManifest
	if err = dec.Decode(&meta); err != nil {
		return nil, fmt.Errorf("cannot parse manifest.yaml: %w", err)
	}

	return meta, nil
}

func coalesce[T comparable](vals ...T) T {
	res, _ := lo.Coalesce(vals...)
	return res
}
