package lang

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

func NewLinkerPathQuery[T any](pkg common.PackageKind, path []string) *LinkerQuery[T] {
	return &LinkerQuery[T]{
		pkg:  pkg,
		path: path,
	}
}

func NewLinkerRefQuery[T any](pkg common.PackageKind, ref string) *LinkerQuery[T] {
	return &LinkerQuery[T]{
		pkg: pkg,
		ref: ref,
	}
}

type LinkerQuery[T any] struct {
	pkg  common.PackageKind
	path []string
	ref  string
	link T
}

func (r *LinkerQuery[T]) Assign(obj any) {
	r.link = obj.(T)
}

func (r *LinkerQuery[T]) Link() T {
	return r.link
}

func (r *LinkerQuery[T]) Package() common.PackageKind {
	return r.pkg
}

func (r *LinkerQuery[T]) Path() []string {
	return r.path
}

func (r *LinkerQuery[T]) Ref() string {
	return r.ref
}

func NewLinkerQueryList[T any](pkg common.PackageKind, path []string) *LinkerQueryList[T] {
	return &LinkerQueryList[T]{
		pkg:  pkg,
		path: path,
	}
}

type LinkerQueryList[T any] struct {
	pkg   common.PackageKind
	path  []string
	links []T
}

func (r *LinkerQueryList[T]) AssignList(obj []any) {
	r.links = utils.CastSliceItems[any, T](obj)
}

func (r *LinkerQueryList[T]) Links() []T {
	return r.links
}

func (r *LinkerQueryList[T]) Package() common.PackageKind {
	return r.pkg
}

func (r *LinkerQueryList[T]) Path() []string {
	return r.path
}

func NewLinkerQueryRendererPath(pkg common.PackageKind, path []string) *LinkerQueryRenderer {
	return &LinkerQueryRenderer{
		LinkerQuery: *NewLinkerPathQuery[render.LangRenderer](pkg, path),
	}
}

func NewLinkerQueryRendererRef(pkg common.PackageKind, ref string) *LinkerQueryRenderer {
	return &LinkerQueryRenderer{
		LinkerQuery: *NewLinkerRefQuery[render.LangRenderer](pkg, ref),
	}
}

type LinkerQueryRenderer struct {
	LinkerQuery[render.LangRenderer]
}

func (r *LinkerQueryRenderer) RenderDefinition(ctx *render.Context) []*jen.Statement {
	return r.link.RenderDefinition(ctx)
}

func (r *LinkerQueryRenderer) RenderUsage(ctx *render.Context) []*jen.Statement {
	if r.pkg != "" && ctx.CurrentPackage != r.pkg {
		t := ctx.ForceImportPackage
		ctx.ForceImportPackage = path.Join(ctx.ImportBase, string(r.pkg))
		defer func() { ctx.ForceImportPackage = t }()
	}
	return r.link.RenderUsage(ctx)
}

func (r *LinkerQueryRenderer) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func NewLinkerQueryTypePath(pkg common.PackageKind, path []string) *LinkerQueryType {
	return &LinkerQueryType{
		LinkerQuery: *NewLinkerPathQuery[LangType](pkg, path),
	}
}

func NewLinkerQueryTypeRef(pkg common.PackageKind, ref string) *LinkerQueryType {
	return &LinkerQueryType{
		LinkerQuery: *NewLinkerRefQuery[LangType](pkg, ref),
	}
}

type LinkerQueryType struct {
	LinkerQuery[LangType]
}

func (r *LinkerQueryType) GetName() string {
	return r.link.GetName()
}

func (r *LinkerQueryType) canBePointer() bool {
	return r.link.canBePointer()
}

func (r *LinkerQueryType) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *LinkerQueryType) RenderDefinition(ctx *render.Context) []*jen.Statement {
	return r.link.RenderDefinition(ctx)
}

func (r *LinkerQueryType) RenderUsage(ctx *render.Context) []*jen.Statement {
	if r.pkg != "" && ctx.CurrentPackage != r.pkg {
		t := ctx.ForceImportPackage
		ctx.ForceImportPackage = ctx.ImportBase + "/" + string(r.pkg)
		defer func() { ctx.ForceImportPackage = t }()
	}
	return r.link.RenderUsage(ctx)
}
