package assemble

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

func NewLinkPathQuery[T any](pkg common.PackageKind, path []string) *LinkQuery[T] {
	return &LinkQuery[T]{
		pkg:  pkg,
		path: path,
	}
}

func NewLinkRefQuery[T any](pkg common.PackageKind, ref string) *LinkQuery[T] {
	return &LinkQuery[T]{
		pkg: pkg,
		ref: ref,
	}
}

type LinkQuery[T any] struct {
	pkg  common.PackageKind
	path []string
	ref  string
	link T
}

func (r *LinkQuery[T]) Assign(obj any) {
	r.link = obj.(T)
}

func (r *LinkQuery[T]) Link() T {
	return r.link
}

func (r *LinkQuery[T]) Package() common.PackageKind {
	return r.pkg
}

func (r *LinkQuery[T]) Path() []string {
	return r.path
}

func (r *LinkQuery[T]) Ref() string {
	return r.ref
}

func NewLinkQueryList[T any](pkg common.PackageKind, path []string) *LinkQueryList[T] {
	return &LinkQueryList[T]{
		pkg:  pkg,
		path: path,
	}
}

type LinkQueryList[T any] struct {
	pkg   common.PackageKind
	path  []string
	links []T
}

func (r *LinkQueryList[T]) AssignList(obj []any) {
	r.links = utils.CastSliceItems[any, T](obj)
}

func (r *LinkQueryList[T]) Links() []T {
	return r.links
}

func (r *LinkQueryList[T]) Package() common.PackageKind {
	return r.pkg
}

func (r *LinkQueryList[T]) Path() []string {
	return r.path
}

func NewLinkQueryRendererPath(pkg common.PackageKind, path []string) *LinkQueryRenderer {
	return &LinkQueryRenderer{
		LinkQuery: *NewLinkPathQuery[common.Assembled](pkg, path),
	}
}

func NewLinkQueryRendererRef(pkg common.PackageKind, ref string) *LinkQueryRenderer {
	return &LinkQueryRenderer{
		LinkQuery: *NewLinkRefQuery[common.Assembled](pkg, ref),
	}
}

type LinkQueryRenderer struct {
	LinkQuery[common.Assembled]
}

func (r *LinkQueryRenderer) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return r.link.AssembleDefinition(ctx)
}

func (r *LinkQueryRenderer) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if r.pkg != "" && ctx.CurrentPackage != r.pkg {
		t := ctx.ForceImportPackage
		ctx.ForceImportPackage = path.Join(ctx.ImportBase, string(r.pkg))
		defer func() { ctx.ForceImportPackage = t }()
	}
	return r.link.AssembleUsage(ctx)
}

func (r *LinkQueryRenderer) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func NewLinkQueryTypePath(pkg common.PackageKind, path []string) *LinkQueryType {
	return &LinkQueryType{
		LinkQuery: *NewLinkPathQuery[common.GolangType](pkg, path),
	}
}

func NewLinkQueryTypeRef(pkg common.PackageKind, ref string) *LinkQueryType {
	return &LinkQueryType{
		LinkQuery: *NewLinkRefQuery[common.GolangType](pkg, ref),
	}
}

type LinkQueryType struct {
	LinkQuery[common.GolangType]
}

func (r *LinkQueryType) TypeName() string {
	return r.link.TypeName()
}

func (r *LinkQueryType) CanBePointer() bool {
	return r.link.CanBePointer()
}

func (r *LinkQueryType) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *LinkQueryType) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return r.link.AssembleDefinition(ctx)
}

func (r *LinkQueryType) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if r.pkg != "" && ctx.CurrentPackage != r.pkg {
		t := ctx.ForceImportPackage
		ctx.ForceImportPackage = ctx.ImportBase + "/" + string(r.pkg)
		defer func() { ctx.ForceImportPackage = t }()
	}
	return r.link.AssembleUsage(ctx)
}
