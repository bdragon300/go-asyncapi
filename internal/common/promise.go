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
	FindCallback() func(item CompileObject, path []string) bool
}

type ObjectListPromise interface {
	AssignList(objs []any)
	Assigned() bool
	FindCallback() func(item CompileObject, path []string) bool
	Ref() string
}
