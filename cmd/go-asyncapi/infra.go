package main

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/renderer"
	"github.com/bdragon300/go-asyncapi/internal/specurl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/templates/infra"
	"github.com/samber/lo"
	"io/fs"
	"os"
	"time"
)

type InfraCmd struct {
	Spec string `arg:"required,positional" help:"AsyncAPI specification file path or url" placeholder:"PATH"`

	Format string `arg:"-f,--format" help:"Output file format" placeholder:"FORMAT"`
	ConfigFile     string `arg:"-c,--config-file" help:"YAML configuration file path" placeholder:"PATH"`
	OutputFile string `arg:"-o,--output" help:"Output file path" placeholder:"PATH"`

	TemplateDir string `arg:"-T,--template-dir" help:"Directory with custom templates" placeholder:"DIR"`
	AllowRemoteRefs bool `arg:"--allow-remote-refs" help:"Allow resolver to fetch the files from remote $ref URLs"`
	ResolverSearchDir   string        `arg:"--resolver-search-dir" help:"Directory to search the local spec files for [default: current working directory]" placeholder:"PATH"`
	ResolverTimeout time.Duration `arg:"--resolver-timeout" help:"Timeout for resolver to resolve a spec file" placeholder:"DURATION"`
	ResolverCommand string        `arg:"--resolver-command" help:"Custom resolver executable to use instead of built-in resolver" placeholder:"PATH"`
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
	fileResolver := getResolver(cmdConfig)
	specURL := specurl.Parse(cmd.Spec)
	compileOpts := common.CompileOpts{
		AllowRemoteRefs:     cmdConfig.Resolver.AllowRemoteReferences,
		RuntimeModule:       cmdConfig.RuntimeModule,
		GeneratePublishers:  true,
		GenerateSubscribers: true,
	}
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

	// TODO: refactor RenderOpts -- it almost not needed here, it's related to codegen.
	//       Also consider to include add ConfigInfraServer (replace RenderOpts to interface in manager?)
	renderOpts, err := getRenderOpts(cmdConfig, cmdConfig.Directories.Target, false)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrongCliArgs, err)
	}
	renderManager := manager.NewTemplateRenderManager(renderOpts)

	// Module objects
	logger.Debug("Run objects rendering")
	templateDirs := []fs.FS{infra.TemplateFS}
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
	// TODO: check if all server variables are set in config, error if not
	serverVariables := toServerConfig(cmdConfig.Infra.Servers)

	if err = renderer.RenderInfra(renderQueue, activeProtocols, cmdConfig.Infra.OutputFile, serverVariables, renderManager); err != nil {
		return fmt.Errorf("render infra: %w", err)
	}

	f, err := os.OpenFile(cmdConfig.Infra.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	defer f.Close()

	states := renderManager.CommittedStates()
	if _, err = f.Write(states[cmdConfig.Infra.OutputFile].Buffer.Bytes()); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync output file: %w", err)
	}
	logger.Infof("Output file saved to %q", cmdConfig.Infra.OutputFile)

	return nil
}

func cliInfraMergeConfig(globalConfig toolConfig, cmd *InfraCmd) (toolConfig, error) {
	res := globalConfig

	res.Infra.Format = coalesce(cmd.Format, globalConfig.Infra.Format)
	var outputFile string
	switch res.Infra.Format {
	case "docker":
		outputFile = "./docker-compose.yaml"
	default:
		return res, fmt.Errorf("unknown file format: %s", cmd.Format)
	}
	res.Infra.OutputFile = coalesce(cmd.OutputFile, outputFile)

	res.Directories.Templates = coalesce(cmd.TemplateDir, globalConfig.Directories.Templates)

	res.Resolver.AllowRemoteReferences = coalesce(cmd.AllowRemoteRefs, globalConfig.Resolver.AllowRemoteReferences)
	res.Resolver.SearchDirectory = coalesce(cmd.ResolverSearchDir, globalConfig.Resolver.SearchDirectory)
	res.Resolver.Timeout = coalesce(cmd.ResolverTimeout, globalConfig.Resolver.Timeout)
	res.Resolver.Command = coalesce(cmd.ResolverCommand, globalConfig.Resolver.Command)

	return res, nil
}

func toServerConfig(servers []toolConfigInfraServer) []common.ConfigInfraServer {
	res := make([]common.ConfigInfraServer, 0)

	for _, server := range servers {
		switch server.Variables.Selector {
		case 0:
			res = append(res, common.ConfigInfraServer{
				Name: server.Name,
				VariableGroups: [][]common.ConfigServerVariable{
					lo.Map(server.Variables.V0.Entries(), func(item lo.Entry[string, string], _ int) common.ConfigServerVariable {
						return common.ConfigServerVariable{Name: item.Key, Value: item.Value}
					}),
				},
			})
		case 1:
			res = append(res, common.ConfigInfraServer{
				Name: server.Name,
				VariableGroups: lo.Map(server.Variables.V1, func(item types.OrderedMap[string, string], _ int) []common.ConfigServerVariable {
					return lo.Map(item.Entries(), func(item lo.Entry[string, string], _ int) common.ConfigServerVariable {
						return common.ConfigServerVariable{Name: item.Key, Value: item.Value}
					})
				}),
			})
		}
	}
	return res
}