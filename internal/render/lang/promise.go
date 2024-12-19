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

func newPromise[T any](ref string, origin common.PromiseOrigin) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin}
}

func newAssignCbPromise[T any](ref string, origin common.PromiseOrigin, assignCb func(obj any) T) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin, assignCb: assignCb}
}

func NewInternalPromise[T any](ref string) *Promise[T] {
	return &Promise[T]{ref: ref, origin: common.PromiseOriginInternal, assignCb: defaultAssignCb[T]}
}

func NewInternalCbPromise[T any](findCb func(item common.CompileObject, path []string) bool) *Promise[T] {
	return &Promise[T]{origin: common.PromiseOriginInternal, findCb: findCb}
}

type Promise[T any] struct {
	AssignErrorNote string // Optional error message additional note to be shown when assignment fails
	ref             string
	origin          common.PromiseOrigin
	findCb          func(item common.CompileObject, path []string) bool
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

func (r *Promise[T]) FindCallback() func(item common.CompileObject, path []string) bool {
	return r.findCb
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
	if v, ok := any(r.target).(common.GolangType); ok {
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

func NewListCbPromise[T any](findCb func(item common.CompileObject, path []string) bool) *ListPromise[T] {
	return &ListPromise[T]{findCb: findCb}
}

type ListPromise[T any] struct {
	AssignErrorNote string // Optional error message additional note to be shown when assignment fails
	ref             string
	findCb          func(item common.CompileObject, path []string) bool

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

func (r *ListPromise[T]) FindCallback() func(item common.CompileObject, path []string) bool {
	return r.findCb
}

func (r *ListPromise[T]) T() []T {
	return r.targets
}

func (r *ListPromise[T]) Ref() string {
	return r.ref
}

func NewUserPromise(ref string, name string, selectable *bool) *RenderablePromise {
	return &RenderablePromise{
		Promise: *newPromise[common.Renderable](ref, common.PromiseOriginUser),
		selectable: selectable,
		name: name,
	}
}

type RenderablePromise struct {
	Promise[common.Renderable]
	selectable *bool
	name string
}

func (r *RenderablePromise) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *RenderablePromise) Selectable() bool {
	if r.selectable == nil {
		return r.origin == common.PromiseOriginUser && r.target.Selectable()
	}
	return r.origin == common.PromiseOriginUser && *r.selectable
}

func (r *RenderablePromise) Visible() bool {
	return r.origin == common.PromiseOriginUser && r.target.Visible()
}

func (r *RenderablePromise) String() string {
	return "RenderablePromise -> " + r.ref
}

func (r *RenderablePromise) GetOriginalName() string {
	n, _ := lo.Coalesce(r.name, r.target.GetOriginalName())
	return n
}

func (r *RenderablePromise) UnwrapRenderable() common.Renderable {
	return unwrapRenderablePromise(r.target)
}

func NewInternalGolangTypePromise(ref string) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *newPromise[common.GolangType](ref, common.PromiseOriginInternal),
	}
}

func NewInternalGolangTypeAssignCbPromise(ref string, assignCb func(obj any) common.GolangType) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *newAssignCbPromise[common.GolangType](ref, common.PromiseOriginInternal, assignCb),
	}
}

type GolangTypePromise struct {
	Promise[common.GolangType]
}

func (r *GolangTypePromise) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *GolangTypePromise) Selectable() bool {
	return r.origin == common.PromiseOriginUser && r.target.Selectable()
}

func (r *GolangTypePromise) Visible() bool {
	return r.origin == common.PromiseOriginUser && r.target.Visible()
}

func (r *GolangTypePromise) GetOriginalName() string {
	return r.target.GetOriginalName()
}

func (r *GolangTypePromise) UnwrapGolangType() common.GolangType {
	return unwrapGolangPromise(r.target)
}

func (r *GolangTypePromise) UnwrapRenderable() common.Renderable {
	return unwrapRenderablePromise(r.target)
}

func (r *GolangTypePromise) IsPointer() bool {
	return r.target.IsPointer()
}

func (r *GolangTypePromise) GoTemplate() string {
	return r.target.GoTemplate()
}

func (r *GolangTypePromise) String() string {
	return "GolangTypePromise -> " + r.ref
}

func unwrapRenderablePromise(val common.Renderable) common.Renderable {
	type renderableUnwrapper interface {
		UnwrapRenderable() common.Renderable
	}

	if o, ok := val.(renderableUnwrapper); ok {
		return o.UnwrapRenderable()
	}
	return val
}


func unwrapGolangPromise(val common.GolangType) common.GolangType {
	type golangTypeUnwrapper interface {
		UnwrapGolangType() common.GolangType
	}

	if o, ok := val.(golangTypeUnwrapper); ok {
		val = o.UnwrapGolangType()
	}
	return val
}