package selector

import (
	"regexp"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

// Select filters the given artifacts from list based on the given selection.
func Select(artifacts []common.CompileArtifact, selection common.ConfigSelectionItem) []common.CompileArtifact {
	filtersChain := buildFiltersChain(selection)
	res := lo.Filter(artifacts, func(object common.CompileArtifact, _ int) bool {
		for _, filter := range filtersChain {
			if !filter(object) {
				return false
			}
		}
		return true
	})
	return res
}

type filterFunc func(common.CompileArtifact) bool

type protoObjectSelector interface {
	SelectProtoObject(protocol string) common.Renderable
}

func buildFiltersChain(selection common.ConfigSelectionItem) []filterFunc {
	// TODO: trace logging in every filter
	var filterChain []filterFunc
	filterChain = append(filterChain, func(object common.CompileArtifact) bool {
		return object.Selectable()
	})
	if len(selection.Protocols) > 0 {
		filterChain = append(filterChain, func(object common.CompileArtifact) bool {
			// Check if object has at least one of the proto objects of the given protocols
			if o, ok := object.Renderable.(protoObjectSelector); ok {
				return lo.SomeBy(selection.Protocols, func(protocol string) bool {
					return o.SelectProtoObject(protocol) != nil
				})
			}
			return false
		})
	}
	if len(selection.ObjectKinds) > 0 {
		filterChain = append(filterChain, func(object common.CompileArtifact) bool {
			return lo.Contains(selection.ObjectKinds, string(object.Kind()))
		})
	}
	if selection.ModuleURLRe != "" {
		re := regexp.MustCompile(selection.ModuleURLRe)
		filterChain = append(filterChain, func(object common.CompileArtifact) bool {
			return re.MatchString(object.ObjectURL.Location())
		})
	}
	if selection.PathRe != "" {
		re := regexp.MustCompile(selection.PathRe)
		filterChain = append(filterChain, func(object common.CompileArtifact) bool {
			return re.MatchString(object.ObjectURL.PointerString())
		})
	}
	if selection.NameRe != "" {
		re := regexp.MustCompile(selection.NameRe)
		filterChain = append(filterChain, func(object common.CompileArtifact) bool {
			return re.MatchString(object.Renderable.Name())
		})
	}
	return filterChain
}
