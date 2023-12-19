package linker

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/bdragon300/asyncapi-codegen-go/internal/compiler"

	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
)

func NewSpecLinker() *SpecLinker {
	return &SpecLinker{
		objPromises:     make(map[string][]common.ObjectPromise),
		objListPromises: make(map[string][]common.ObjectListPromise),
		logger:          types.NewLogger("Linking 🔗"),
	}
}

type SpecLinker struct {
	objPromises     map[string][]common.ObjectPromise     // Promises by source spec ID
	objListPromises map[string][]common.ObjectListPromise // Promises by source spec ID
	logger          *types.Logger
}

func (l *SpecLinker) AddPromise(p common.ObjectPromise, sourceSpecID string) {
	l.objPromises[sourceSpecID] = append(l.objPromises[sourceSpecID], p)
}

func (l *SpecLinker) AddListPromise(p common.ObjectListPromise, sourceSpecID string) {
	l.objListPromises[sourceSpecID] = append(l.objListPromises[sourceSpecID], p)
}

type ObjectSource interface {
	AllObjects() []compiler.Object // TODO: make this as interface and move link.go to linker
}

func (l *SpecLinker) ProcessRefs(sources map[string]ObjectSource) {
	assigned := 0
	prevAssigned := 0
	promisesCount := len(lo.Flatten(lo.Values(l.objPromises)))
	for assigned < promisesCount {
		for srcSpecID, promises := range l.objPromises {
			for _, p := range promises {
				if p.Assigned() {
					continue // Assigned on previous iterations
				}

				if res, ok := resolvePromise(p, srcSpecID, sources); ok {
					switch p.Origin() {
					case common.PromiseOriginInternal:
						l.logger.Trace("Internal ref resolved", "$ref", p.Ref(), "target", res)
					case common.PromiseOriginUser:
						l.logger.Debug("Ref resolved", "$ref", p.Ref(), "target", res)
					default:
						panic(fmt.Sprintf("Unknown link origin %v, this must not happen", p.Origin()))
					}

					p.Assign(res)
					assigned++
				}
			}
		}
		if assigned == prevAssigned {
			l.logger.Trace("no more refs to resolve on this iteration, leave it and go ahead")
			break
		}
		prevAssigned = assigned
	}
}

func (l *SpecLinker) ProcessListPromises(sources map[string]ObjectSource) {
	assigned := 0
	prevAssigned := 0
	promisesCount := len(lo.Flatten(lo.Values(l.objListPromises)))
	for assigned < promisesCount {
		for srcSpecID, promises := range l.objListPromises {
			for _, p := range promises {
				if p.Assigned() {
					continue // Assigned on previous iterations
				}
				if res, ok := resolveListPromise(p, srcSpecID, sources); ok {
					targets := strings.Join(
						lo.Map(lo.Slice(res, 0, 2), func(item common.Renderer, _ int) string { return item.String() }),
						", ",
					)
					if len(res) > 2 {
						targets += ", ..."
					}
					l.logger.Trace("Internal list link resolved", "count", len(res), "targets", targets)

					p.AssignList(lo.ToAnySlice(res))
					assigned++
				}
			}
		}
		if assigned == prevAssigned {
			l.logger.Debug("some internal list promises has not been resolved")
			break
		}
		prevAssigned = assigned
	}
}

func (l *SpecLinker) DanglingPromisesCount() int {
	return lo.CountBy(lo.Flatten(lo.Values(l.objPromises)), func(item common.ObjectPromise) bool {
		return !item.Assigned()
	}) + lo.CountBy(lo.Flatten(lo.Values(l.objListPromises)), func(item common.ObjectListPromise) bool {
		return !item.Assigned()
	})
}

func (l *SpecLinker) DanglingRefs() []string {
	return lo.FilterMap(lo.Flatten(lo.Values(l.objPromises)), func(item common.ObjectPromise, index int) (string, bool) {
		return item.Ref(), !item.Assigned()
	})
}

func (l *SpecLinker) Stats() string {
	promises := lo.Flatten(lo.Values(l.objPromises))
	listPromises := lo.Flatten(lo.Values(l.objListPromises))
	return fmt.Sprintf(
		"Linker: %d refs (%d user (%d dangling), %d internal (%d dangling)), %d internal list promises (%d dangling)",
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
func resolvePromise(p common.ObjectPromise, srcSpecID string, sources map[string]ObjectSource) (common.Renderer, bool) {
	tgtSpecID, refPointer := utils.SplitSpecPath(p.Ref())
	if tgtSpecID == "" {
		tgtSpecID = srcSpecID // `#/ref` references
	}
	if _, ok := sources[tgtSpecID]; !ok {
		return nil, false
	}

	srcObjects := sources[tgtSpecID].AllObjects()
	refPath := splitRefPointer(refPointer)
	cb := func(_ common.Renderer, path []string) bool { return utils.SlicesEqual(path, refPath) }
	if qcb := p.FindCallback(); qcb != nil {
		cb = qcb
	}
	found := lo.Filter(srcObjects, func(obj compiler.Object, _ int) bool { return cb(obj.Object, obj.Path) })
	if len(found) != 1 {
		panic(fmt.Sprintf("Ref %q must point to one object, got %d objects", p.Ref(), len(found)))
	}

	obj := found[0]
	switch v := obj.Object.(type) {
	case common.ObjectPromise:
		if !v.Assigned() {
			return nil, false
		}
		return resolvePromise(v, tgtSpecID, sources)
	case common.ObjectListPromise:
		panic(fmt.Sprintf("Ref %q must point to one object, but points to a nested list link", p.Ref()))
	case common.Renderer:
		return v, true
	default:
		panic(fmt.Sprintf("Ref %q points to an object of unexpected type %T", p.Ref(), v))
	}
}

// TODO: detect ref loops to avoid infinite recursion
func resolveListPromise(p common.ObjectListPromise, srcSpecID string, sources map[string]ObjectSource) ([]common.Renderer, bool) {
	// Exclude links from selection in order to avoid duplicates in list
	cb := func(obj common.Renderer, _ []string) bool { return !isPromise(obj) }
	if qcb := p.FindCallback(); qcb != nil {
		cb = qcb
	}
	srcObjects := sources[srcSpecID].AllObjects()
	found := lo.Filter(srcObjects, func(obj compiler.Object, _ int) bool {
		return cb(obj.Object, obj.Path)
	})

	var results []common.Renderer
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
		case common.Renderer:
			results = append(results, v)
		default:
			panic(fmt.Sprintf("Found an object of unexpected type %T", v))
		}
	}
	return results, true
}

func splitRefPointer(refPointer string) []string {
	parts := strings.Split(refPointer, "/")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	return parts
}

func isPromise(obj any) bool {
	_, ok1 := obj.(common.ObjectPromise)
	_, ok2 := obj.(common.ObjectListPromise)
	return ok1 || ok2
}
