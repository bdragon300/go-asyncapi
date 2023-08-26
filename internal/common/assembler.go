package common

import "github.com/dave/jennifer/jen"

type Assembled interface {
	AllowRender() bool
	AssembleDefinition(ctx *AssembleContext) []*jen.Statement
	AssembleUsage(ctx *AssembleContext) []*jen.Statement
}

type AssembleContext struct {
	CurrentPackage     PackageKind
	ImportBase         string
	ForceImportPackage string
	RuntimePackage     string // TODO: replace on package params in appropriate assemble.* struct
}
