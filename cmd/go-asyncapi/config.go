package main

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/assets"
	"gopkg.in/yaml.v3"
	"io"
	"os"
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

	toolConfig struct {
		Selections []toolConfigSelection `yaml:"selections"`
	}
)

func loadConfig(fileName string) (toolConfig, error) {
	var conf toolConfig

	var f io.ReadCloser
	var err error
	if fileName == "" {
		f, err = assets.AssetFS.Open(defaultConfigFileName)
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

