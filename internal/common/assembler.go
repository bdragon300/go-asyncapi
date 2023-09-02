package common

import (
	"path"

	"github.com/dave/jennifer/jen"
)

type Assembler interface {
	AllowRender() bool
	AssembleDefinition(ctx *AssembleContext) []*jen.Statement
	AssembleUsage(ctx *AssembleContext) []*jen.Statement
}

type AssembleContext struct {
	CurrentPackage PackageKind
	ImportBase     string
}

func (a AssembleContext) RuntimePackage() string {
	return path.Join(a.ImportBase, "runtime")
}
