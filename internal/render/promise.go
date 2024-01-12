package render

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func NewPromise[T any](ref string, origin common.PromiseOrigin) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin}
}

func NewCbPromise[T any](findCb func(item common.Renderer, path []string) bool, origin common.PromiseOrigin) *Promise[T] {
	return &Promise[T]{findCb: findCb, origin: origin}
}

type Promise[T any] struct {
	ref    string
	origin common.PromiseOrigin
	findCb func(item common.Renderer, path []string) bool

	target   T
	assigned bool
}

func (r *Promise[T]) Assign(obj any) {
	r.target = obj.(T)
	r.assigned = true
}

func (r *Promise[T]) Assigned() bool {
	return r.assigned
}

func (r *Promise[T]) FindCallback() func(item common.Renderer, path []string) bool {
	return r.findCb
}

func (r *Promise[T]) Target() T {
	return r.target
}

func (r *Promise[T]) Ref() string {
	return r.ref
}

func (r *Promise[T]) Origin() common.PromiseOrigin {
	return r.origin
}

func (r *Promise[T]) WrappedGolangType() (common.GolangType, bool) {
	if !r.assigned {
		return nil, false
	}
	v, ok := any(r.target).(common.GolangType)
	return v, ok
}

func (r *Promise[T]) IsPointer() bool {
	if !r.assigned {
		return false
	}
	if v, ok := any(r.target).(golangPointerType); ok {
		return v.IsPointer()
	}
	return false
}

func (r *Promise[T]) IsStruct() bool {
	if !r.assigned {
		return false
	}
	if v, ok := any(r.target).(golangStructType); ok {
		return v.IsStruct()
	}
	return false
}

// List links can only be PromiseOriginInternal, no way to set a callback in spec
func NewListCbPromise[T any](findCb func(item common.Renderer, path []string) bool) *ListPromise[T] {
	return &ListPromise[T]{findCb: findCb}
}

type ListPromise[T any] struct {
	findCb func(item common.Renderer, path []string) bool

	targets  []T
	assigned bool
}

func (r *ListPromise[T]) AssignList(objs []any) {
	var ok bool
	r.targets, ok = lo.FromAnySlice[T](objs)
	if !ok {
		panic(fmt.Sprintf("Cannot assign slice of %+v to type %T", objs, r.targets))
	}
	r.assigned = true
}

func (r *ListPromise[T]) Assigned() bool {
	return r.assigned
}

func (r *ListPromise[T]) FindCallback() func(item common.Renderer, path []string) bool {
	return r.findCb
}

func (r *ListPromise[T]) Targets() []T {
	return r.targets
}

func NewRendererPromise(ref string, origin common.PromiseOrigin) *RendererPromise {
	return &RendererPromise{
		Promise: *NewPromise[common.Renderer](ref, origin),
	}
}

type RendererPromise struct {
	Promise[common.Renderer]
}

func (r *RendererPromise) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderDefinition(ctx)
}

func (r *RendererPromise) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderUsage(ctx)
}

func (r *RendererPromise) DirectRendering() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *RendererPromise) ID() string {
	if r.Assigned() {
		return r.target.ID()
	}
	return ""
}

func (r *RendererPromise) String() string {
	return "RendererPromise for " + r.ref
}

func NewGolangTypePromise(ref string, origin common.PromiseOrigin) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *NewPromise[common.GolangType](ref, origin),
	}
}

type GolangTypePromise struct {
	Promise[common.GolangType]
}

func (r *GolangTypePromise) TypeName() string {
	return r.target.TypeName()
}

func (r *GolangTypePromise) DirectRendering() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *GolangTypePromise) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderDefinition(ctx)
}

func (r *GolangTypePromise) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	return r.target.RenderUsage(ctx)
}

func (r *GolangTypePromise) ID() string {
	return "GolangTypePromise"
}

func (r *GolangTypePromise) String() string {
	return "GolangTypePromise for " + r.ref
}
