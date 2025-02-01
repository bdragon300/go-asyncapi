package tmpl

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
)

// ErrNotDefined is returned when the package location where the object is defined is not known yet.
var ErrNotDefined = errors.New("not defined")

type importsManager interface {
	Imports() []manager.ImportItem
	AddImport(importPath string, pkgName string) string
}

type CodeTemplateContext struct {
	RenderOpts       common.RenderOpts
	CurrentSelection common.ConfigSelectionItem
	PackageName      string
	Object         common.Renderable
	ImportsManager importsManager
}

func (t CodeTemplateContext) Imports() []manager.ImportItem {
	return t.ImportsManager.Imports()
}

type ImplTemplateContext struct {
	Manifest implementations.ImplManifestItem
	Directory string
	Package string
}

type AppTemplateContext struct {
	RenderQueue     []common.Renderable
	ActiveProtocols []string
	ImportsManager  importsManager
}

func (t AppTemplateContext) Imports() []manager.ImportItem {
	return t.ImportsManager.Imports()
}
