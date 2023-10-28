package common

const RunPackagePath = "github.com/bdragon300/asyncapi-codegen-go/pkg/run"

type SchemaTag string

const (
	SchemaTagNoInline    SchemaTag = "noinline"
	SchemaTagPackageDown SchemaTag = "packageDown"
)

const TagName = "cgen"

type Linker interface {
	Add(query LinkQuerier)
	AddMany(query ListQuerier)
}

type LinkQuerier interface {
	Assign(obj any)
	Assigned() bool
	FindCallback() func(item Assembler, path []string) bool
	Ref() string
}

type ListQuerier interface {
	AssignList(objs []any)
	Assigned() bool
	FindCallback() func(item Assembler, path []string) bool
}
