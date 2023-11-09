package render

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func NewRefLink[T any](ref string, origin common.LinkOrigin) *Link[T] {
	return &Link[T]{ref: ref, origin: origin}
}

func NewCbLink[T any](findCb func(item common.Renderer, path []string) bool, origin common.LinkOrigin) *Link[T] {
	return &Link[T]{findCb: findCb, origin: origin}
}

type Link[T any] struct {
	ref    string
	origin common.LinkOrigin
	findCb func(item common.Renderer, path []string) bool

	target   T
	assigned bool
}

func (r *Link[T]) Assign(obj any) {
	r.target = obj.(T)
	r.assigned = true
}

func (r *Link[T]) Assigned() bool {
	return r.assigned
}

func (r *Link[T]) FindCallback() func(item common.Renderer, path []string) bool {
	return r.findCb
}

func (r *Link[T]) Target() T {
	return r.target
}

func (r *Link[T]) Ref() string {
	return r.ref
}

func (r *Link[T]) Origin() common.LinkOrigin {
	return r.origin
}

// List links can only be LinkOriginInternal, no way to set a callback in spec
func NewListCbLink[T any](findCb func(item common.Renderer, path []string) bool) *LinkList[T] {
	return &LinkList[T]{findCb: findCb}
}

type LinkList[T any] struct {
	findCb func(item common.Renderer, path []string) bool

	targets  []T
	assigned bool
}

func (r *LinkList[T]) AssignList(objs []any) {
	var ok bool
	r.targets, ok = lo.FromAnySlice[T](objs)
	if !ok {
		panic(fmt.Sprintf("Cannot assign slice of %+v to %T", objs, r.targets))
	}
	r.assigned = true
}

func (r *LinkList[T]) Assigned() bool {
	return r.assigned
}

func (r *LinkList[T]) FindCallback() func(item common.Renderer, path []string) bool {
	return r.findCb
}

func (r *LinkList[T]) Targets() []T {
	return r.targets
}

func NewRefLinkAsRenderer(ref string, origin common.LinkOrigin) *LinkAsRenderer {
	return &LinkAsRenderer{
		Link: *NewRefLink[common.Renderer](ref, origin),
	}
}

type LinkAsRenderer struct {
	Link[common.Renderer]
}

func (r *LinkAsRenderer) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderDefinition(ctx)
}

func (r *LinkAsRenderer) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderUsage(ctx)
}

func (r *LinkAsRenderer) DirectRendering() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *LinkAsRenderer) String() string {
	return "Ref to " + r.ref
}

func NewRefLinkAsGolangType(ref string, origin common.LinkOrigin) *LinkAsGolangType {
	return &LinkAsGolangType{
		Link: *NewRefLink[common.GolangType](ref, origin),
	}
}

type LinkAsGolangType struct {
	Link[common.GolangType]
}

func (r *LinkAsGolangType) TypeName() string {
	return r.target.TypeName()
}

func (r *LinkAsGolangType) DirectRendering() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *LinkAsGolangType) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderDefinition(ctx)
}

func (r *LinkAsGolangType) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderUsage(ctx)
}

func (r *LinkAsGolangType) String() string {
	return "Ref to " + r.ref
}
