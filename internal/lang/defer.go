package lang

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/dave/jennifer/jen"
)

type DeferRenderer struct {
	Package  common.PackageKind
	RefQuery *scan.RefQuery[render.LangRenderer]
}

func (r DeferRenderer) RenderDefinition(ctx *render.Context) []*jen.Statement {
	return r.RefQuery.Link.RenderDefinition(ctx)
}

func (r DeferRenderer) RenderUsage(ctx *render.Context) []*jen.Statement {
	if ctx.CurrentPackage != r.Package {
		t := ctx.ForceImportPackage
		ctx.ForceImportPackage = ctx.ImportBase + "/" + string(r.Package)
		defer func() { ctx.ForceImportPackage = t }()
	}
	return r.RefQuery.Link.RenderUsage(ctx)
}

func (r DeferRenderer) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r DeferRenderer) AdditionalImports() map[string]string {
	return r.RefQuery.Link.AdditionalImports()
}

type DeferTypeRenderer struct {
	Package  common.PackageKind
	RefQuery *scan.RefQuery[LangType]
}

func (r DeferTypeRenderer) GetName() string {
	return r.RefQuery.Link.GetName()
}

func (r DeferTypeRenderer) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r DeferTypeRenderer) AdditionalImports() map[string]string {
	return r.RefQuery.Link.AdditionalImports()
}

func (r DeferTypeRenderer) RenderDefinition(ctx *render.Context) []*jen.Statement {
	return r.RefQuery.Link.RenderDefinition(ctx)
}

func (r DeferTypeRenderer) RenderUsage(ctx *render.Context) []*jen.Statement {
	if ctx.CurrentPackage != r.Package {
		t := ctx.ForceImportPackage
		ctx.ForceImportPackage = ctx.ImportBase + "/" + string(r.Package)
		defer func() { ctx.ForceImportPackage = t }()
	}
	return r.RefQuery.Link.RenderUsage(ctx)
}

func (r DeferTypeRenderer) canBePointer() bool {
	return r.RefQuery.Link.canBePointer()
}
