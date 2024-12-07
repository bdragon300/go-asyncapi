package lang

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

func defaultAssignCb[T any](obj any) T {
	t, ok := obj.(T)
	if !ok {
		panic(fmt.Sprintf("Object %+v is not a type %T", obj, new(T)))
	}
	return t
}

func NewPromise[T any](ref string, origin common.PromiseOrigin) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin, assignCb: defaultAssignCb[T]}
}

func NewAssignCbPromise[T any](ref string, origin common.PromiseOrigin, assignCb func(obj any) T) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin, assignCb: assignCb}
}

type Promise[T any] struct {
	AssignErrorNote string // Optional error message additional note to be shown when assignment fails
	ref             string
	origin          common.PromiseOrigin
	assignCb        func(obj any) T

	target   T
	assigned bool
}

func (r *Promise[T]) Assign(obj any) {
	if r.assignCb != nil {
		r.target = r.assignCb(obj)
		r.assigned = true
		return
	}
	t, ok := obj.(T)
	if !ok {
		panic(fmt.Sprintf("Object %+v is not a type %T in promise %q. %s", obj, new(T), r.ref, r.AssignErrorNote))
	}
	r.target = t
	r.assigned = true
}

func (r *Promise[T]) Assigned() bool {
	return r.assigned
}

func (r *Promise[T]) T() T {
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

func NewListCbPromise[T any](findCb func(item common.Renderable, path []string) bool) *ListPromise[T] {
	return &ListPromise[T]{findCb: findCb}
}

type ListPromise[T any] struct {
	AssignErrorNote string // Optional error message additional note to be shown when assignment fails
	ref             string
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

func (r *ListPromise[T]) T() []T {
	return r.targets
}

func (r *ListPromise[T]) Ref() string {
	return r.ref
}

func NewRenderablePromise(ref string, origin common.PromiseOrigin) *RenderablePromise {
	return &RenderablePromise{
		Promise: *NewPromise[common.Renderable](ref, origin),
	}
}

type RenderablePromise struct {
	Promise[common.Renderable]
}

func (r *RenderablePromise) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *RenderablePromise) Selectable() bool {
	return r.origin == common.PromiseOriginUser && r.target.Selectable()
}

func (r *RenderablePromise) String() string {
	return "RenderablePromise -> " + r.ref
}

func NewGolangTypePromise(ref string, origin common.PromiseOrigin) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *NewPromise[common.GolangType](ref, origin),
	}
}

func NewGolangTypeAssignCbPromise(ref string, origin common.PromiseOrigin, assignCb func(obj any) common.GolangType) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *NewAssignCbPromise[common.GolangType](ref, origin, assignCb),
	}
}

type GolangTypePromise struct {
	Promise[common.GolangType]
}

func (r *GolangTypePromise) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *GolangTypePromise) TypeName() string {
	return r.target.TypeName()
}

func (r *GolangTypePromise) Selectable() bool {
	return r.origin == common.PromiseOriginUser && r.target.Selectable()
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
