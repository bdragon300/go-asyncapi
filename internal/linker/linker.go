package linker

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/compiler"

	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
)

func NewSpecLinker() *SpecLinker {
	return &SpecLinker{logger: common.NewLogger("Linking 🔗")}
}

type SpecLinker struct {
	externalQueriesQueue []common.LinkQuerier
	queries              []common.LinkQuerier
	listQueries          []common.ListQuerier
	logger               *common.Logger
}

func (l *SpecLinker) Add(query common.LinkQuerier) {
	l.queries = append(l.queries, query)
}

func (l *SpecLinker) AddMany(query common.ListQuerier) {
	l.listQueries = append(l.listQueries, query)
}

type linkerSource interface {
	AllItems() []compiler.PackageItem
}

func (l *SpecLinker) Process(source linkerSource) error {
	objects := source.AllItems()

	assigned := 0
	prevAssigned := 0
	for assigned < len(l.queries) {
		for _, query := range l.queries {
			if res, ok := resolveLink(query, objects); ok {
				switch query.Origin() {
				case common.LinkOriginInternal:
					l.logger.Trace("Internal ref resolved", "$ref", query.Ref(), "target", res)
				case common.LinkOriginUser:
					l.logger.Debug("Ref resolved", "$ref", query.Ref(), "target", res)
				default:
					panic(fmt.Sprintf("Unknown link origin, this must not happen: %v", query.Origin()))
				}

				query.Assign(res)
				assigned++
			}
		}
		if assigned == prevAssigned {
			l.logger.Trace("no more ref links to assign on this iteration, leave it and go ahead")
			break
		}
		prevAssigned = assigned
	}

	assigned = 0
	prevAssigned = 0
	for assigned < len(l.listQueries) {
		for _, query := range l.listQueries {
			if res, ok := resolveListLink(query, objects); ok {
				targets := strings.Join(
					lo.Map(lo.Slice(res, 0, 2), func(item common.Renderer, _ int) string { return item.String() }),
					", ",
				)
				if len(res) > 2 {
					targets += ", ..."
				}
				l.logger.Trace("Internal list link resolved", "count", len(res), "targets", targets)

				query.AssignList(lo.ToAnySlice(res))
				assigned++
			}
		}
		if assigned == prevAssigned {
			l.logger.Trace("no more list links to assign on this iteration, leave it and go ahead")
			break
		}
		prevAssigned = assigned
	}

	l.logger.Trace("iteration completed")

	return nil
}

func (l *SpecLinker) UserQueriesCount() int {
	return lo.CountBy(l.queries, func(item common.LinkQuerier) bool {
		return item.Origin() == common.LinkOriginUser
	})
}

func (l *SpecLinker) PopExternalQuery() (common.LinkQuerier, bool) {
	if len(l.externalQueriesQueue) == 0 {
		return nil, false
	}

	res := l.externalQueriesQueue[len(l.externalQueriesQueue)-1]
	l.externalQueriesQueue = l.externalQueriesQueue[:len(l.externalQueriesQueue)-1]
	return res, true
}

func (l *SpecLinker) DanglingQueries() []string {
	return lo.Flatten([][]string{
		lo.FilterMap(l.queries, func(item common.LinkQuerier, index int) (string, bool) {
			return item.Ref(), !item.Assigned()
		}),
		lo.FilterMap(l.listQueries, func(item common.ListQuerier, index int) (string, bool) {
			return "<internal list query>", !item.Assigned()
		}),
	})
}

func (l *SpecLinker) Stats() string {
	return fmt.Sprintf(
		"Linker: %d refs (%d user (%d dangling), %d internal (%d dangling)), %d internal list queries (%d dangling)",
		len(l.queries),
		lo.CountBy(l.queries, func(l common.LinkQuerier) bool { return l.Origin() == common.LinkOriginUser }),
		lo.CountBy(l.queries, func(l common.LinkQuerier) bool { return l.Origin() == common.LinkOriginUser && !l.Assigned() }),
		lo.CountBy(l.queries, func(l common.LinkQuerier) bool { return l.Origin() == common.LinkOriginInternal }),
		lo.CountBy(l.queries, func(l common.LinkQuerier) bool { return l.Origin() == common.LinkOriginInternal && !l.Assigned() }),
		len(l.listQueries),
		lo.CountBy(l.listQueries, func(l common.ListQuerier) bool { return !l.Assigned() }),
	)
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
// TODO: external refs can not be resolved at first time -- leave them unresolved
func resolveLink(q common.LinkQuerier, objects []compiler.PackageItem) (common.Renderer, bool) {
	refPath := getPathByRef(q.Ref())
	cb := func(_ common.Renderer, path []string) bool { return utils.SlicesEqual(path, refPath) }
	if qcb := q.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(objects, func(obj compiler.PackageItem, _ int) bool {
		return cb(obj.Typ, obj.Path)
	})
	if len(found) != 1 {
		panic(fmt.Sprintf("Ref %q must point to one object, got %d objects", q.Ref(), len(found)))
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
	case common.Renderer:
		return v, true
	default:
		panic(fmt.Sprintf("Ref %q points to an object of unexpected type %T", q.Ref(), v))
	}
}

// TODO: detect ref loops to avoid infinite recursion
func resolveListLink(q common.ListQuerier, objects []compiler.PackageItem) ([]common.Renderer, bool) {
	// Exclude links from selection in order to avoid duplicates in list
	cb := func(obj common.Renderer, _ []string) bool { return !isLinker(obj) }
	if qcb := q.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(objects, func(obj compiler.PackageItem, _ int) bool {
		return cb(obj.Typ, obj.Path)
	})

	var results []common.Renderer
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
		case common.Renderer:
			results = append(results, v)
		default:
			panic(fmt.Sprintf("Found an object of unexpected type %T", v))
		}
	}
	return results, true
}
