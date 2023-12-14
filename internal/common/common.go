package common

const RunPackagePath = "github.com/bdragon300/asyncapi-codegen-go/pkg/run"

type SchemaTag string

const (
	SchemaTagNoInline    SchemaTag = "noinline"
	SchemaTagPackageDown SchemaTag = "packageDown"
)

type LinkOrigin int

const (
	LinkOriginUser LinkOrigin = iota
	LinkOriginInternal
)

const TagName = "cgen"

type LinkQuerier interface {
	Assign(obj any)
	Assigned() bool
	FindCallback() func(item Renderer, path []string) bool
	Ref() string
	Origin() LinkOrigin
}

type ListQuerier interface {
	AssignList(objs []any)
	Assigned() bool
	FindCallback() func(item Renderer, path []string) bool
}
