package tmpl

import (
	"errors"
	"iter"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/bdragon300/go-asyncapi/templates/codeextra"
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
	CurrentLayoutItem common.LayoutItemOpts
	// PackageName is the package name of the current file.
	PackageName string
	// ImportsManager keeps the imports list for the current file.
	ImportsManager importsManager
}

// CodeExtraTemplateContext is a context that is passed to the codeExtra templates.
type CodeExtraTemplateContext struct {
	RenderOpts common.RenderOpts
	Protocol string
	// Directory is the directory (related to target directory) where the code is placed.
	Directory string
	// PackageName is the package name of the current file.
	PackageName string
	// Manifest is the built-in implementation manifest if we rendering it as implementation code. This field
	// is nil when rendering the user-defined implementation templates or other extra code.
	Manifest *codeextra.ImplementationManifest
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
	ServerConfig []common.InfraServerOpts
	// Objects is rendering objects queue.
	Objects []common.Artifact
	// ActiveProtocols is a list of supported protocols, that are used in AsyncAPI document.
	ActiveProtocols []string
}

// ServerVariableGroups returns the server variables that are set in tool's config filtered by the server name.
func (i InfraTemplateContext) ServerVariableGroups(serverName string) [][]common.InfraServerVariableOpts {
	res, ok := lo.Find(i.ServerConfig, func(v common.InfraServerOpts) bool {
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

	Config common.DiagramRenderOpts
}

func (d DiagramTemplateContext) ObjectsGroupedByLocation() iter.Seq2[string, []common.Artifact] {
	groups := lo.GroupBy(d.Objects, func(item common.Artifact) string {
		return item.Pointer().Location()
	})
	return utils.OrderedKeysIter(groups)
}

// UITemplateContext is a context that is passed to the UI template.
type UITemplateContext struct {
	// DocumentContents is the passed AsyncAPI document parsed to map.
	DocumentContents map[string]any

	// DocumentURL is the JSON Pointer to the passed AsyncAPI document.
	DocumentURL jsonpointer.JSONPointer

	// Resources is a list of resources to include in the generated html. Typically, CSS and JS files.
	Resources []common.UIHTMLResourceOpts

	Config common.UIRenderOpts
}
