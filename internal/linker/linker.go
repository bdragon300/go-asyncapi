package linker

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/specurl"

	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

type ObjectSource interface {
	AllObjects() []common.CompileObject // TODO: make this as interface and move promise.go to linker
	Promises() []common.ObjectPromise
	ListPromises() []common.ObjectListPromise
}

func AssignRefs(sources map[string]ObjectSource) {
	logger := log.GetLogger(log.LoggerPrefixLinking)
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
						logger.Debug("Processing an internal ref", "$ref", p.Ref(), "target", res, "addr", fmt.Sprintf("%p", res))
					case common.PromiseOriginUser:
						logger.Debug("Processing a ref", "$ref", p.Ref(), "target", res, "addr", fmt.Sprintf("%p", res))
					default:
						panic(fmt.Sprintf("Unknown promise origin %v, this is a bug", p.Origin()))
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
	logger := log.GetLogger(log.LoggerPrefixLinking)
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
	cb := func(_ common.CompileObject, path []string) bool { return ref.MatchPointer(path) }
	userCallback := p.FindCallback()
	if userCallback != nil {
		cb = userCallback
	}
	found := lo.Filter(srcObjects, func(obj common.CompileObject, _ int) bool { return cb(obj, obj.ObjectURL.Pointer) })
	if len(found) != 1 {
		panic(fmt.Sprintf("Ref %q must point to one object, but %d objects found", p.Ref(), len(found)))
	}

	obj := found[0]

	// If we set a callback, let it decide which objects should get to promise, don't do recursive resolving
	if userCallback != nil {
		return obj.Renderable, true
	}

	switch v := obj.Renderable.(type) {
	case common.ObjectPromise:
		if !v.Assigned() {
			return nil, false
		}
		return resolvePromise(v, tgtSpecID, sources)
	case common.ObjectListPromise:
		panic(fmt.Sprintf("Ref %q must point to one object, but it points to another list promise", p.Ref()))
	case common.Renderable:
		return v, true
	default:
		panic(fmt.Sprintf("Ref %q points to an object of unexpected type %T", p.Ref(), v))
	}
}

func resolveListPromise(p common.ObjectListPromise, srcSpecID string, sources map[string]ObjectSource) ([]common.Renderable, bool) {
	// Exclude links from selection in order to avoid duplicates in list
	cb := p.FindCallback()
	if cb == nil {
		panic("List promise must have a callback, this is a bug")
	}
	srcObjects := sources[srcSpecID].AllObjects()
	found := lo.Filter(srcObjects, func(obj common.CompileObject, _ int) bool {
		return cb(obj, obj.ObjectURL.Pointer)
	})

	// Let the callaback decide which objects should be promise targets, don't do recursive resolving
	return lo.Map(found, func(item common.CompileObject, _ int) common.Renderable { return item.Renderable }), true
}
