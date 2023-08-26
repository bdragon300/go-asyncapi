package linker

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/utils"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
)

type LocalLinker struct {
	queries     []common.LinkQuerier
	listQueries []common.ListQuerier
}

func (l *LocalLinker) Add(query common.LinkQuerier) {
	l.queries = append(l.queries, query)
}

func (l *LocalLinker) AddMany(query common.ListQuerier) {
	l.listQueries = append(l.listQueries, query)
}

func (l *LocalLinker) Process(ctx *common.Context) {
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
		items := utils.CastSliceItems[common.Assembled, any](pkg.List(query.Path()))
		query.AssignList(items)
	}
}

func (l *LocalLinker) getPathByRef(ref string) []string {
	if !strings.HasPrefix(ref, "#/") {
		panic("We don't support external refs yet")
	}
	path, _ := strings.CutPrefix(ref, "#/")
	return strings.Split(path, "/")
}
