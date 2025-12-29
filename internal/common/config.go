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
		Layout                 []ConfigLayoutItem
		UtilCodeOpts           ConfigUtilCodeOpts
		ImplementationCodeOpts ConfigImplementationCodeOpts
	}

	ConfigLayoutItem struct {
		Protocols        []string
		ArtifactKinds    []string
		ModuleURLRe      string
		PathRe           string
		NameRe           string
		Not              bool
		Render           ConfigLayoutItemRender
		ReusePackagePath string
	}

	ConfigLayoutItemRender struct {
		Template         string
		File             string
		Package          string
		Protocols        []string
		ProtoObjectsOnly bool
	}

	ConfigUtilCodeOpts struct {
		Directory string
	}

	ConfigImplementationCodeOpts struct {
		Directory string // Template expression, relative to the target directory
		Disable   bool
		Overrides []ConfigImplementationProtocol
	}

	ConfigImplementationProtocol struct {
		Protocol          string
		Name              string
		Disable           bool
		TemplateDirectory string
		Package           string
		ReusePackagePath  string
	}

	ConfigInfraServerOpt struct {
		ServerName     string
		VariableGroups [][]ConfigServerVariable
	}

	ConfigServerVariable struct {
		Name  string
		Value string
	}

	ConfigDiagram struct {
		ShowChannels        bool
		ShowServers         bool
		ShowDocumentBorders bool
		D2DiagramDirection  D2DiagramDirection
	}

	ConfigUI struct {
		ReferenceOnly bool
	}
)

func (r ConfigLayoutItem) AppliedToProtocol(proto string) bool {
	if len(r.Render.Protocols) > 0 {
		return lo.Contains(r.Render.Protocols, proto)
	}
	return true
}
