package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/locator"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/selector"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/writer"
	"github.com/bdragon300/go-asyncapi/templates/client"
	templates "github.com/bdragon300/go-asyncapi/templates/code"
	"github.com/bdragon300/go-asyncapi/templates/codeextra"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

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

	TemplateDir            string `arg:"-T,--template-dir" help:"User templates directory (excepting the util and implementation code, see config file reference)" placeholder:"DIR"`
	PreambleTemplate       string `arg:"--preamble-template" help:"Preamble template name" placeholder:"NAME"`
	DisableFormatting      bool   `arg:"--disable-formatting" help:"Disable code formatting"`
	DisableImplementations bool   `arg:"--disable-implementations" help:"Do not generate implementations code"`

	AllowRemoteRefs bool          `arg:"--allow-remote-refs" help:"Allow locator to fetch the documents from remote hosts"`
	LocatorRootDir  string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout  time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand  string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`

	ClientApp     bool   `arg:"--client-app" help:"Generate the sample client application code as well"`
	goModTemplate string `arg:"-"`
}

func cliCode(cmd *CodeCmd, globalConfig toolConfig) error {
	logger := log.GetLogger("")
	cmdConfig := cliCodeMergeConfig(globalConfig, cmd)

	if logger.GetLevel() == log.TraceLevel {
		buf := lo.Must(yaml.Marshal(cmdConfig))
		logger.Trace("Use the resulting config", "value", string(buf))
	}

	compileOpts := getCompileOpts(cmdConfig)
	renderOpts, err := getRenderOpts(cmdConfig, cmdConfig.Code.TargetDir, true)
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
	documents, err := runCompilationAndLinking(fileLocator, rootDocumentURL, compileOpts)
	if err != nil {
		return fmt.Errorf("compilation: %w", err)
	}

	checkArtifacts(documents)

	//
	// Rendering
	//
	activeProtocols := collectAllProtocols(documents)
	logger.Debug("Collected active protocols", "value", activeProtocols)

	// Extra code: utils code
	logger.Debug("Run util code rendering", "protocols", activeProtocols)
	renderManager.TemplateLoader = tmpl.NewTemplateLoader("", codeextra.TemplateFS) // No main template there
	if err := renderer.RenderUtilCode(activeProtocols, renderOpts, renderManager, codeextra.TemplateFS); err != nil {
		return fmt.Errorf("render util code: %w", err)
	}
	logger.Debug("Extra code rendering complete")

	// Extra code: implementations code
	if !renderOpts.ImplementationCodeOpts.Disable {
		activeProtocols = collectActiveServersProtocols(documents)
		logger.Debug("Run implementations code rendering", "protocols", activeProtocols)
		if err = renderer.RenderImplementationCode(activeProtocols, renderOpts, renderManager, codeextra.TemplateFS); err != nil {
			return fmt.Errorf("render implementation code: %w", err)
		}
		logger.Debug("Implementations rendering complete")
	}

	// Document objects
	logger.Debug("Run objects rendering")
	templateDirs := []fs.FS{templates.TemplateFS}
	if cmdConfig.TemplatesDir != "" {
		logger.Debug("Custom templates location", "directory", cmdConfig.TemplatesDir)
		templateDirs = append(templateDirs, os.DirFS(cmdConfig.TemplatesDir))
	}
	tplLoader := tmpl.NewTemplateLoader(defaultMainTemplateName, templateDirs...)
	logger.Trace("Parse templates", "dirs", templateDirs)
	renderManager.TemplateLoader = tplLoader
	if err = tplLoader.ParseRecursive(renderManager); err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}
	allArtifacts := selector.GatherArtifacts(lo.Values(documents)...)
	logger.Debug("Select artifacts")
	renderQueue := selectArtifacts(allArtifacts, renderOpts.Layout)
	logger.Debug("Rendering the artifacts", "allArtifacts", len(allArtifacts), "selectedArtifacts", len(renderQueue))
	if err = renderer.RenderArtifacts(renderQueue, renderManager); err != nil {
		return fmt.Errorf("render artifacts: %w", err)
	}
	logger.Debug("Objects rendering complete")

	// Client app
	if cmd.ClientApp {
		logger.Debug("Run client app rendering")
		templateDirs = []fs.FS{templates.TemplateFS, client.TemplateFS}
		if cmdConfig.TemplatesDir != "" {
			logger.Debug("Custom templates location", "directory", cmdConfig.TemplatesDir)
			templateDirs = append(templateDirs, os.DirFS(cmdConfig.TemplatesDir))
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
	if !cmdConfig.Code.DisableFormatting {
		logger.Debug("Run postprocessing")
		if err = postprocessGoFiles(files); err != nil {
			return fmt.Errorf("formatting: %w", err)
		}
		logger.Debug("Postprocessing complete")
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

// collectActiveServersProtocols returns a list of protocols that are used in servers that are active and selectable, i.e.
// those, which will appear in the generated code. Used to determine which implementations to generate.
func collectActiveServersProtocols(documents map[string]*compiler.Document) []string {
	servers := collectVisibleArtifactsByType[*render.Server](documents)
	r := lo.Uniq(lo.FilterMap(servers, func(obj *render.Server, _ int) (string, bool) {
		return obj.Protocol, obj.Selectable()
	}))
	return r
}

// collectAllProtocols returns a list of all protocols that are used both in bindings and active servers. Used to
// determine which util code to generate.
func collectAllProtocols(documents map[string]*compiler.Document) []string {
	bindingsArtifacts := collectVisibleArtifactsByType[*render.Bindings](documents)
	bindingProtocols := lo.FlatMap(bindingsArtifacts, func(obj *render.Bindings, _ int) []string {
		// Bindings are always non-selectable, so we don't check Selectable() here
		return obj.Protocols()
	})
	r := lo.Uniq(append(bindingProtocols, collectActiveServersProtocols(documents)...))
	return r
}

func cliCodeMergeConfig(globalConfig toolConfig, cmd *CodeCmd) toolConfig {
	res := globalConfig

	res.ProjectModule = coalesce(cmd.ProjectModule, res.ProjectModule)
	res.RuntimeModule = coalesce(cmd.RuntimeModule, res.RuntimeModule)
	res.TemplatesDir = coalesce(cmd.TemplateDir, res.TemplatesDir)

	res.Locator.AllowRemoteReferences = coalesce(cmd.AllowRemoteRefs, res.Locator.AllowRemoteReferences)
	res.Locator.RootDirectory = coalesce(cmd.LocatorRootDir, res.Locator.RootDirectory)
	res.Locator.Timeout = coalesce(cmd.LocatorTimeout, res.Locator.Timeout)
	res.Locator.Command = coalesce(cmd.LocatorCommand, res.Locator.Command)

	res.Code.OnlyPublish = coalesce(cmd.OnlyPub, res.Code.OnlyPublish)
	res.Code.OnlySubscribe = coalesce(cmd.OnlySub, res.Code.OnlySubscribe)
	res.Code.TargetDir = coalesce(cmd.TargetDir, res.Code.TargetDir)
	res.Code.PreambleTemplate = coalesce(cmd.PreambleTemplate, res.Code.PreambleTemplate)
	res.Code.DisableFormatting = coalesce(cmd.DisableFormatting, res.Code.DisableFormatting)

	res.Code.Implementation.Disable = coalesce(cmd.DisableImplementations, res.Code.Implementation.Disable)

	res.Client.GoModTemplate = coalesce(cmd.goModTemplate, res.Client.GoModTemplate)

	return res
}

func getCompileOpts(cfg toolConfig) compile.CompilationOpts {
	isPub := cfg.Code.OnlyPublish || !cfg.Code.OnlySubscribe
	isSub := cfg.Code.OnlySubscribe || !cfg.Code.OnlyPublish
	return compile.CompilationOpts{
		AllowRemoteRefs:     cfg.Locator.AllowRemoteReferences,
		GeneratePublishers:  isPub,
		GenerateSubscribers: isSub,
	}
}

func getRenderOpts(conf toolConfig, targetDir string, findProjectModule bool) (common.RenderOpts, error) {
	logger := log.GetLogger("")
	res := common.RenderOpts{
		RuntimeModule:    conf.RuntimeModule,
		PreambleTemplate: conf.Code.PreambleTemplate,
		UtilCodeOpts: common.UtilCodeOpts{
			Directory: conf.Code.Util.Directory,
			Custom: lo.Map(conf.Code.Util.Custom, func(item toolConfigCodeUtilProtocol, _ int) common.UtilCodeCustomOpts {
				return common.UtilCodeCustomOpts{
					Protocol:          item.Protocol,
					TemplateDirectory: item.TemplateDirectory,
				}
			}),
		},
		ImplementationCodeOpts: common.ImplementationCodeOpts{
			Directory: conf.Code.Implementation.Directory,
			Disable:   conf.Code.Implementation.Disable,
			Custom: lo.Map(conf.Code.Implementation.Custom, func(item toolConfigImplementationProtocol, _ int) common.ImplementationCodeCustomOpts {
				return common.ImplementationCodeCustomOpts{
					Protocol:          item.Protocol,
					Name:              item.Name,
					Disable:           item.Disable,
					TemplateDirectory: item.TemplateDirectory,
					Package:           item.Package,
				}
			}),
		},
	}

	// Layout
	for _, item := range conf.Code.Layout {
		l := common.CodeLayoutItemOpts{
			Protocols:     item.Protocols,
			ArtifactKinds: item.ArtifactKinds,
			ModuleURLRe:   item.ModuleURLRe,
			PathRe:        item.PathRe,
			NameRe:        item.NameRe,
			Not:           item.Not,
			Render: common.CodeLayoutItemRenderOpts{
				Template:  item.Render.Template,
				File:      item.Render.File,
				Package:   item.Render.Package,
				Protocols: item.Render.Protocols,
			},
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

// selectArtifacts selects artifacts from the list of all artifacts based on the layout configuration.
func selectArtifacts(artifacts []common.Artifact, layout []common.CodeLayoutItemOpts) (res []renderer.RenderQueueItem) {
	logger := log.GetLogger("")

	for _, l := range layout {
		logger.Trace("-> Process layout filters", "item", l)
		selected := selector.ApplyFilters(artifacts, l)
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
	ResolveURL(base, target *jsonpointer.JSONPointer) (*jsonpointer.JSONPointer, error)
}

func getLocator(conf toolConfig) documentLocator {
	logger := log.GetLogger(log.LoggerPrefixLocating)
	if conf.Locator.Command != "" {
		return locator.Subprocess{
			CommandLine:     conf.Locator.Command,
			RunTimeout:      conf.Locator.Timeout,
			ShutdownTimeout: defaultSubprocessLocatorShutdownTimeout,
			RootDirectory:   conf.Locator.RootDirectory,
			Logger:          logger,
		}
	}
	res := locator.Default{
		Client:        &http.Client{Timeout: conf.Locator.Timeout},
		RootDirectory: conf.Locator.RootDirectory,
		Logger:        logger,
	}
	return res
}

func runCompilationAndLinking(
	locator documentLocator,
	docURL *jsonpointer.JSONPointer,
	compileOpts compile.CompilationOpts,
) (map[string]*compiler.Document, error) {
	logger := log.GetLogger("")

	logger.Debug("Run compilation")
	compileContext := compile.NewCompileContext(compileOpts)
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

		// Resolve and add external URLs to the compile queue
		var externalURLs []*jsonpointer.JSONPointer
		for _, u := range document.ExternalURLs() {
			joined, err := locator.ResolveURL(docURL, u)
			if err != nil {
				return nil, fmt.Errorf("join base %q and target %q: %w", docURL, u, err)
			}
			externalURLs = append(externalURLs, joined)
			logger.Trace("Resolved external document location for $ref", "ref", u.String(), "url", joined.Location())
		}
		compileQueue = append(compileQueue, externalURLs...)
	}

	return documents, nil
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

// checkArtifacts briefly checks for the common mistakes in documents, that can lead to incorrect code generation or runtime errors.
// The main purpose of this function is to inform the user about this.
func checkArtifacts(documents map[string]*compiler.Document) {
	logger := log.GetLogger(log.LoggerPrefixLinking)

	// Servers, channels and operations have the common names
	artifacts := lo.Flatten([][]common.Artifact{
		lo.Map(collectVisibleArtifactsByType[*render.Server](documents), func(v *render.Server, _ int) common.Artifact { return v }),
		lo.Map(collectVisibleArtifactsByType[*render.Channel](documents), func(v *render.Channel, _ int) common.Artifact { return v }),
		lo.Map(collectVisibleArtifactsByType[*render.Operation](documents), func(v *render.Operation, _ int) common.Artifact { return v }),
	})
	duplications := lo.FindDuplicatesBy(artifacts, func(item common.Artifact) string {
		return item.Name()
	})
	if len(duplications) > 0 {
		logger.Warn("Some servers, channels or operations have common names. The generated code may contain errors", "names", duplications)
	}

	// Messages in operation is not a subset of channel messages.
	// And Messages in operation reply is not a subset of its channel messages
	operations := collectVisibleArtifactsByType[*render.Operation](documents)
	for _, op := range operations {
		if !lo.Every(op.Channel().Messages(), op.Messages()) {
			logger.Warn("Messages list in Operation is not a subset of Messages list in the Operation's Channel. The generated code may contain errors", "operation", op.Pointer(), "channel", op.Channel().Pointer())
		}

		if op.OperationReply() == nil {
			continue
		}
		ch := op.Channel()
		if op.OperationReply().Channel() != nil {
			ch = op.OperationReply().Channel()
		}
		if !lo.Every(ch.Messages(), op.OperationReply().Messages()) {
			logger.Warn("Messages list in OperationReply is not a subset of Messages list in the OperationReply's Channel. The generated code may contain errors", "operationReply", op.OperationReply().Pointer(), "channel", ch.Pointer())
		}
	}
}

func collectVisibleArtifactsByType[T common.Artifact](documents map[string]*compiler.Document) []T {
	artifacts := lo.FlatMap(maps.Values(documents), func(doc *compiler.Document, _ int) []common.Artifact {
		return doc.Artifacts()
	})
	r := lo.FilterMap(artifacts, func(obj common.Artifact, _ int) (res T, k bool) {
		if v, ok := obj.(T); ok && v.Visible() {
			return v, true
		}
		return res, false
	})
	// Sort by name to keep idempotency
	slices.SortStableFunc(r, func(a, b T) int {
		return strings.Compare(a.Name(), b.Name())
	})
	return r
}

// postprocessGoFiles formats the file buffers in-place applying go fmt.
func postprocessGoFiles(files map[string]*bytes.Buffer) error {
	logger := log.GetLogger(log.LoggerPrefixFormatting)

	keys := lo.Keys(files)
	slices.Sort(keys)
	for _, fileName := range keys {
		if !strings.HasSuffix(fileName, ".go") {
			logger.Debug("Skip a file", "name", fileName)
			continue
		}
		buf := files[fileName]
		logger.Debug("File", "name", fileName, "bytes", buf.Len())
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return types.MultilineError{err, buf.Bytes()}
		}
		buf.Reset()
		buf.Write(formatted)
		logger.Debug("-> File formatted", "name", fileName, "bytes", buf.Len())
	}

	logger.Info("Formatting complete", "files", len(files))
	return nil
}

// coalesce return the first non-zero value from the list of arguments.
func coalesce[T comparable](vals ...T) T {
	res, _ := lo.Coalesce(vals...)
	return res
}
