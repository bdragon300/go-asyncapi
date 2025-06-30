package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

type promiseAssignCbFunc[T common.Artifact] func(obj common.Artifact) T

// NewPromise returns a new promise, that addresses the object by ref URL.
//
// If assignCb is not nil, it will be called when linker assigns the object to the promise and its return value will be
// assigned to the target. This callback could be used to convert the object or do some additional checks.
func NewPromise[T common.Artifact](ref string, assignCb promiseAssignCbFunc[T]) *Promise[T] {
	return newPromise[T](ref, common.PromiseOriginInternal, nil, assignCb)
}

// NewCbPromise returns a new promise, that uses find callback to find the object. Linker calls this callback for every
// artifact in all storages to find the one that matches this promise. Should return true only once.
//
// *NOTE*: in the find callback don't rely on other promises were resolved, because resolving order is not guaranteed.
//
// If assignCb is not nil, it will be called to get the value to assign to target.
// This callback could be used for type converting or do some additional checks.
func NewCbPromise[T common.Artifact](findCb common.PromiseFindCbFunc, assignCb promiseAssignCbFunc[T]) *Promise[T] {
	return newPromise("", common.PromiseOriginInternal, findCb, assignCb)
}

func newPromise[T common.Artifact](
	ref string,
	origin common.PromiseOrigin,
	findCb common.PromiseFindCbFunc,
	assignCb promiseAssignCbFunc[T],
) *Promise[T] {
	return &Promise[T]{ref: ref, origin: origin, findCb: findCb, assignCb: assignCb}
}

// Promise is the object-placeholder used for late-binding between objects, when one of them (target) is unavailable on
// the compilation stage. The linker finds a target object on the linker stage using metadata contained in a promise
// and assigns it to this promise. So on the rendering stage we could access the object by this promise, that was
// not available on the compilation stage.
//
// Basically, the promise is just a pointer (nil initially) with metadata how to find an object(s) it should point to.
// The target object can be addresses by ref URL or by a callback function.
type Promise[T common.Artifact] struct {
	// AssignErrorNote is the additional note to be shown in error message when assignment fails
	AssignErrorNote string

	ref      string
	origin   common.PromiseOrigin
	findCb   common.PromiseFindCbFunc
	assignCb promiseAssignCbFunc[T]

	target   T
	assigned bool
}

// Assign binds the object to the promise. Called by the linker.
func (r *Promise[T]) Assign(obj common.Artifact) {
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

// Assigned returns true if the object is already bound to the promise.
func (r *Promise[T]) Assigned() bool {
	return r.assigned
}

// T returns the target object.
func (r *Promise[T]) T() T {
	return r.target
}

// Ref returns the reference to the object if any.
func (r *Promise[T]) Ref() string {
	return r.ref
}

// Origin returns the origin of the promise. See [common.PromiseOrigin] for details.
func (r *Promise[T]) Origin() common.PromiseOrigin {
	return r.origin
}

// FindCallback returns the callback function to find the object in the storage. If not nil, Ref is ignored.
// Called by the linker.
func (r *Promise[T]) FindCallback() common.PromiseFindCbFunc {
	return r.findCb
}

func (r *Promise[T]) DerefGolangType() common.GolangType {
	switch v := any(r.target).(type) {
	case GolangReferenceType:
		return v.DerefGolangType()
	case common.GolangType:
		return v
	}
	return nil
}

func (r *Promise[T]) CanBeAddressed() bool {
	if !r.assigned {
		return false
	}
	if v, ok := any(r.target).(common.GolangType); ok {
		return v.CanBeAddressed()
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

// NewGolangTypePromise returns a new promise, that addresses the GolangType object by ref URL. See [NewPromise] for details.
func NewGolangTypePromise(ref string, assignCb promiseAssignCbFunc[common.GolangType]) *GolangTypePromise {
	return &GolangTypePromise{
		Promise: *newPromise[common.GolangType](ref, common.PromiseOriginInternal, nil, assignCb),
	}
}

// GolangTypePromise is the promise object that can be substituted where the GolangType is expected.
type GolangTypePromise struct {
	BaseJSONPointed
	Promise[common.GolangType]
}

func (r *GolangTypePromise) Name() string {
	return r.target.Name()
}

func (r *GolangTypePromise) Kind() common.ArtifactKind {
	return r.target.Kind()
}

func (r *GolangTypePromise) Selectable() bool {
	return r.origin == common.PromiseOriginRef && r.target.Selectable()
}

func (r *GolangTypePromise) Visible() bool {
	return r.origin == common.PromiseOriginRef && r.target.Visible()
}

func (r *GolangTypePromise) String() string {
	return "GolangTypePromise -> " + r.ref
}

func (r *GolangTypePromise) CanBeAddressed() bool {
	return r.target.CanBeAddressed()
}

func (r *GolangTypePromise) CanBeDereferenced() bool {
	return r.target.CanBeDereferenced()
}

func (r *GolangTypePromise) GoTemplate() string {
	return r.target.GoTemplate()
}

func (r *GolangTypePromise) Unwrap() common.Artifact {
	return common.DerefArtifact(r.target)
}

// NewListCbPromise returns a new promise, that uses find callback to find the list of objects. Linker calls this
// callback for every artifact in all storages. All objects for which the findCb returns true, will be added to the
// target list and assigned to the promise.
//
// *NOTE*: in the find callback don't rely on other promises were resolved, because resolving order is not guaranteed.
//
// If assignItemCb is not nil, it will be called for every target object to get the values list to assign to target.
// This callback could be used for type converting or do some additional checks.
func NewListCbPromise[T common.Artifact](findItemCb common.PromiseFindCbFunc, assignItemCb promiseAssignCbFunc[T]) *ListPromise[T] {
	return &ListPromise[T]{findCb: findItemCb, assignItemCb: assignItemCb}
}

// ListPromise is the promise object like ObjectPromise but can target to list of objects. It can't be referenced by
// ref and intended only for internal use.
type ListPromise[T common.Artifact] struct {
	// AssignErrorNote is the optional error message additional note to be shown to user when assignment fails
	AssignErrorNote string

	findCb       common.PromiseFindCbFunc
	assignItemCb promiseAssignCbFunc[T]

	targets  []T
	assigned bool
}

// AssignList binds the list of objects to the promise. Called by the linker.
func (r *ListPromise[T]) AssignList(objs []common.Artifact) {
	if r.assignItemCb != nil {
		r.targets = lo.Map(objs, func(item common.Artifact, _ int) T {
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

// Assigned returns true if the list of objects is already bound to the promise.
func (r *ListPromise[T]) Assigned() bool {
	return r.assigned
}

// FindCallback returns the callback function to find the object in the storage. Called by the linker.
func (r *ListPromise[T]) FindCallback() common.PromiseFindCbFunc {
	return r.findCb
}

// T returns the target list.
func (r *ListPromise[T]) T() []T {
	return r.targets
}
