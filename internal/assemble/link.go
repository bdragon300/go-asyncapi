package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func NewRefLink[T any](pkg common.PackageKind, ref string) *Link[T] {
	return &Link[T]{
		pkg: pkg,
		ref: ref,
	}
}

func NewCbLink[T any](pkg common.PackageKind, findCb func(item any, path []string) bool) *Link[T] {
	return &Link[T]{
		pkg:    pkg,
		findCb: findCb,
	}
}

type Link[T any] struct {
	pkg    common.PackageKind
	ref    string
	findCb func(item any, path []string) bool

	link T
}

func (r *Link[T]) Assign(obj any) {
	r.link = obj.(T)
}

func (r *Link[T]) FindCallback() func(item any, path []string) bool {
	return r.findCb
}

func (r *Link[T]) Obj() T {
	return r.link
}

func (r *Link[T]) Package() common.PackageKind {
	return r.pkg
}

func (r *Link[T]) Ref() string {
	return r.ref
}

func NewListCbLink[T any](pkg common.PackageKind, findCb func(item any, path []string) bool) *LinkList[T] {
	return &LinkList[T]{
		pkg:    pkg,
		findCb: findCb,
	}
}

type LinkList[T any] struct {
	pkg    common.PackageKind
	findCb func(item any, path []string) bool

	links []T
}

func (r *LinkList[T]) AssignList(obj []any) {
	var ok bool
	r.links, ok = lo.FromAnySlice[T](obj)
	if !ok {
		panic(fmt.Sprintf("Cannot assign slice of %+v to %T", obj, r.links))
	}
}

func (r *LinkList[T]) FindCallback() func(item any, path []string) bool {
	return r.findCb
}

func (r *LinkList[T]) Links() []T {
	return r.links
}

func (r *LinkList[T]) Package() common.PackageKind {
	return r.pkg
}

func NewRefLinkAsAssembler(pkg common.PackageKind, ref string) *LinkAsAssembler {
	return &LinkAsAssembler{
		Link: *NewRefLink[common.Assembler](pkg, ref),
	}
}

type LinkAsAssembler struct {
	Link[common.Assembler]
}

func (r *LinkAsAssembler) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return r.link.AssembleDefinition(ctx)
}

func (r *LinkAsAssembler) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	return r.link.AssembleUsage(ctx)
}

func (r *LinkAsAssembler) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func NewRefLinkAsGolangType(pkg common.PackageKind, ref string) *LinkAsGolangType {
	return &LinkAsGolangType{
		Link: *NewRefLink[common.GolangType](pkg, ref),
	}
}

type LinkAsGolangType struct {
	Link[common.GolangType]
}

func (r *LinkAsGolangType) TypeName() string {
	return r.link.TypeName()
}

func (r *LinkAsGolangType) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *LinkAsGolangType) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return r.link.AssembleDefinition(ctx)
}

func (r *LinkAsGolangType) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	return r.link.AssembleUsage(ctx)
}
