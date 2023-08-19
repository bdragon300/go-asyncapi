package templates

import (
	"embed"
)

//go:embed **/*.gotmpl
var Content embed.FS
