package lang

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

func NewPromise[T any](ref string, origin common.PromiseOrigin) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin}
}

func NewCbPromise[T any](findCb func(item common.Renderable, path []string) bool, origin common.PromiseOrigin) *Promise[T] {
	return &Promise[T]{findCb: findCb, origin: origin}
}

type Promise[T any] struct {
	AssignErrorNote string // Optional error message additional note to be shown when assignment fails
	ref             string
	origin          common.PromiseOrigin
	findCb          func(item common.Renderable, path []string) bool

	target   T
	assigned bool
}

func (r *Promise[T]) Assign(obj any) {
	t, ok := obj.(T)
	if !ok {
		panic(fmt.Sprintf("Cannot assign an object %+v to a promise of type %T. %s", obj, r.target, r.AssignErrorNote))
	}
	r.target = t
	r.assigned = true
}

func (r *Promise[T]) Assigned() bool {
	return r.assigned
}

func (r *Promise[T]) FindCallback() func(item common.Renderable, path []string) bool {
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

func (r *Promise[T]) UnwrapGolangType() (common.GolangType, bool) {
	if !r.assigned {
		return nil, false
	}
	if v, ok := any(r.target).(GolangTypeWrapperType); ok {
		return v.UnwrapGolangType()
	}
	v, ok := any(r.target).(common.GolangType)
	return v, ok
}

func (r *Promise[T]) IsPointer() bool {
	if !r.assigned {
		return false
	}
	if v, ok := any(r.target).(GolangPointerType); ok {
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
func NewListCbPromise[T any](findCb func(item common.Renderable, path []string) bool) *ListPromise[T] {
	return &ListPromise[T]{findCb: findCb}
}

type ListPromise[T any] struct {
	AssignErrorNote string // Optional error message additional note to be shown when assignment fails
	findCb          func(item common.Renderable, path []string) bool

	targets  []T
	assigned bool
}

func (r *ListPromise[T]) AssignList(objs []any) {
	var ok bool
	r.targets, ok = lo.FromAnySlice[T](objs)
	if !ok {
		panic(fmt.Sprintf("Cannot assign slice of %+v to a promise of type %T. %s", objs, r.targets, r.AssignErrorNote))
	}
	r.assigned = true
}

func (r *ListPromise[T]) Assigned() bool {
	return r.assigned
}

func (r *ListPromise[T]) FindCallback() func(item common.Renderable, path []string) bool {
	return r.findCb
}

func (r *ListPromise[T]) Targets() []T {
	return r.targets
}

func NewRenderablePromise(ref string, origin common.PromiseOrigin) *RenderablePromise {
	return &RenderablePromise{
		Promise: *NewPromise[common.Renderable](ref, origin),
	}
}

type RenderablePromise struct {
	Promise[common.Renderable]
	// DirectRender marks the promise to be rendered directly, even if object it points to not marked to do so.
	// Be careful, in order to avoid duplicated object appearing in the output, this flag should be set only for
	// objects which are not marked to be rendered directly
	DirectRender bool
}

func (r *RenderablePromise) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *RenderablePromise) D() string {
	return r.target.D()
}

func (r *RenderablePromise) U() string {
	return r.target.U()
}

func (r *RenderablePromise) Selectable() bool {
	return r.DirectRender // Prevent rendering the object we're point to for several times
}

func (r *RenderablePromise) String() string {
	return "RenderablePromise -> " + r.ref
}

func NewGolangTypePromise(ref string, origin common.PromiseOrigin) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *NewPromise[common.GolangType](ref, origin),
	}
}

type GolangTypePromise struct {
	Promise[common.GolangType]
	// DirectRender marks the promise to be rendered directly, even if object it points to not marked to do so.
	// Be careful, in order to avoid duplicated object appearing in the output, this flag should be set only for
	// objects which are not marked to be rendered directly
	DirectRender bool  // TODO: rework or remove
}

func (r *GolangTypePromise) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *GolangTypePromise) TypeName() string {
	return r.target.TypeName()
}

func (r *GolangTypePromise) Selectable() bool {
	return r.DirectRender // Prevent rendering the object we're point to for several times
}

func (r *GolangTypePromise) IsPointer() bool {
	return r.target.IsPointer()
}

func (r *GolangTypePromise) D() string {
	return r.target.D()
}

func (r *GolangTypePromise) U() string {
	return r.target.U()
}

func (r *GolangTypePromise) DefinitionInfo() (*common.GolangTypeDefinitionInfo, error) {
	return r.target.DefinitionInfo()
}

func (r *GolangTypePromise) String() string {
	return "GolangTypePromise -> " + r.ref
}
