package common

import (
	"path"

	"github.com/dave/jennifer/jen"
)

type Renderer interface {
	AllowRender() bool
	RenderDefinition(ctx *RenderContext) []*jen.Statement
	RenderUsage(ctx *RenderContext) []*jen.Statement
	String() string
}

type RenderContext struct {
	CurrentPackage string
	ImportBase     string
}

func (a RenderContext) RuntimePackage(subPackage string) string {
	return path.Join(RunPackagePath, subPackage)
}

func (a RenderContext) GeneratedPackage(subPackage string) string {
	return path.Join(a.ImportBase, subPackage)
}
