package common

type PromiseOrigin int

const (
	PromiseOriginUser PromiseOrigin = iota
	PromiseOriginInternal
)

type ObjectPromise interface {
	Assign(obj any)
	Assigned() bool
	FindCallback() func(item Renderable, path []string) bool
	Ref() string
	Origin() PromiseOrigin
}

type ObjectListPromise interface {
	AssignList(objs []any)
	Assigned() bool
	FindCallback() func(item Renderable, path []string) bool
}
