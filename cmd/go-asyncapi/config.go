package main

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/assets"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"time"
)

type (
	toolConfigSelectionRender struct {
		Protocols        []string          `yaml:"protocols"`
		ProtoObjectsOnly bool              `yaml:"protoObjectsOnly"`
		Template         string            `yaml:"template"`
		File             string            `yaml:"file"`
		Package          string            `yaml:"package"`
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

	toolConfigDirectories struct {
		Templates string `yaml:"templates"`
		Target    string `yaml:"target"`
		Implementations string `yaml:"implementations"`
	}

	toolConfigRender struct {
		PreambleTemplate string `yaml:"preambleTemplate"`
		DisableFormatting bool `yaml:"disableFormatting"`
	}

	toolConfigResolver struct {
		AllowRemoteReferences bool          `yaml:"allowRemoteReferences"`
		SearchDirectory string        `yaml:"searchDirectory"`
		Timeout time.Duration `yaml:"timeout"`
		Command string        `yaml:"command"`
	}

	toolConfigImplementation struct {
		Name string `yaml:"name"`
	}

	toolConfig struct {
		ConfigVersion int`yaml:"configVersion"`
		ProjectModule string `yaml:"projectModule"`
		RuntimeModule string `yaml:"runtimeModule"`
		Directories toolConfigDirectories `yaml:"directories"`

		Selections        []toolConfigSelection `yaml:"selections"`
		Resolver          toolConfigResolver    `yaml:"resolver"`
		NoImplementations bool                  `yaml:"noImplementations"`
		Render toolConfigRender `yaml:"render"`

		Implementations map[string]toolConfigImplementation `yaml:"implementations"`
	}
)

func loadDefaultConfig() (toolConfig, error) {
	f, err := assets.AssetFS.Open(defaultConfigFileName)
	if err != nil {
		return toolConfig{}, fmt.Errorf("cannot open default config file in assets, this is a bug: %w", err)
	}
	defer f.Close()

	return parseConfigFile(f)
}

func loadConfig(fileName string) (toolConfig, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return toolConfig{}, fmt.Errorf("cannot open config file: %w", err)
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
