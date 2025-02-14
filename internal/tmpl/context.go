package tmpl

import (
	"errors"
	"github.com/bdragon300/go-asyncapi/implementations"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/samber/lo"
)

// ErrNotDefined is returned when the package location where the object is defined is not known yet.
var ErrNotDefined = errors.New("not defined")

type importsManager interface {
	Imports() []manager.ImportItem
}

type CodeTemplateContext struct {
	RenderOpts       common.RenderOpts
	CurrentSelection common.ConfigSelectionItem
	PackageName      string
	Object         common.Renderable
	ImportsManager importsManager
}

type ImplTemplateContext struct {
	Manifest implementations.ImplManifestItem
	Directory string
	Package string
}

type ClientAppTemplateContext struct {
	RenderOpts       common.RenderOpts
	Objects         []common.Renderable
	ActiveProtocols []string
}

type InfraTemplateContext struct {
	ServerConfig []common.ConfigInfraServer
	Objects      []common.Renderable
	ActiveProtocols []string
}

func (i InfraTemplateContext) ServerVariableGroups(serverName string) [][]common.ConfigServerVariable {
	res, ok := lo.Find(i.ServerConfig, func(v common.ConfigInfraServer) bool {
		return v.Name == serverName
	})
	if !ok {
		return nil
	}
	return res.VariableGroups
}