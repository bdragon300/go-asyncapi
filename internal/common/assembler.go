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
	CurrentPackage string
	ImportBase     string
}

func (a AssembleContext) RuntimePackage(subPackage string) string {
	return path.Join(RunPackagePath, subPackage)
}

func (a AssembleContext) GeneratedPackage(subPackage string) string {
	return path.Join(a.ImportBase, subPackage)
}
