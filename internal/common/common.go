package common

const RunPackagePath = "github.com/bdragon300/asyncapi-codegen-go/run"

type PromiseOrigin int

const (
	PromiseOriginUser PromiseOrigin = iota
	PromiseOriginInternal
)

type ObjectPromise interface {
	Assign(obj any)
	Assigned() bool
	FindCallback() func(item Renderer, path []string) bool
	Ref() string
	Origin() PromiseOrigin
}

type ObjectListPromise interface {
	AssignList(objs []any)
	Assigned() bool
	FindCallback() func(item Renderer, path []string) bool
}
