package implementations

import (
	"embed"
)

//go:embed *
var ImplementationFS embed.FS

type ImplManifestItem struct {
	Protocol string `yaml:"protocol"`
	Name string `yaml:"name"`
	URL string `yaml:"url"`
	Dir string `yaml:"dir"`
	Default bool `yaml:"default"`
}

type ImplManifest []ImplManifestItem
