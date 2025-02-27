package common

import "github.com/bdragon300/go-asyncapi/implementations"

// Rendering options, that come from the configuration file.
type (
	ConfigSelectionItem struct {
		Protocols        []string
		ArtifactKinds    []string
		ModuleURLRe      string
		PathRe           string
		NameRe           string
		Render           ConfigSelectionItemRender
		ReusePackagePath string

		AllSupportedProtocols []string
	}

	ConfigSelectionItemRender struct {
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

	ConfigInfraServer struct {
		Name           string
		VariableGroups [][]ConfigServerVariable
	}

	ConfigServerVariable struct {
		Name  string
		Value string
	}
)

func (r ConfigSelectionItem) RenderProtocols() []string {
	if len(r.Render.Protocols) > 0 {
		return r.Render.Protocols
	}
	return r.AllSupportedProtocols
}

type RenderOpts struct {
	RuntimeModule     string
	ImportBase        string
	TargetDir         string
	DisableFormatting bool
	PreambleTemplate  string
	Selections        []ConfigSelectionItem
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
