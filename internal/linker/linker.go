package linker

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
)

func NewLocalLinker() *LocalLinker {
	return &LocalLinker{logger: common.NewLogger("Linking 🔗")}
}

type LocalLinker struct {
	queries     []common.LinkQuerier
	listQueries []common.ListQuerier
	logger      *common.Logger
}

func (l *LocalLinker) Add(query common.LinkQuerier) {
	l.queries = append(l.queries, query)
}

func (l *LocalLinker) AddMany(query common.ListQuerier) {
	l.listQueries = append(l.listQueries, query)
}

func (l *LocalLinker) Process(ctx *common.CompileContext) error {
	objects := lo.Flatten(lo.MapToSlice(ctx.Packages, func(k string, v *common.Package) []common.PackageItem {
		return v.Items()
	}))

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
					panic(fmt.Sprintf("Unknown link origin: %v", query.Origin()))
				}

				query.Assign(res)
				assigned++
			}
		}
		if assigned == prevAssigned {
			notAssigned := lo.FilterMap(l.queries, func(item common.LinkQuerier, index int) (string, bool) {
				return item.Ref(), !item.Assigned()
			})
			// FIXME: here can be also internal refs, not only those from spec. It's better not to show them to user
			return fmt.Errorf(
				"orphan $refs (target not found, reference recursion, etc.): %s",
				strings.Join(notAssigned, "; "),
			)
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
			notAssigned := lo.FilterMap(l.queries, func(item common.LinkQuerier, index int) (string, bool) {
				return item.Ref(), !item.Assigned()
			})
			// FIXME: here can be also internal refs, not only those from spec. It's better not to show them to user
			return fmt.Errorf(
				"orphan $refs (target not found, reference recursion, etc.): %s",
				strings.Join(notAssigned, "; "),
			)
		}
		prevAssigned = assigned
	}

	l.logger.Info("Finished", "refs", l.UserQueriesCount())

	return nil
}

func (l *LocalLinker) UserQueriesCount() int {
	return lo.CountBy(l.queries, func(item common.LinkQuerier) bool {
		return item.Origin() == common.LinkOriginUser
	})
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
func resolveLink(q common.LinkQuerier, objects []common.PackageItem) (common.Renderer, bool) {
	refPath := getPathByRef(q.Ref())
	cb := func(_ common.Renderer, path []string) bool { return utils.SlicesEqual(path, refPath) }
	if qcb := q.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(objects, func(obj common.PackageItem, _ int) bool {
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
	case common.Renderer:
		return v, true
	default:
		panic(fmt.Sprintf("Ref %q points to an object of unexpected type %T", q.Ref(), v))
	}
}

// TODO: detect ref loops to avoid infinite recursion
func resolveListLink(q common.ListQuerier, objects []common.PackageItem) ([]common.Renderer, bool) {
	// Exclude links from selection in order to avoid duplicates in list
	cb := func(obj common.Renderer, _ []string) bool { return !isLinker(obj) }
	if qcb := q.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(objects, func(obj common.PackageItem, _ int) bool {
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
