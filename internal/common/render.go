package common

import "github.com/bdragon300/go-asyncapi/implementations"

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
