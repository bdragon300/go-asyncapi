package selector

import (
	"cmp"
	"regexp"
	"slices"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/samber/lo"
)

type artifactStorage interface {
	Artifacts() []common.Artifact
	DocumentURL() jsonpointer.JSONPointer
}

// GatherArtifacts gathers artifacts from storages and joins them into a single list of dereferenced artifacts
// (i.e. without refs).
func GatherArtifacts[T artifactStorage](docs ...T) []common.Artifact {
	// Sort documents by their URL to have a deterministic order
	docs = slices.SortedFunc(slices.Values(docs), func(a, b T) int {
		return cmp.Compare(a.DocumentURL().String(), b.DocumentURL().String())
	})
	r := lo.FlatMap(docs, func(doc T, _ int) []common.Artifact {
		return lo.FilterMap(doc.Artifacts(), func(artifact common.Artifact, _ int) (common.Artifact, bool) {
			if !artifact.Selectable() {
				return nil, false
			}
			return common.DerefArtifact(artifact), true
		})
	})
	// Remove duplicates if any artifact was referenced multiple times
	r = lo.Uniq(r)
	return r
}

// ApplyFilters selects the artifacts from the given list based on the code layout rule.
func ApplyFilters(artifacts []common.Artifact, layoutItem common.LayoutItemOpts) []common.Artifact {
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

type protocolArtifact interface {
	ActiveProtocols() []string
}

func buildFiltersChain(layoutItem common.LayoutItemOpts) []filterFunc {
	var filterChain []filterFunc
	logger := log.GetLogger("")

	// Filter out the artifacts that don't match to protocols in selections config (e.g. Channel, Operation, Message)
	if len(layoutItem.Protocols) > 0 {
		logger.Trace("-> Use Protocol filter", "index", len(filterChain))
		filterChain = append(filterChain, func(object common.Artifact) bool {
			if o, ok := object.(protocolArtifact); ok {
				return lo.Some(layoutItem.Protocols, o.ActiveProtocols())
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
