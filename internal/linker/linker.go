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
	objects := lo.Flatten(lo.MapToSlice(ctx.Packages, func(k common.PackageKind, v common.Package) []common.PackageItem[common.Assembler] {
		return v.Items()
	}))

	assigned := 0
	prevAssigned := 0
	for assigned < len(l.queries) {
		for _, query := range l.queries {
			if res, ok := resolveLink(query, objects); ok {
				query.Assign(res)
				assigned++
			}
		}
		if assigned == prevAssigned {
			panic(fmt.Sprintf("%d refs in schema are not resolvable", len(l.queries)-assigned))
		}
		prevAssigned = assigned
	}

	assigned = 0
	prevAssigned = 0
	for assigned < len(l.listQueries) {
		for _, query := range l.listQueries {
			if res, ok := resolveListLink(query, objects); ok {
				query.AssignList(lo.ToAnySlice(res))
				assigned++
			}
		}
		if assigned == prevAssigned {
			panic(fmt.Sprintf("%d refs in schema are not resolvable", len(l.queries)-assigned))
		}
		prevAssigned = assigned
	}
}

func getPathByRef(ref string) []string {
	if !strings.HasPrefix(ref, "#/") {
		panic("We don't support external refs yet")
	}
	path, _ := strings.CutPrefix(ref, "#/")
	return strings.Split(path, "/")
}

func isLinker(obj any) bool {
	_, ok1 := obj.(common.LinkQuerier)
	_, ok2 := obj.(common.ListQuerier)
	return ok1 || ok2
}

// TODO: detect ref loops to avoid infinite recursion
func resolveLink(q common.LinkQuerier, objects []common.PackageItem[common.Assembler]) (common.Assembler, bool) {
	refPath := getPathByRef(q.Ref())
	cb := func(_ common.Assembler, path []string) bool { return utils.SlicesEqual(path, refPath) }
	if qcb := q.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(objects, func(obj common.PackageItem[common.Assembler], _ int) bool {
		return cb(obj.Typ, obj.Path)
	})
	if len(found) != 1 {
		panic(fmt.Sprintf("Ref %q points to %d objects", q.Ref(), len(found)))
	}

	obj := found[0]
	switch v := obj.Typ.(type) {
	case common.LinkQuerier:
		if !v.Assigned() {
			return nil, false
		}
		return resolveLink(v, objects)
	case common.ListQuerier:
		panic(fmt.Sprintf("Ref %q must point to one object, but points to a nested list link", q.Ref()))
	case common.Assembler:
		return v, true
	default:
		panic(fmt.Sprintf("Ref %q points to an object of unexpected type %T", q.Ref(), v))
	}
}

// TODO: detect ref loops to avoid infinite recursion
func resolveListLink(q common.ListQuerier, objects []common.PackageItem[common.Assembler]) ([]common.Assembler, bool) {
	// Exclude links from selection in order to avoid duplicates in list
	cb := func(obj common.Assembler, _ []string) bool { return !isLinker(obj) }
	if qcb := q.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(objects, func(obj common.PackageItem[common.Assembler], _ int) bool {
		return cb(obj.Typ, obj.Path)
	})

	var results []common.Assembler
	for _, obj := range found {
		switch v := obj.Typ.(type) {
		case common.LinkQuerier:
			if !v.Assigned() {
				return results, false
			}
			resolved, ok := resolveLink(v, objects)
			if !ok {
				return results, false
			}
			results = append(results, resolved)
		case common.ListQuerier:
			if !v.Assigned() {
				return results, false
			}
			resolved, ok := resolveListLink(v, objects)
			if !ok {
				return results, false
			}
			results = append(results, resolved...)
		case common.Assembler:
			results = append(results, v)
		default:
			panic(fmt.Sprintf("Found an object of unexpected type %T", v))
		}
	}
	return results, true
}
