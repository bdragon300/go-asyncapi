package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/selector"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/renderer"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/templates/infra"
	"github.com/samber/lo"
)

type InfraCmd struct {
	Document string `arg:"required,positional" help:"AsyncAPI document file or url" placeholder:"FILE"`

	Engine     string `arg:"-e,--engine" help:"Target infra engine" placeholder:"NAME"`
	OutputFile string `arg:"-o,--output" help:"Output file path or '-' to print to stdout. If omitted, the file name depends on selected engine" placeholder:"FILE"`

	TemplateDir     string        `arg:"-T,--template-dir" help:"User templates directory" placeholder:"DIR"`
	AllowRemoteRefs bool          `arg:"--allow-remote-refs" help:"Allow locator to fetch the files from remote $ref URLs"`
	LocatorRootDir  string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout  time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand  string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`
}

func cliInfra(cmd *InfraCmd, globalConfig toolConfig) error {
	logger := log.GetLogger("")
	cmdConfig, err := cliInfraMergeConfig(globalConfig, cmd)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	//
	// Compilation & linking
	//
	fileLocator := getLocator(cmdConfig)
	docURL, err := jsonpointer.Parse(cmd.Document)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	compileOpts := compile.CompilationOpts{
		AllowRemoteRefs:     cmdConfig.Locator.AllowRemoteReferences,
		GeneratePublishers:  true,
		GenerateSubscribers: true,
	}
	documents, err := runCompilationAndLinking(fileLocator, docURL, compileOpts)
	if err != nil {
		return fmt.Errorf("compilation: %w", err)
	}

	//
	// Rendering
	//
	activeProtocols := collectActiveServersProtocols(documents)
	logger.Debug("Renders protocols", "value", activeProtocols)

	// TODO: refactor RenderOpts -- it almost not needed here, it's related to codegen.
	//       Also consider to include add InfraServerOpts (replace RenderOpts to interface in manager?)
	renderOpts, err := getRenderOpts(cmdConfig, cmdConfig.Code.TargetDir, false)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	renderManager := manager.NewTemplateRenderManager(renderOpts)

	// Document objects
	logger.Debug("Run objects rendering")
	templateDirs := []fs.FS{infra.TemplateFS}
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
	allArtifacts := selector.GatherArtifacts(lo.Values(documents)...)
	visibleArtifacts := lo.Filter(allArtifacts, func(a common.Artifact, _ int) bool {
		return a.Visible()
	})
	logger.Debug("Rendering the artifacts", "allArtifacts", len(allArtifacts), "visibleArtifacts", len(visibleArtifacts))

	// TODO: check if all server variables are set in config, error if not
	serverConfig := getInfraServerConfig(cmdConfig.Infra.ServerOpts)

	if err = renderer.RenderInfra(visibleArtifacts, activeProtocols, cmdConfig.Infra.OutputFile, serverConfig, renderManager); err != nil {
		return fmt.Errorf("render infra: %w", err)
	}

	states := renderManager.CommittedStates()
	outBuf := states[cmdConfig.Infra.OutputFile].Buffer
	if cmdConfig.Infra.OutputFile == "-" {
		logger.Info("Output file to stdout")
		lo.Must(os.Stdout.ReadFrom(outBuf))
		return nil
	}

	return writeToFile(cmdConfig.Infra.OutputFile, outBuf)
}

func writeToFile(fileName string, buf io.Reader) error {
	logger := log.GetLogger("")

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	defer f.Close()

	if _, err = f.ReadFrom(buf); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync output file: %w", err)
	}
	logger.Infof("Output file saved to %q", fileName)
	return nil
}

func cliInfraMergeConfig(globalConfig toolConfig, cmd *InfraCmd) (toolConfig, error) {
	res := globalConfig

	res.Infra.Engine = coalesce(cmd.Engine, globalConfig.Infra.Engine)
	var outputFile string
	switch res.Infra.Engine {
	case "docker":
		outputFile = "./docker-compose.yaml"
	default:
		return res, fmt.Errorf("unknown engine: %s", cmd.Engine)
	}
	res.Infra.OutputFile = coalesce(cmd.OutputFile, outputFile)

	res.Code.TemplatesDir = coalesce(cmd.TemplateDir, globalConfig.Code.TemplatesDir)

	res.Locator.AllowRemoteReferences = coalesce(cmd.AllowRemoteRefs, globalConfig.Locator.AllowRemoteReferences)
	res.Locator.RootDirectory = coalesce(cmd.LocatorRootDir, globalConfig.Locator.RootDirectory)
	res.Locator.Timeout = coalesce(cmd.LocatorTimeout, globalConfig.Locator.Timeout)
	res.Locator.Command = coalesce(cmd.LocatorCommand, globalConfig.Locator.Command)

	return res, nil
}

func getInfraServerConfig(opts []toolConfigInfraServerOpt) []common.InfraServerOpts {
	res := make([]common.InfraServerOpts, 0)

	for _, opt := range opts {
		switch opt.Variables.Selector {
		case 0:
			res = append(res, common.InfraServerOpts{
				ServerName: opt.ServerName,
				VariableGroups: [][]common.InfraServerVariableOpts{
					lo.Map(opt.Variables.V0.Entries(), func(item lo.Entry[string, string], _ int) common.InfraServerVariableOpts {
						return common.InfraServerVariableOpts{Name: item.Key, Value: item.Value}
					}),
				},
			})
		case 1:
			res = append(res, common.InfraServerOpts{
				ServerName: opt.ServerName,
				VariableGroups: lo.Map(opt.Variables.V1, func(item types.OrderedMap[string, string], _ int) []common.InfraServerVariableOpts {
					return lo.Map(item.Entries(), func(item lo.Entry[string, string], _ int) common.InfraServerVariableOpts {
						return common.InfraServerVariableOpts{Name: item.Key, Value: item.Value}
					})
				}),
			})
		}
	}
	return res
}
