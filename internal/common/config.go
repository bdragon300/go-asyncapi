package common

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
		D2DiagramDirection  D2DiagramDirection
	}
)

func (r ConfigLayoutItem) RenderProtocols() []string {
	if len(r.Render.Protocols) > 0 {
		return r.Render.Protocols
	}
	return r.AllSupportedProtocols
}
