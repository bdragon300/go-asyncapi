package render

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func NewPromise[T any](ref string, origin common.PromiseOrigin) *Link[T] {
	return &Link[T]{ref: ref, origin: origin}
}

func NewCbPromise[T any](findCb func(item common.Renderer, path []string) bool, origin common.PromiseOrigin) *Link[T] {
	return &Link[T]{findCb: findCb, origin: origin}
}

type Link[T any] struct {
	ref    string
	origin common.PromiseOrigin
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

func (r *Link[T]) Origin() common.PromiseOrigin {
	return r.origin
}

func (r *Link[T]) WrappedGolangType() (common.GolangType, bool) {
	if !r.assigned {
		return nil, false
	}
	v, ok := any(r.target).(common.GolangType)
	return v, ok
}

func (r *Link[T]) IsPointer() bool {
	if !r.assigned {
		return false
	}
	if v, ok := any(r.target).(pointerGolangType); ok {
		return v.IsPointer()
	}
	return false
}

func (r *Link[T]) IsStruct() bool {
	if !r.assigned {
		return false
	}
	if v, ok := any(r.target).(structGolangType); ok {
		return v.IsStruct()
	}
	return false
}

// List links can only be PromiseOriginInternal, no way to set a callback in spec
func NewListCbPromise[T any](findCb func(item common.Renderer, path []string) bool) *LinkList[T] {
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
		panic(fmt.Sprintf("Cannot assign slice of %+v to type %T", objs, r.targets))
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

func NewRendererPromise(ref string, origin common.PromiseOrigin) *LinkAsRenderer {
	return &LinkAsRenderer{
		Link: *NewPromise[common.Renderer](ref, origin),
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

func NewGolangTypePromise(ref string, origin common.PromiseOrigin) *LinkAsGolangType {
	return &LinkAsGolangType{
		Link: *NewPromise[common.GolangType](ref, origin),
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
	return r.ref
}
