package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/locator"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/writer"
	"github.com/bdragon300/go-asyncapi/templates/client"
	templates "github.com/bdragon300/go-asyncapi/templates/code"
	"gopkg.in/yaml.v3"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/tcp"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/udp"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/ip"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/redis"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/ws"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi/mqtt"

	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/amqp"
	asyncapiHTTP "github.com/bdragon300/go-asyncapi/internal/asyncapi/http"
	"github.com/bdragon300/go-asyncapi/internal/asyncapi/kafka"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/linker"
	"github.com/bdragon300/go-asyncapi/internal/renderer"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
)

type CodeCmd struct {
	Document string `arg:"required,positional" help:"AsyncAPI document file or url" placeholder:"FILE"`

	TargetDir string `arg:"-t,--target-dir" help:"Directory to save the generated code" placeholder:"DIR"`

	OnlyPub bool `arg:"--only-pub" help:"Generate only the publisher code"`
	OnlySub bool `arg:"--only-sub" help:"Generate only the subscriber code"`

	ProjectModule string `arg:"-M,--module" help:"Project module name in the generated code. By default, read get from go.mod in the current working directory" placeholder:"MODULE"`
	RuntimeModule string `arg:"--runtime-module" help:"Runtime module path" placeholder:"MODULE"`

	TemplateDir       string `arg:"-T,--template-dir" help:"User templates directory" placeholder:"DIR"`
	PreambleTemplate  string `arg:"--preamble-template" help:"Preamble template name" placeholder:"NAME"`
	DisableFormatting bool   `arg:"--disable-formatting" help:"Disable code formatting"`

	ImplementationsDir     string `arg:"--implementations-dir" help:"Directory to save the implementations code, counts from target dir" placeholder:"DIR"`
	DisableImplementations bool   `arg:"--disable-implementations" help:"Do not generate implementations code"`

	AllowRemoteRefs  bool          `arg:"--allow-remote-refs" help:"Allow locator to fetch the documents from remote hosts"`
	LocatorSearchDir string        `arg:"--locator-search-dir" help:"Directory to search the documents for [default: current working directory]" placeholder:"PATH"`
	LocatorTimeout   time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand   string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`

	ClientApp     bool   `arg:"--client-app" help:"Generate the sample client application code as well"`
	goModTemplate string `arg:"-"`
}

// protocolBuilders is a list of protocol-specific artifact builders. This list determines the global list of
// protocols supported by go-asyncapi.
var protocolBuilders = []compile.ProtocolBuilder{
	amqp.ProtoBuilder{},
	asyncapiHTTP.ProtoBuilder{},
	kafka.ProtoBuilder{},
	mqtt.ProtoBuilder{},
	ws.ProtoBuilder{},
	redis.ProtoBuilder{},
	ip.ProtoBuilder{},
	tcp.ProtoBuilder{},
	udp.ProtoBuilder{},
}

func cliCode(cmd *CodeCmd, globalConfig toolConfig) error {
	logger := log.GetLogger("")
	cmdConfig := cliCodeMergeConfig(globalConfig, cmd)

	if logger.GetLevel() == log.TraceLevel {
		buf := lo.Must(yaml.Marshal(cmdConfig))
		logger.Trace("Use the resulting config", "value", string(buf))
	}

	supportedProtocols := lo.Map(protocolBuilders, func(item compile.ProtocolBuilder, _ int) string { return item.Protocol() })
	compileOpts := getCompileOpts(cmdConfig)
	renderOpts, err := getRenderOpts(cmdConfig, cmdConfig.Code.TargetDir, true, supportedProtocols)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	renderManager := manager.NewTemplateRenderManager(renderOpts)

	//
	// Compilation & linking
	//
	fileLocator := getLocator(cmdConfig)
	rootDocumentURL, err := jsonpointer.Parse(cmd.Document)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	documents, err := runCompilationAndLinking(fileLocator, rootDocumentURL, compileOpts, protocolBuilders)
	if err != nil {
		return fmt.Errorf("compilation: %w", err)
	}

	//
	// Rendering
	//
	rootDocument := documents[rootDocumentURL.Location()]
	activeProtocols := collectActiveProtocols(rootDocument.Artifacts())
	logger.Debug("Renders protocols", "value", activeProtocols)

	// Implementations
	implementationOpts := getImplementationOpts(cmdConfig)
	if !implementationOpts.Disable {
		tplLoader := tmpl.NewTemplateLoader(defaultMainTemplateName, implementations.ImplementationFS)
		renderManager.TemplateLoader = tplLoader

		protocols := lo.Intersect(supportedProtocols, activeProtocols)
		if len(protocols) < len(activeProtocols) {
			logger.Warn("Some protocols have no implementations", "protocols", lo.Without(activeProtocols, protocols...))
		}

		// Render only implementations for protocols that are actually used in document
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

	// Document objects
	logger.Debug("Run objects rendering")
	templateDirs := []fs.FS{templates.TemplateFS}
	if cmdConfig.Code.TemplatesDir != "" {
		logger.Debug("Custom templates location", "directory", cmdConfig.Code.TemplatesDir)
		templateDirs = append(templateDirs, os.DirFS(cmdConfig.Code.TemplatesDir))
	}
	tplLoader := tmpl.NewTemplateLoader(defaultMainTemplateName, templateDirs...)
	logger.Trace("Parse templates", "dirs", templateDirs)
	renderManager.TemplateLoader = tplLoader
	if err = tplLoader.ParseRecursive(renderManager); err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}
	allArtifacts := lo.FlatMap(lo.Values(documents), func(m *compiler.Document, _ int) []common.Artifact { return m.Artifacts() })
	logger.Debug("Select artifacts")
	renderQueue := selectArtifacts(allArtifacts, renderOpts.Layout)
	if err = renderer.RenderArtifacts(renderQueue, renderManager); err != nil {
		return fmt.Errorf("render artifacts: %w", err)
	}
	logger.Debug("Objects rendering complete")

	// Client app
	if cmd.ClientApp {
		logger.Debug("Run client app rendering")
		templateDirs = []fs.FS{templates.TemplateFS, client.TemplateFS}
		if cmdConfig.Code.TemplatesDir != "" {
			logger.Debug("Custom templates location", "directory", cmdConfig.Code.TemplatesDir)
			templateDirs = append(templateDirs, os.DirFS(cmdConfig.Code.TemplatesDir))
		}
		tplLoader = tmpl.NewTemplateLoader(defaultMainTemplateName, templateDirs...)
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
	if err = writer.WriteBuffersToFiles(files, cmdConfig.Code.TargetDir); err != nil {
		return fmt.Errorf("writing: %w", err)
	}
	logger.Debug("Writing complete")

	logger.Info("Code generation finished")
	return nil
}

// collectActiveProtocols returns a list of protocols that are used in servers that are active and selectable, i.e.
// those, which will appear in the generated code.
func collectActiveProtocols(allObjects []common.Artifact) []string {
	return lo.Uniq(lo.FilterMap(allObjects, func(obj common.Artifact, _ int) (string, bool) {
		if obj.Kind() != common.ArtifactKindServer && !obj.Selectable() || !obj.Visible() {
			return "", false
		}
		obj2 := common.DerefArtifact(obj)
		v, ok := obj2.(*render.Server)
		if !ok {
			return "", false
		}
		return v.Protocol, true
	}))
}

func cliCodeMergeConfig(globalConfig toolConfig, cmd *CodeCmd) toolConfig {
	res := globalConfig

	res.ProjectModule = coalesce(cmd.ProjectModule, res.ProjectModule)
	res.RuntimeModule = coalesce(cmd.RuntimeModule, res.RuntimeModule)

	res.Locator.AllowRemoteReferences = coalesce(cmd.AllowRemoteRefs, res.Locator.AllowRemoteReferences)
	res.Locator.SearchDirectory = coalesce(cmd.LocatorSearchDir, res.Locator.SearchDirectory)
	res.Locator.Timeout = coalesce(cmd.LocatorTimeout, res.Locator.Timeout)
	res.Locator.Command = coalesce(cmd.LocatorCommand, res.Locator.Command)

	res.Code.OnlyPublish = coalesce(cmd.OnlyPub, res.Code.OnlyPublish)
	res.Code.OnlySubscribe = coalesce(cmd.OnlySub, res.Code.OnlySubscribe)
	res.Code.TemplatesDir = coalesce(cmd.TemplateDir, res.Code.TemplatesDir)
	res.Code.TargetDir = coalesce(cmd.TargetDir, res.Code.TargetDir)
	res.Code.PreambleTemplate = coalesce(cmd.PreambleTemplate, res.Code.PreambleTemplate)
	res.Code.DisableFormatting = coalesce(cmd.DisableFormatting, res.Code.DisableFormatting)
	res.Code.ImplementationsDir = coalesce(cmd.ImplementationsDir, res.Code.ImplementationsDir)
	res.Code.DisableImplementations = coalesce(cmd.DisableImplementations, res.Code.DisableImplementations)

	res.Client.GoModTemplate = coalesce(cmd.goModTemplate, res.Client.GoModTemplate)

	return res
}

func getCompileOpts(cfg toolConfig) compile.CompilationOpts {
	isPub := cfg.Code.OnlyPublish || !cfg.Code.OnlySubscribe
	isSub := cfg.Code.OnlySubscribe || !cfg.Code.OnlyPublish
	return compile.CompilationOpts{
		AllowRemoteRefs:     cfg.Locator.AllowRemoteReferences,
		RuntimeModule:       cfg.RuntimeModule,
		GeneratePublishers:  isPub,
		GenerateSubscribers: isSub,
	}
}

func getRenderOpts(conf toolConfig, targetDir string, findProjectModule bool, allProtocols []string) (common.RenderOpts, error) {
	logger := log.GetLogger("")
	res := common.RenderOpts{
		RuntimeModule:     conf.RuntimeModule,
		TargetDir:         targetDir,
		DisableFormatting: conf.Code.DisableFormatting,
		PreambleTemplate:  conf.Code.PreambleTemplate,
	}

	// Layout
	for _, item := range conf.Layout {
		l := common.ConfigLayoutItem{
			Protocols:     item.Protocols,
			ArtifactKinds: item.ArtifactKinds,
			ModuleURLRe:   item.ModuleURLRe,
			PathRe:        item.PathRe,
			NameRe:        item.NameRe,
			Render: common.ConfigLayoutItemRender{
				Template:         item.Render.Template,
				File:             item.Render.File,
				Package:          item.Render.Package,
				Protocols:        item.Render.Protocols,
				ProtoObjectsOnly: item.Render.ProtoObjectsOnly,
			},
			ReusePackagePath:      item.ReusePackagePath,
			AllSupportedProtocols: allProtocols,
		}
		logger.Debug("Use layout item", "value", l)
		res.Layout = append(res.Layout, l)
	}

	// ImportBase
	res.ImportBase = conf.ProjectModule
	if res.ImportBase == "" && findProjectModule {
		m, err := getProjectModule()
		if err != nil {
			return res, fmt.Errorf("determine the module name (use -M arg to override): %w", err)
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

func getImplementationOpts(conf toolConfig) common.RenderImplementationsOpts {
	return common.RenderImplementationsOpts{
		Disable:   conf.Code.DisableImplementations,
		Directory: conf.Code.ImplementationsDir,
		Protocols: lo.Map(conf.Implementations, func(item toolConfigImplementation, _ int) common.ConfigImplementationProtocol {
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

// selectArtifacts selects artifacts from the list of all artifacts based on the layout configuration.
func selectArtifacts(artifacts []common.Artifact, layout []common.ConfigLayoutItem) (res []renderer.RenderQueueItem) {
	logger := log.GetLogger("")

	for _, l := range layout {
		logger.Trace("-> Process layout", "item", l)
		selected := selector.Select(artifacts, l)
		for _, obj := range selected {
			res = append(res, renderer.RenderQueueItem{LayoutItem: l, Object: obj})
		}
		logger.Debug("-> Selected", "artifacts", len(selected))
	}
	return
}

// getProjectModule returns the module name from the go.mod file in the current working directory.
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

type documentLocator interface {
	Locate(docURL *jsonpointer.JSONPointer) (io.ReadCloser, error)
}

func getLocator(conf toolConfig) documentLocator {
	logger := log.GetLogger(log.LoggerPrefixLocating)
	if conf.Locator.Command != "" {
		return locator.Subprocess{
			CommandLine:     conf.Locator.Command,
			RunTimeout:      conf.Locator.Timeout,
			ShutdownTimeout: defaultSubprocessLocatorShutdownTimeout,
			Logger:          logger,
		}
	}
	res := locator.Default{
		Client:    &http.Client{Timeout: conf.Locator.Timeout},
		Directory: conf.Locator.SearchDirectory,
		Logger:    logger,
	}
	return res
}

func runCompilationAndLinking(
	locator documentLocator,
	docURL *jsonpointer.JSONPointer,
	compileOpts compile.CompilationOpts,
	protoBuilders []compile.ProtocolBuilder,
) (map[string]*compiler.Document, error) {
	logger := log.GetLogger("")

	logger.Debug("Run compilation")
	compileContext := compile.NewCompileContext(compileOpts, protoBuilders)
	documents, err := runCompilation(docURL, compileContext, locator)
	if err != nil {
		return nil, err
	}
	logger.Debug("Compilation complete", "files", len(documents))
	objSources := lo.MapValues(documents, func(value *compiler.Document, _ string) linker.ObjectSource { return value })

	logger.Debug("Run linking")
	if err = runLinking(objSources); err != nil {
		return nil, fmt.Errorf("linking: %w", err)
	}
	logger.Debug("Linking complete")
	return documents, nil
}

func runCompilation(
	docURL *jsonpointer.JSONPointer,
	compileContext *compile.Context,
	locator documentLocator,
) (map[string]*compiler.Document, error) {
	logger := log.GetLogger(log.LoggerPrefixCompilation)
	compileQueue := []*jsonpointer.JSONPointer{docURL} // Queue of document urls to compile
	documents := make(map[string]*compiler.Document)   // Documents by url
	for len(compileQueue) > 0 {
		docURL, compileQueue = compileQueue[0], compileQueue[1:] // Pop an item from queue
		if _, ok := documents[docURL.Location()]; ok {
			continue // Skip if a document has been already compiled
		}

		logger.Info("Compile a document", "url", docURL)
		document := compiler.NewDocument(docURL)
		documents[docURL.Location()] = document

		if !compileContext.CompileOpts.AllowRemoteRefs && docURL.URI != nil {
			return nil, fmt.Errorf(
				"%s: external requests are forbidden by default for security reasons, use --allow-remote-refs flag to allow them",
				docURL,
			)
		}
		logger.Debug("Loading a document", "url", docURL)
		if err := document.Load(locator); err != nil {
			return nil, fmt.Errorf("load a document: %w", err)
		}
		logger.Debug("Compiling a document", "url", docURL)
		if err := document.Compile(compileContext); err != nil {
			return nil, fmt.Errorf("compilation a document: %w", err)
		}
		logger.Debugf("Compiler stats: %s", document.Stats())
		// Add external URLs to the compile queue
		compileQueue = append(compileQueue, document.ExternalURLs()...)
	}

	return documents, nil
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
	linker.ResolvePromises(objSources)
	unresolved := linker.UnresolvedPromises(objSources)
	logger.Debugf("Linker stats: %s", linker.Stats(objSources))
	if len(unresolved) > 0 {
		logger.Error("Some refs remain dangling", "refs", unresolved)
		return fmt.Errorf("cannot resolve all refs")
	}

	// Linking list promises
	logger.Debug("Run linking the list promises")
	linker.ResolveListPromises(objSources)
	unresolvedCount := linker.UnresolvedPromisesCount(objSources)
	logger.Debugf("Linker stats: %s", linker.Stats(objSources))
	if unresolvedCount > 0 {
		logger.Error("Cannot assign internal list promises", "promises", unresolvedCount)
		return fmt.Errorf("cannot finish linking")
	}

	refsCount := lo.SumBy(lo.Values(objSources), func(item linker.ObjectSource) int {
		return lo.CountBy(item.Promises(), func(p common.ObjectPromise) bool {
			return p.Origin() == common.PromiseOriginRef
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

// coalesce return the first non-zero value from the list of arguments.
func coalesce[T comparable](vals ...T) T {
	res, _ := lo.Coalesce(vals...)
	return res
}
