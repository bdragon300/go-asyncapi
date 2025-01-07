package lang

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

type promiseAssignCbFunc[T common.Renderable] func(obj common.Renderable) T

func NewPromise[T common.Renderable](ref string, assignCb promiseAssignCbFunc[T]) *Promise[T] {
	return newPromise[T](ref, common.PromiseOriginInternal, nil, assignCb)
}

func NewCbPromise[T common.Renderable](findCb common.PromiseFindCbFunc, assignCb promiseAssignCbFunc[T]) *Promise[T] {
	return newPromise("", common.PromiseOriginInternal, findCb, assignCb)
}

func newPromise[T common.Renderable](
	ref string,
	origin common.PromiseOrigin,
	findCb common.PromiseFindCbFunc,
	assignCb promiseAssignCbFunc[T],
) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin, findCb: findCb, assignCb: assignCb}
}

type Promise[T common.Renderable] struct{
	// AssignErrorNote is the optional error message additional note to be shown to user when assignment fails
	AssignErrorNote string

	ref             string
	origin   common.PromiseOrigin
	findCb   common.PromiseFindCbFunc
	assignCb promiseAssignCbFunc[T]

	target   T
	assigned bool
}

func (r *Promise[T]) Assign(obj common.Renderable) {
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

func (r *Promise[T]) FindCallback() common.PromiseFindCbFunc {
	return r.findCb
}

func (r *Promise[T]) UnwrapGolangType() common.GolangType {
	if v, ok := any(r.target).(common.GolangType); ok {
		return unwrapGolangPromise(v)
	}
	return nil
}

func (r *Promise[T]) Addressable() bool {
	if !r.assigned {
		return false
	}
	if v, ok := any(r.target).(common.GolangType); ok {
		return v.Addressable()
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

func NewGolangTypePromise(ref string, assignCb promiseAssignCbFunc[common.GolangType]) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *newPromise[common.GolangType](ref, common.PromiseOriginInternal, nil, assignCb),
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

func (r *GolangTypePromise) Name() string {
	return r.target.Name()
}

func (r *GolangTypePromise) UnwrapRenderable() common.Renderable {
	return common.DerefRenderable(r.target)
}

func (r *GolangTypePromise) Addressable() bool {
	return r.target.Addressable()
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

func NewListCbPromise[T common.Renderable](findCb common.PromiseFindCbFunc, assignItemCb promiseAssignCbFunc[T]) *ListPromise[T] {
	return &ListPromise[T]{findCb: findCb, assignItemCb: assignItemCb}
}

type ListPromise[T common.Renderable] struct {
	// AssignErrorNote is the optional error message additional note to be shown to user when assignment fails
	AssignErrorNote string

	findCb common.PromiseFindCbFunc
	assignItemCb promiseAssignCbFunc[T]

	targets  []T
	assigned bool
}

func (r *ListPromise[T]) AssignList(objs []common.Renderable) {
	if r.assignItemCb != nil {
		r.targets = lo.Map(objs, func(item common.Renderable, _ int) T {
			return r.assignItemCb(item)
		})
		r.assigned = true
		return
	}

	for _, obj := range objs {
		v, ok := obj.(T)
		if !ok {
			panic(fmt.Sprintf("Object %+v is not a type %T in list promise. %s", obj, new(T), r.AssignErrorNote))
		}
		r.targets = append(r.targets, v)
	}
	r.assigned = true
}

func (r *ListPromise[T]) Assigned() bool {
	return r.assigned
}

func (r *ListPromise[T]) FindCallback() common.PromiseFindCbFunc {
	return r.findCb
}

func (r *ListPromise[T]) T() []T {
	return r.targets
}

func unwrapGolangPromise(val common.GolangType) common.GolangType {
	if o, ok := val.(GolangTypeWrapper); ok {
		val = o.UnwrapGolangType()
	}
	return val
}
