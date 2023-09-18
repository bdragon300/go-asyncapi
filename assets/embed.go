package assets

import "embed"

//go:embed runtime/3rd/uritemplates/*.go
var AssetsFS embed.FS
