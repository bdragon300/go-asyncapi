package linker

import (
	"fmt"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/specurl"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/compiler"

	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

type ObjectSource interface {
	AllObjects() []compiler.Object // TODO: make this as interface and move promise.go to linker
	Promises() []common.ObjectPromise
	ListPromises() []common.ObjectListPromise
}

func AssignRefs(sources map[string]ObjectSource) {
	logger := types.NewLogger("Linking 🔗")
	assigned := 1

	for assigned > 0 {
		assigned = 0
		for srcSpecID, source := range sources {
			for _, p := range source.Promises() {
				if p.Assigned() {
					continue // Assigned previously
				}

				if res, ok := resolvePromise(p, srcSpecID, sources); ok {
					switch p.Origin() {
					case common.PromiseOriginInternal:
						logger.Debug("Processing an internal ref", "$ref", p.Ref(), "target", res)
					case common.PromiseOriginUser:
						logger.Debug("Processing a ref", "$ref", p.Ref(), "target", res)
					default:
						panic(fmt.Sprintf("Unknown promise origin %v, this must not happen", p.Origin()))
					}

					p.Assign(res)
					assigned++
				}
			}
		}
	}
	logger.Trace("no more refs left to resolve on this iteration, leave it and go ahead")
}

func AssignListPromises(sources map[string]ObjectSource) {
	logger := types.NewLogger("Linking 🔗")
	totalAssigned := 0
	assigned := 1
	promisesCount := lo.SumBy(lo.Values(sources), func(item ObjectSource) int { return len(item.ListPromises()) })

	for assigned > 0 {
		assigned = 0
		for srcSpecID, source := range sources {
			for _, p := range source.ListPromises() {
				if p.Assigned() {
					continue // Assigned on previous iterations
				}
				if res, ok := resolveListPromise(p, srcSpecID, sources); ok {
					targets := strings.Join(
						lo.Map(lo.Slice(res, 0, 2), func(item common.Renderable, _ int) string { return item.String() }),
						", ",
					)
					if len(res) > 2 {
						targets += ", ..."
					}
					logger.Trace("Internal list promise resolved", "count", len(res), "targets", targets)

					p.AssignList(lo.ToAnySlice(res))
					assigned++
				}
			}
		}
		totalAssigned += assigned
	}
	if totalAssigned < promisesCount {
		logger.Debug("some internal list promises has not been resolved")
	}
}

func DanglingPromisesCount(sources map[string]ObjectSource) int {
	c := lo.SumBy(lo.Values(sources), func(item ObjectSource) int {
		return lo.CountBy(item.ListPromises(), func(p common.ObjectListPromise) bool { return !p.Assigned() })
	})
	return c + len(DanglingRefs(sources))
}

func DanglingRefs(sources map[string]ObjectSource) []string {
	return lo.FlatMap(lo.Values(sources), func(src ObjectSource, _ int) []string {
		return lo.FilterMap(src.Promises(), func(p common.ObjectPromise, _ int) (string, bool) {
			return p.Ref(), !p.Assigned()
		})
	})
}

func Stats(sources map[string]ObjectSource) string {
	promises := lo.FlatMap(lo.Values(sources), func(item ObjectSource, _ int) []common.ObjectPromise { return item.Promises() })
	listPromises := lo.FlatMap(lo.Values(sources), func(item ObjectSource, _ int) []common.ObjectListPromise { return item.ListPromises() })
	return fmt.Sprintf(
		"Linker: %d refs (%d user-defined (%d dangling), %d internal (%d dangling)), %d internal list promises (%d dangling)",
		len(promises),
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginUser }),
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginUser && !l.Assigned() }),
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginInternal }),
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginInternal && !l.Assigned() }),
		len(listPromises),
		lo.CountBy(listPromises, func(l common.ObjectListPromise) bool { return !l.Assigned() }),
	)
}

// TODO: detect ref loops to avoid infinite recursion
// TODO: external refs can not be resolved at first time -- leave them unresolved
func resolvePromise(p common.ObjectPromise, srcSpecID string, sources map[string]ObjectSource) (common.Renderable, bool) {
	tgtSpecID := srcSpecID

	ref := specurl.Parse(p.Ref())
	if ref.IsExternal() {
		tgtSpecID = ref.SpecID
	}
	if _, ok := sources[tgtSpecID]; !ok {
		return nil, false
	}

	srcObjects := sources[tgtSpecID].AllObjects()
	cb := func(_ common.Renderable, path []string) bool { return ref.MatchPointer(path) }
	if qcb := p.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(srcObjects, func(obj compiler.Object, _ int) bool { return cb(obj.Object, obj.ModuleURL.Pointer) })
	if len(found) != 1 {
		panic(fmt.Sprintf("Ref %q must point to one object, but %d objects found", p.Ref(), len(found)))
	}

	obj := found[0]
	switch v := obj.Object.(type) {
	case common.ObjectPromise:
		if !v.Assigned() {
			return nil, false
		}
		return resolvePromise(v, tgtSpecID, sources)
	case common.ObjectListPromise:
		panic(fmt.Sprintf("Ref %q must point to one object, but it points to a another list promise", p.Ref()))
	case common.Renderable:
		return v, true
	default:
		panic(fmt.Sprintf("Ref %q points to an object of unexpected type %T", p.Ref(), v))
	}
}

// TODO: detect ref loops to avoid infinite recursion
func resolveListPromise(p common.ObjectListPromise, srcSpecID string, sources map[string]ObjectSource) ([]common.Renderable, bool) {
	// Exclude links from selection in order to avoid duplicates in list
	cb := func(obj common.Renderable, _ []string) bool { return !isPromise(obj) }
	if qcb := p.FindCallback(); qcb != nil {
		cb = qcb
	}
	srcObjects := sources[srcSpecID].AllObjects()
	found := lo.Filter(srcObjects, func(obj compiler.Object, _ int) bool {
		return cb(obj.Object, obj.ModuleURL.Pointer)
	})

	var results []common.Renderable
	for _, obj := range found {
		switch v := obj.Object.(type) {
		case common.ObjectPromise:
			if !v.Assigned() {
				return results, false
			}
			resolved, ok := resolvePromise(v, srcSpecID, sources)
			if !ok {
				return results, false
			}
			results = append(results, resolved)
		case common.ObjectListPromise:
			if !v.Assigned() {
				return results, false
			}
			resolved, ok := resolveListPromise(v, srcSpecID, sources)
			if !ok {
				return results, false
			}
			results = append(results, resolved...)
		case common.Renderable:
			results = append(results, v)
		default:
			panic(fmt.Sprintf("Found an object of unexpected type %T", v))
		}
	}
	return results, true
}

func isPromise(obj any) bool {
	_, ok1 := obj.(common.ObjectPromise)
	_, ok2 := obj.(common.ObjectListPromise)
	return ok1 || ok2
}
