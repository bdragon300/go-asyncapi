package common

type PromiseOrigin int

const (
	PromiseOriginUser PromiseOrigin = iota
	PromiseOriginInternal
)

type ObjectPromise interface {
	Assign(obj Renderable)
	Assigned() bool
	Ref() string
	Origin() PromiseOrigin
	FindCallback() PromiseFindCbFunc
}

type ObjectListPromise interface {
	AssignList(objs []Renderable)
	Assigned() bool
	FindCallback() PromiseFindCbFunc
}

type PromiseFindCbFunc func(item CompileObject, path []string) bool
