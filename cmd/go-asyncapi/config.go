package main

import (
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"
)

// Structures, that represent the tool's configuration file
type (
	toolConfig struct {
		ConfigVersion int    `yaml:"configVersion"`
		ProjectModule string `yaml:"projectModule"`
		RuntimeModule string `yaml:"runtimeModule"`

		Selections      []toolConfigSelection      `yaml:"selections"`
		Locator         toolConfigLocator          `yaml:"locator"`
		Implementations []toolConfigImplementation `yaml:"implementations"`

		Code   toolConfigCode   `yaml:"code"`
		Client toolConfigClient `yaml:"client"`
		Infra  toolConfigInfra  `yaml:"infra"`
	}

	toolConfigSelection struct {
		NameRe           string                    `yaml:"nameRe"`
		ArtifactKinds    []string                  `yaml:"artifactKinds"`
		ModuleURLRe      string                    `yaml:"moduleURLRe"`
		PathRe           string                    `yaml:"pathRe"`
		Protocols        []string                  `yaml:"protocols"`
		Render           toolConfigSelectionRender `yaml:"render"`
		ReusePackagePath string                    `yaml:"reusePackagePath"`
	}

	toolConfigSelectionRender struct {
		Protocols        []string `yaml:"protocols"`
		ProtoObjectsOnly bool     `yaml:"protoObjectsOnly"`
		Template         string   `yaml:"template"`
		File             string   `yaml:"file"`
		Package          string   `yaml:"package"`
	}

	toolConfigLocator struct {
		AllowRemoteReferences bool          `yaml:"allowRemoteReferences"`
		SearchDirectory       string        `yaml:"searchDirectory"`
		Timeout               time.Duration `yaml:"timeout"`
		Command               string        `yaml:"command"`
	}

	toolConfigCode struct {
		OnlyPublish       bool   `yaml:"onlyPublish"`
		OnlySubscribe     bool   `yaml:"onlySubscribe"`
		PreambleTemplate  string `yaml:"preambleTemplate"`
		DisableFormatting bool   `yaml:"disableFormatting"`
		TemplatesDir      string `yaml:"templatesDir"`
		TargetDir         string `yaml:"targetDir"`

		DisableImplementations bool   `yaml:"disableImplementations"`
		ImplementationsDir     string `yaml:"implementationsDir"` // Template expression, relative to the target directory
	}

	toolConfigImplementation struct {
		Protocol         string `yaml:"protocol"`
		Name             string `yaml:"name"`
		Disable          bool   `yaml:"disable"`
		Directory        string `yaml:"directory"` // Template expression, relative to the target directory
		Package          string `yaml:"package"`
		ReusePackagePath string `yaml:"reusePackagePath"`
	}

	toolConfigClient struct {
		OutputFile       string `yaml:"outputFile"`
		OutputSourceFile string `yaml:"outputSourceFile"`
		KeepSource       bool   `yaml:"keepSource"`
		GoModTemplate    string `yaml:"goModTemplate"`
		TempDir          string `yaml:"tempDir"`
	}

	toolConfigInfra struct {
		Servers    []toolConfigInfraServer `yaml:"servers"`
		Format     string                  `yaml:"format"`
		OutputFile string                  `yaml:"outputFile"`
	}

	toolConfigInfraServer struct {
		Name      string                                                                             `yaml:"name"`
		Variables types.Union2[types.OrderedMap[string, string], []types.OrderedMap[string, string]] `yaml:"variables"`
	}
)

// loadConfig loads and parses the configuration file from the given filesystem.
func loadConfig(filesystem fs.FS, fileName string) (toolConfig, error) {
	f, err := filesystem.Open(fileName)
	if err != nil {
		return toolConfig{}, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	return parseConfigFile(f)
}

func parseConfigFile(f io.Reader) (toolConfig, error) {
	var conf toolConfig

	buf, err := io.ReadAll(f)
	if err != nil {
		return conf, fmt.Errorf("cannot read config file: %w", err)
	}

	if err = yaml.Unmarshal(buf, &conf); err != nil {
		return conf, fmt.Errorf("cannot parse YAML config file: %w", err)
	}
	return conf, nil
}

// mergeConfig merges the default configuration with the user-provided one.
func mergeConfig(defaultConf, userConf toolConfig) toolConfig {
	var res toolConfig

	res.ConfigVersion = coalesce(userConf.ConfigVersion, defaultConf.ConfigVersion)
	res.ProjectModule = coalesce(userConf.ProjectModule, defaultConf.ProjectModule)
	res.RuntimeModule = coalesce(userConf.RuntimeModule, defaultConf.RuntimeModule)
	res.Code.TemplatesDir = coalesce(userConf.Code.TemplatesDir, defaultConf.Code.TemplatesDir)
	res.Code.TargetDir = coalesce(userConf.Code.TargetDir, defaultConf.Code.TargetDir)

	// *Replace* selections
	res.Selections = defaultConf.Selections
	if len(userConf.Selections) > 0 {
		res.Selections = userConf.Selections
	}

	res.Locator.AllowRemoteReferences = coalesce(userConf.Locator.AllowRemoteReferences, defaultConf.Locator.AllowRemoteReferences)
	res.Locator.SearchDirectory = coalesce(userConf.Locator.SearchDirectory, defaultConf.Locator.SearchDirectory)
	res.Locator.Timeout = coalesce(userConf.Locator.Timeout, defaultConf.Locator.Timeout)
	res.Locator.Command = coalesce(userConf.Locator.Command, defaultConf.Locator.Command)

	res.Code.PreambleTemplate = coalesce(userConf.Code.PreambleTemplate, defaultConf.Code.PreambleTemplate)
	res.Code.DisableFormatting = coalesce(userConf.Code.DisableFormatting, defaultConf.Code.DisableFormatting)

	res.Code.DisableImplementations = coalesce(userConf.Code.DisableImplementations, defaultConf.Code.DisableImplementations)
	res.Code.ImplementationsDir = coalesce(userConf.Code.ImplementationsDir, defaultConf.Code.ImplementationsDir)
	res.Implementations = defaultConf.Implementations
	// *Replace* implementations.protocols
	if len(userConf.Implementations) > 0 {
		res.Implementations = userConf.Implementations
	}

	res.Client.GoModTemplate = coalesce(userConf.Client.GoModTemplate, defaultConf.Client.GoModTemplate)
	res.Client.OutputFile = coalesce(userConf.Client.OutputFile, defaultConf.Client.OutputFile)
	res.Client.OutputSourceFile = coalesce(userConf.Client.OutputSourceFile, defaultConf.Client.OutputSourceFile)
	res.Client.KeepSource = coalesce(userConf.Client.KeepSource, defaultConf.Client.KeepSource)

	res.Infra.Format = coalesce(userConf.Infra.Format, defaultConf.Infra.Format)
	res.Infra.OutputFile = coalesce(userConf.Infra.OutputFile, defaultConf.Infra.OutputFile)
	res.Infra.Servers = defaultConf.Infra.Servers
	// *Replace* infra.servers
	if len(userConf.Infra.Servers) > 0 {
		res.Infra.Servers = userConf.Infra.Servers
	}

	return res
}
