package tmpl

import (
	"errors"
	"iter"

	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

// ErrNotPinned is occurred when the object has not been pinned to any output file prior to importing it in template code.
// Without pinning we don't know the package where to import this object from.
var ErrNotPinned = errors.New("not pinned or declared before usage")

type importsManager interface {
	Imports() []manager.ImportItem
}

// CodeTemplateContext is a context that is passed to the code templates.
type CodeTemplateContext struct {
	// RenderOpts is the render options. Comes from tool's config and cli flags.
	RenderOpts common.RenderOpts
	// Object is the current object to render.
	Object common.Artifact
	// CurrentLayoutItem is the layout config item that is used to select the Object.
	CurrentLayoutItem common.ConfigLayoutItem
	// PackageName is the package name of the current file.
	PackageName string
	// ImportsManager keeps the imports list for the current file.
	ImportsManager importsManager
}

// ImplTemplateContext is a context that is passed to the implementation templates.
type ImplTemplateContext struct {
	// Manifest implementations.ImplManifestItem object describes the rendering implementation.
	Manifest implementations.ImplManifestItem
	// Directory is the directory (related to target directory) where the implementation for a particular protocol should be placed.
	Directory string
	// Package is the package name for the implementation.
	Package string
}

// ClientAppTemplateContext is a context that is passed to the client application templates.
type ClientAppTemplateContext struct {
	// RenderOpts is the render options. Comes from tool's config and cli flags.
	RenderOpts common.RenderOpts
	// Objects is rendering objects queue.
	Objects []common.Artifact
	// ActiveProtocols is a list of supported protocols, that are used in AsyncAPI document.
	ActiveProtocols []string
}

// InfraTemplateContext is a context that is passed to the infrastructure templates.
type InfraTemplateContext struct {
	// ServerConfig is servers tool config for infra generation process
	ServerConfig []common.ConfigInfraServerOpt
	// Objects is rendering objects queue.
	Objects []common.Artifact
	// ActiveProtocols is a list of supported protocols, that are used in AsyncAPI document.
	ActiveProtocols []string
}

// ServerVariableGroups returns the server variables that are set in tool's config filtered by the server name.
func (i InfraTemplateContext) ServerVariableGroups(serverName string) [][]common.ConfigServerVariable {
	res, ok := lo.Find(i.ServerConfig, func(v common.ConfigInfraServerOpt) bool {
		return v.ServerName == serverName
	})
	if !ok {
		return nil
	}
	return res.VariableGroups
}

// DiagramTemplateContext is a context that is passed to the diagram templates.
type DiagramTemplateContext struct {
	// Objects is rendering objects queue.
	Objects []common.Artifact

	Config common.ConfigDiagram
}

func (d DiagramTemplateContext) ObjectsGroupedByLocation() iter.Seq2[string, []common.Artifact] {
	groups := lo.GroupBy(d.Objects, func(item common.Artifact) string {
		return item.Pointer().Location()
	})
	return utils.OrderedKeysIter(groups)
}
