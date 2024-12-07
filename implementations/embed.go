package implementations

import (
	"embed"
)

//go:embed *
var ImplementationsFS embed.FS

type ImplManifestItem struct {
	URL string `json:"url"`
	Dir string `json:"dir"`
}

type ImplManifest map[string]map[string]ImplManifestItem
