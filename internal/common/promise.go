package common

type PromiseOrigin int

const (
	PromiseOriginUser PromiseOrigin = iota
	PromiseOriginInternal
)

type ObjectPromise interface {
	Assign(obj any)
	Assigned() bool
	Ref() string
	Origin() PromiseOrigin
	FindCallback() PromiseFindCbFunc
}

type ObjectListPromise interface {
	AssignList(objs []any)
	Assigned() bool
	FindCallback() PromiseFindCbFunc
}

type PromiseFindCbFunc func(item CompileObject, path []string) bool
