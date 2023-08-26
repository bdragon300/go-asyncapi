package scan

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
)

type LinkQuerier interface {
	Assign(obj any)
	Package() common.PackageKind
	Path() []string
	Ref() string
}

type ListQuerier interface {
	AssignList(obj []any)
	Package() common.PackageKind
	Path() []string
}

type Linker struct {
	queries     []LinkQuerier
	listQueries []ListQuerier
}

func (l *Linker) Add(query LinkQuerier) {
	l.queries = append(l.queries, query)
}

func (l *Linker) AddMany(query ListQuerier) {
	l.listQueries = append(l.listQueries, query)
}

func (l *Linker) Process(ctx *Context) {
	// TODO: resolve recursive refs
	for _, query := range l.queries {
		pkg := ctx.Packages[query.Package()]
		path := query.Path()
		if query.Ref() != "" {
			path = l.getPathByRef(query.Ref())
		}
		item, ok := pkg.Find(path)
		if !ok {
			panic(fmt.Sprintf("Cannot find %s path in the package %s", path, query.Package()))
		}
		query.Assign(item)
	}
	for _, query := range l.listQueries {
		pkg := ctx.Packages[query.Package()]
		items := utils.CastSliceItems[render.LangRenderer, any](pkg.List(query.Path()))
		query.AssignList(items)
	}
}

func (l *Linker) getPathByRef(ref string) []string {
	if !strings.HasPrefix(ref, "#/") {
		panic("We don't support external refs yet")
	}
	path, _ := strings.CutPrefix(ref, "#/")
	return strings.Split(path, "/")
}
