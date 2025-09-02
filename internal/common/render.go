package common

import "github.com/bdragon300/go-asyncapi/implementations"

type DiagramOutputFormat string

const (
	DiagramOutputFormatSVG DiagramOutputFormat = "svg"
	DiagramOutputFormatD2  DiagramOutputFormat = "d2"
)

// Rendering options, that come from the configuration file.
type (
	ConfigLayoutItem struct {
		Protocols        []string
		ArtifactKinds    []string
		ModuleURLRe      string
		PathRe           string
		NameRe           string
		Not              bool
		Render           ConfigLayoutItemRender
		ReusePackagePath string

		AllSupportedProtocols []string
	}

	ConfigLayoutItemRender struct {
		Template         string
		File             string
		Package          string
		Protocols        []string
		ProtoObjectsOnly bool
	}

	ConfigImplementationProtocol struct {
		Protocol         string
		Name             string
		Disable          bool
		Directory        string
		Package          string
		ReusePackagePath string
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
	}
)

func (r ConfigLayoutItem) RenderProtocols() []string {
	if len(r.Render.Protocols) > 0 {
		return r.Render.Protocols
	}
	return r.AllSupportedProtocols
}

type RenderOpts struct {
	RuntimeModule    string
	ImportBase       string
	PreambleTemplate string
	Layout           []ConfigLayoutItem
}

type RenderImplementationsOpts struct {
	Disable   bool
	Directory string
	Protocols []ConfigImplementationProtocol
}

type ImplementationObject struct {
	Manifest implementations.ImplManifestItem
	Config   ConfigImplementationProtocol
}
