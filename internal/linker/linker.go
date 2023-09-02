package linker

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/utils"

	"github.com/samber/lo"

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

func (l *LocalLinker) Process(ctx *common.CompileContext) {
	// TODO: resolve recursive refs
	for _, query := range l.queries {
		pkg := ctx.Packages[query.Package()]
		refPath := l.getPathByRef(query.Ref())

		cb := func(_ any, path []string) bool { return utils.SlicesEqual(path, refPath) }
		if qcb := query.FindCallback(); qcb != nil {
			cb = qcb
		}

		item, ok := pkg.FindBy(cb)
		if !ok {
			panic(fmt.Sprintf("Cannot find %s path in the package %s", refPath, query.Package()))
		}
		query.Assign(item)
	}
	for _, query := range l.listQueries {
		pkg := ctx.Packages[query.Package()]

		cb := func(_ any, _ []string) bool { return true }
		if qcb := query.FindCallback(); qcb != nil {
			cb = qcb
		}

		items := lo.ToAnySlice(pkg.ListBy(cb))
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
