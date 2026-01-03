// Package linker contains the linking stage logic.
//
// The linking stage goes after the compilation stage. It gets a bunch of promise objects from all documents
// on the one hand and the bunch of compilation artifacts from all documents on the other, and then tries to bind them together.
// That is, the linker performs the late binding for promises.
//
// Basically, promises act as placeholders for the related objects that may not be yet compiled.
// Internally, the promise is just a pointer (nil initially) with metadata on how to find an object(s) it should point to.
// Linker searches for the object and assigns it to that pointer.
// See also [asyncapi] package description.
//
// There are several types of promises:
//
//   - a regular promise, points to a single object. May address the object by ref or by a callback. Internal purposes only.
//   - [lang.Ref] object. A promise that is used especially for $ref urls in documents.
//     Addresses a single object only by ref.
//   - a list promise. For internal purposes only, points to a list of objects. Addresses the objects by a callback only.
//
// Due to different implementations the list promises and the other types are resolved separately.
//
// Linking stage is considered successful if all promises have been resolved. If not, it returns an error.
package linker

import (
	"fmt"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"

	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

type ObjectSource interface {
	Artifacts() []common.Artifact // TODO: make this as interface and move promise.go to linker
	Promises() []common.ObjectPromise
	ListPromises() []common.ObjectListPromise
}

func ResolvePromises(sources map[string]ObjectSource) {
	logger := log.GetLogger(log.LoggerPrefixLinking)
	assigned := 1

	for assigned > 0 {
		assigned = 0
		for docPath, source := range sources {
			for _, p := range source.Promises() {
				if p.Assigned() {
					continue // Assigned previously
				}

				if res, ok := resolvePromise(p, docPath, sources); ok {
					switch p.Origin() {
					case common.PromiseOriginInternal:
						logger.Debug("Processing an internal promise", "$ref", p.Ref(), "target", res, "addr", fmt.Sprintf("%p", res))
					case common.PromiseOriginRef:
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

func ResolveListPromises(sources map[string]ObjectSource) {
	logger := log.GetLogger(log.LoggerPrefixLinking)
	totalAssigned := 0
	assigned := 1
	promisesCount := lo.SumBy(lo.Values(sources), func(item ObjectSource) int { return len(item.ListPromises()) })

	for assigned > 0 {
		assigned = 0
		for docURL, source := range sources {
			for _, p := range source.ListPromises() {
				if p.Assigned() {
					continue // Assigned on previous iterations
				}
				if res, ok := resolveListPromise(p, docURL, sources); ok {
					targets := strings.Join(
						lo.Map(lo.Slice(res, 0, 2), func(item common.Artifact, _ int) string { return item.String() }),
						", ",
					)
					if len(res) > 2 {
						targets += ", ..."
					}
					logger.Trace("Internal list promise resolved", "count", len(res), "targets", targets)

					p.AssignList(res)
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

func UnresolvedPromisesCount(sources map[string]ObjectSource) int {
	c := lo.SumBy(lo.Values(sources), func(item ObjectSource) int {
		return lo.CountBy(item.ListPromises(), func(p common.ObjectListPromise) bool { return !p.Assigned() })
	})
	return c + len(UnresolvedPromises(sources))
}

func UnresolvedPromises(sources map[string]ObjectSource) []string {
	return lo.FlatMap(lo.Values(sources), func(src ObjectSource, _ int) []string {
		return lo.FilterMap(src.Promises(), func(p common.ObjectPromise, _ int) (string, bool) {
			return p.Ref(), !p.Assigned()
		})
	})
}

// Stats returns a string with the statistics of the linking stage.
func Stats(sources map[string]ObjectSource) string {
	promises := lo.FlatMap(lo.Values(sources), func(item ObjectSource, _ int) []common.ObjectPromise { return item.Promises() })
	listPromises := lo.FlatMap(lo.Values(sources), func(item ObjectSource, _ int) []common.ObjectListPromise { return item.ListPromises() })
	return fmt.Sprintf(
		"Linker (total/unresolved): refs (%d/%d), promises (%d/%d), list promises (%d/%d); total %d",
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginRef }),
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginRef && !l.Assigned() }),
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginInternal }),
		lo.CountBy(promises, func(l common.ObjectPromise) bool { return l.Origin() == common.PromiseOriginInternal && !l.Assigned() }),
		len(listPromises),
		lo.CountBy(listPromises, func(l common.ObjectListPromise) bool { return !l.Assigned() }),
		len(promises)+len(listPromises),
	)
}

// TODO: detect ref loops to avoid infinite recursion
// TODO: external refs can not be resolved at first time -- leave them unresolved
func resolvePromise(p common.ObjectPromise, docLocation string, sources map[string]ObjectSource) (common.Artifact, bool) {
	var ref *jsonpointer.JSONPointer
	var cb, userCb common.PromiseFindCbFunc

	target := docLocation
	ref = lo.Must(jsonpointer.Parse(p.Ref()))
	if ref.Location() != "" {
		target = ref.Location()
	}

	if _, ok := sources[target]; !ok {
		// Promise points to a document that possibly may not been compiled yet, postpone it
		return nil, false
	}

	srcArtifacts := sources[target].Artifacts()
	if cb = p.FindCallback(); cb == nil {
		cb = func(item common.Artifact) bool { return ref.MatchPointer(item.Pointer().Pointer) }
	}
	found := lo.Filter(srcArtifacts, func(obj common.Artifact, _ int) bool { return cb(obj) })
	if len(found) != 1 {
		panic(fmt.Sprintf("Ref %q must point to one object, but %d objects found", p.Ref(), len(found)))
	}

	obj := found[0]

	// If we set a callback, let it decide which objects should get to promise, don't do recursive resolving
	if userCb != nil {
		return obj, true
	}

	switch v := obj.(type) {
	case common.ObjectPromise:
		if !v.Assigned() {
			return nil, false
		}
		return resolvePromise(v, target, sources)
	case common.ObjectListPromise:
		panic(fmt.Sprintf("Ref %q must point to one object, but it points to another list promise", p.Ref()))
	case common.Artifact:
		return v, true
	default:
		panic(fmt.Sprintf("Ref %q points to an object of unexpected type %T", p.Ref(), v))
	}
}

func resolveListPromise(p common.ObjectListPromise, docURL string, sources map[string]ObjectSource) ([]common.Artifact, bool) {
	// Exclude links from selection in order to avoid duplicates in list
	cb := p.FindCallback()
	if cb == nil {
		panic("List promise must have a callback, this is a bug")
	}
	srcArtifacts := sources[docURL].Artifacts() // FIXME: here we only assign the artifacts from the same document, should be from all documents
	found := lo.Filter(srcArtifacts, func(obj common.Artifact, _ int) bool {
		return cb(obj)
	})

	// Let the callback decide which objects should be the promise targets, don't do recursive resolving
	return found, true
}
