package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"time"
)

type (
	toolConfig struct {
		ConfigVersion int`yaml:"configVersion"`
		ProjectModule string `yaml:"projectModule"`
		RuntimeModule string `yaml:"runtimeModule"`
		Directories toolConfigDirectories `yaml:"directories"`

		Selections        []toolConfigSelection `yaml:"selections"`
		Resolver          toolConfigResolver    `yaml:"resolver"`
		Render toolConfigRender `yaml:"render"`

		Implementations toolConfigImplementations `yaml:"implementations"`
		Client toolConfigClient `yaml:"client"`
	}

	toolConfigDirectories struct {
		Templates string `yaml:"templates"`
		Target    string `yaml:"target"`
	}

	toolConfigSelection struct {
		NameRe		   string            `yaml:"nameRe"`
		ObjectKinds     []string            `yaml:"objectKinds"`
		ModuleURLRe      string            `yaml:"moduleURLRe"`
		PathRe           string            `yaml:"pathRe"`
		Protocols        []string          `yaml:"protocols"`
		Render 		 toolConfigSelectionRender `yaml:"render"`
		ReusePackagePath string `yaml:"reusePackagePath"`
	}

	toolConfigSelectionRender struct {
		Protocols        []string          `yaml:"protocols"`
		ProtoObjectsOnly bool              `yaml:"protoObjectsOnly"`
		Template         string            `yaml:"template"`
		File             string            `yaml:"file"`
		Package          string            `yaml:"package"`
	}

	toolConfigResolver struct {
		AllowRemoteReferences bool          `yaml:"allowRemoteReferences"`
		SearchDirectory string        `yaml:"searchDirectory"`
		Timeout time.Duration `yaml:"timeout"`
		Command string        `yaml:"command"`
	}

	toolConfigRender struct {
		PreambleTemplate string `yaml:"preambleTemplate"`
		DisableFormatting bool `yaml:"disableFormatting"`
	}

	toolConfigImplementations struct {
		Disable bool `yaml:"disable"`
		Directory string `yaml:"directory"` // Template expression, relative to the target directory
		Protocols []toolConfigImplementationProtocol `yaml:"protocols"`
	}

	toolConfigImplementationProtocol struct {
		Protocol  string `yaml:"protocol"`
		Name      string `yaml:"name"`
		Disable   bool   `yaml:"disable"`
		Directory string `yaml:"directory"` // Template expression, relative to the target directory
		Package string `yaml:"package"`
		ReusePackagePath string `yaml:"reusePackagePath"`
	}

	toolConfigClient struct {
		OutputFile string `yaml:"outputFile"`
		OutputSourceFile string `yaml:"outputSourceFile"`
		KeepSource bool `yaml:"keepSource"`
		GoModTemplate string `yaml:"goModTemplate"`
	}
)

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

func mergeConfig(defaultConf, userConf toolConfig) toolConfig {
	var res toolConfig

	res.ConfigVersion = coalesce(userConf.ConfigVersion, defaultConf.ConfigVersion)
	res.ProjectModule = coalesce(userConf.ProjectModule, defaultConf.ProjectModule)
	res.RuntimeModule = coalesce(userConf.RuntimeModule, defaultConf.RuntimeModule)
	res.Directories.Templates = coalesce(userConf.Directories.Templates, defaultConf.Directories.Templates)
	res.Directories.Target = coalesce(userConf.Directories.Target, defaultConf.Directories.Target)

	// *Replace* selections
	res.Selections = defaultConf.Selections
	if len(userConf.Selections) > 0 {
		res.Selections = userConf.Selections
	}

	res.Resolver.AllowRemoteReferences = coalesce(userConf.Resolver.AllowRemoteReferences, defaultConf.Resolver.AllowRemoteReferences)
	res.Resolver.SearchDirectory = coalesce(userConf.Resolver.SearchDirectory, defaultConf.Resolver.SearchDirectory)
	res.Resolver.Timeout = coalesce(userConf.Resolver.Timeout, defaultConf.Resolver.Timeout)
	res.Resolver.Command = coalesce(userConf.Resolver.Command, defaultConf.Resolver.Command)

	res.Render.PreambleTemplate = coalesce(userConf.Render.PreambleTemplate, defaultConf.Render.PreambleTemplate)
	res.Render.DisableFormatting = coalesce(userConf.Render.DisableFormatting, defaultConf.Render.DisableFormatting)

	res.Implementations.Disable = coalesce(userConf.Implementations.Disable, defaultConf.Implementations.Disable)
	res.Implementations.Directory = coalesce(userConf.Implementations.Directory, defaultConf.Implementations.Directory)
	res.Implementations.Protocols = defaultConf.Implementations.Protocols
	// *Replace* implementations.protocols
	if len(userConf.Implementations.Protocols) > 0 {
		res.Implementations.Protocols = userConf.Implementations.Protocols
	}

	res.Client.GoModTemplate = coalesce(userConf.Client.GoModTemplate, defaultConf.Client.GoModTemplate)
	res.Client.OutputFile = coalesce(userConf.Client.OutputFile, defaultConf.Client.OutputFile)
	res.Client.OutputSourceFile = coalesce(userConf.Client.OutputSourceFile, defaultConf.Client.OutputSourceFile)
	res.Client.KeepSource = coalesce(userConf.Client.KeepSource, defaultConf.Client.KeepSource)

	return res
}