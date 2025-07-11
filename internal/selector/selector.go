package selector

import (
	"regexp"

	"github.com/bdragon300/go-asyncapi/internal/log"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

// Select selects the artifacts from the given list based on the code layout rule.
func Select(artifacts []common.Artifact, layoutItem common.ConfigLayoutItem) []common.Artifact {
	logger := log.GetLogger("")

	filtersChain := buildFiltersChain(layoutItem)
	res := lo.Filter(artifacts, func(object common.Artifact, _ int) bool {
		logger.Trace("--> Process filters", "object", object.String())
		for ind, filter := range filtersChain {
			if !filter(object) && !layoutItem.Not {
				logger.Trace("---> Discarded by filter", "index", ind)
				return false
			}
		}
		if layoutItem.Not {
			logger.Trace("---> Discarded by Not filter")
			return false
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

func buildFiltersChain(layoutItem common.ConfigLayoutItem) []filterFunc {
	var filterChain []filterFunc
	logger := log.GetLogger("")

	logger.Trace("-> Use Selectable filter", "index", len(filterChain))
	// Consider only the selectable artifacts
	filterChain = append(filterChain, func(object common.Artifact) bool {
		return object.Selectable()
	})

	if len(layoutItem.Protocols) > 0 {
		logger.Trace("-> Use Protocol filter", "index", len(filterChain))
		filterChain = append(filterChain, func(object common.Artifact) bool {
			// Check if object has at least one of the proto objects of the given protocols
			if o, ok := object.(protoObjectSelector); ok {
				return lo.SomeBy(layoutItem.Protocols, func(protocol string) bool {
					return o.SelectProtoObject(protocol) != nil
				})
			}
			return false
		})
	}
	if len(layoutItem.ArtifactKinds) > 0 {
		logger.Trace("-> Use ArtifactKinds filter", "index", len(filterChain))
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return lo.Contains(layoutItem.ArtifactKinds, string(object.Kind()))
		})
	}
	if layoutItem.ModuleURLRe != "" {
		logger.Trace("-> Use ModuleURLRe filter", "index", len(filterChain))
		re := regexp.MustCompile(layoutItem.ModuleURLRe)
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return re.MatchString(object.Pointer().Location())
		})
	}
	if layoutItem.PathRe != "" {
		logger.Trace("-> Use PathRe filter", "index", len(filterChain))
		re := regexp.MustCompile(layoutItem.PathRe)
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return re.MatchString(object.Pointer().PointerString())
		})
	}
	if layoutItem.NameRe != "" {
		logger.Trace("-> Use NameRe filter", "index", len(filterChain))
		re := regexp.MustCompile(layoutItem.NameRe)
		filterChain = append(filterChain, func(object common.Artifact) bool {
			return re.MatchString(object.Name())
		})
	}
	return filterChain
}
