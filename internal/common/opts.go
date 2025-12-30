package common

import (
	"github.com/samber/lo"
)

type DiagramOutputFormat string

const (
	DiagramOutputFormatSVG DiagramOutputFormat = "svg"
	DiagramOutputFormatD2  DiagramOutputFormat = "d2"
)

type D2DiagramDirection string

const (
	D2DiagramDirectionUp    D2DiagramDirection = "up"
	D2DiagramDirectionDown  D2DiagramDirection = "down"
	D2DiagramDirectionLeft  D2DiagramDirection = "left"
	D2DiagramDirectionRight D2DiagramDirection = "right"
)

type (
	RenderOpts struct {
		RuntimeModule          string
		ImportBase             string
		PreambleTemplate       string
		Layout                 []LayoutItemOpts
		UtilCodeOpts           UtilCodeOpts
		ImplementationCodeOpts ImplementationCodeOpts
	}

	LayoutItemOpts struct {
		Protocols        []string
		ArtifactKinds    []string
		ModuleURLRe      string
		PathRe           string
		NameRe           string
		Not              bool
		Render           LayoutItemRenderOpts
		ReusePackagePath string
	}

	LayoutItemRenderOpts struct {
		Template         string
		File             string
		Package          string
		Protocols        []string
		ProtoObjectsOnly bool
	}

	UtilCodeOpts struct {
		Directory string
	}

	ImplementationCodeOpts struct {
		Directory  string // Template expression, relative to the target directory
		Disable    bool
		Customized []ImplementationCodeCustomizedOpts
	}

	ImplementationCodeCustomizedOpts struct {
		Protocol          string
		Name              string
		Disable           bool
		TemplateDirectory string
		Package           string
		ReusePackagePath  string
	}

	InfraServerOpts struct {
		ServerName     string
		VariableGroups [][]InfraServerVariableOpts
	}

	InfraServerVariableOpts struct {
		Name  string
		Value string
	}

	DiagramRenderOpts struct {
		ShowChannels        bool
		ShowServers         bool
		ShowDocumentBorders bool
		D2DiagramDirection  D2DiagramDirection
	}

	UIRenderOpts struct {
		ReferenceOnly bool
	}

	// UIHTMLResourceOpts represents a CSS or JS resource for UI rendering
	UIHTMLResourceOpts struct {
		Location      string
		Content       string
		Embed         bool
		FileExtension string
	}
)

func (r LayoutItemOpts) AppliedToProtocol(proto string) bool {
	if len(r.Render.Protocols) > 0 {
		return lo.Contains(r.Render.Protocols, proto)
	}
	return true
}
