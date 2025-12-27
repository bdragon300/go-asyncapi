package codeextra

import (
	"embed"
)

//go:embed *
var TemplateFS embed.FS

type ImplementationManifest struct {
	Protocol string `yaml:"protocol"`
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Dir      string `yaml:"dir"`
	Default  bool   `yaml:"default"`
}

type ImplementationManifests []ImplementationManifest
