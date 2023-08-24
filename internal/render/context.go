package render

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type LangRenderer interface {
	AllowRender() bool
	RenderDefinition(ctx *Context) []*jen.Statement
	RenderUsage(ctx *Context) []*jen.Statement
	AdditionalImports() map[string]string
}

type Context struct {
	CurrentPackage     common.PackageKind
	ImportBase         string
	ForceImportPackage string
}
