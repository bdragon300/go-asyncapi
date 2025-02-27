package selector

import (
	"regexp"

	"github.com/bdragon300/go-asyncapi/internal/log"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

// Select filters the given artifacts from list based on the given selection.
func Select(artifacts []common.Artifact, selection common.ConfigSelectionItem) []common.Artifact {
	logger := log.GetLogger("")

	filtersChain := buildFiltersChain(selection)
	res := lo.Filter(artifacts, func(object common.Artifact, _ int) bool {
		logger.Trace("--> Process filters", "object", object.String())
		for ind, filter := range filtersChain {
			if !filter(object) {
				logger.Trace("---> Discarded by filter", "index", ind)
				return false
			}
		}
		logger.Trace("---> Passed")
		return true
	})
	return res
}

type filterFunc func(common.Artifact) bool

type protoObjectSelector interface {
	SelectProtoObject(protocol string) common.Artifact
}

func buildFiltersChain(selection common.ConfigSelectionItem) []filterFunc {
	var filterChain []filterFunc
	logger := log.GetLogger("")

	logger.Trace("-> Use Selectable filter", "index", len(filterChain))
	filterChain = append(filterChain, func(object common.Artifact) bool {
		return object.Selectable()
	})
	if len(selection.Protocols) > 0 {
		logger.Trace("-> Use Protocol filter", "index", len(filterChain))
		filterChain = append(filterChain, func(object common.Artifact) bool {
			// Check if object has at least one of the proto objects of the given protocols
			if o, ok := object.(protoObjectSelector); ok {
				return lo.SomeBy(selection.Protocols, func(protocol string) bool {
					return o.SelectProtoObject(protocol) != nil
				})
			}
			return false
		})
	}
	if len(selection.ArtifactKinds) > 0 {
		logger.Trace("-> Use ArtifactKinds filter", "index", len(filterChain))
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return lo.Contains(selection.ArtifactKinds, string(object.Kind()))
		})
	}
	if selection.ModuleURLRe != "" {
		logger.Trace("-> Use ModuleURLRe filter", "index", len(filterChain))
		re := regexp.MustCompile(selection.ModuleURLRe)
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return re.MatchString(object.Pointer().Location())
		})
	}
	if selection.PathRe != "" {
		logger.Trace("-> Use PathRe filter", "index", len(filterChain))
		re := regexp.MustCompile(selection.PathRe)
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return re.MatchString(object.Pointer().PointerString())
		})
	}
	if selection.NameRe != "" {
		logger.Trace("-> Use NameRe filter", "index", len(filterChain))
		re := regexp.MustCompile(selection.NameRe)
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return re.MatchString(object.Name())
		})
	}
	return filterChain
}
