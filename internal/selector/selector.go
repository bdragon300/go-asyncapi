package selector

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/samber/lo"
	"regexp"
)

func SelectObjects(objects []compiler.Object, filters common.RenderSelectionFilterConfig) []compiler.Object {
	filterChain := getFiltersChain(filters)

	return lo.Filter(objects, func(object compiler.Object, _ int) bool {
		for _, filter := range filterChain {
			if !filter(object) {
				return false
			}
		}
		return true
	})
}

//func FindSelectionByObject(object compiler.Object, selections []common.RenderSelectionConfig) *common.RenderSelectionConfig {
//	// TODO: nested structures defined in Channel or smth like this will not work (ObjectKind==lang), they have no explicit selections
//	for _, selection := range selections {
//		filtersChain := getFiltersChain(selection.RenderSelectionFilterConfig)
//		match := lo.ContainsBy(filtersChain, func(f filterFunc) bool {
//			return f(object)
//		})
//		if match {
//			return &selection
//		}
//	}
//	return nil
//}

type filterFunc func(compiler.Object) bool

func getFiltersChain(filters common.RenderSelectionFilterConfig) []filterFunc {
	var filterChain []filterFunc
	filterChain = append(filterChain, func(object compiler.Object) bool {
		return object.Object.Selectable()
	})
	if filters.ObjectKindRe != "" {
		re := regexp.MustCompile(filters.ObjectKindRe) // TODO: compile 1 time (and below)
		filterChain = append(filterChain, func(object compiler.Object) bool {
			return re.MatchString(string(object.Object.Kind()))
		})
	}
	if filters.ModuleURLRe != "" {
		re := regexp.MustCompile(filters.ModuleURLRe)
		filterChain = append(filterChain, func(object compiler.Object) bool {
			return re.MatchString(object.ModuleURL.SpecID)
		})
	}
	if filters.PathRe != "" {
		re := regexp.MustCompile(filters.PathRe)
		filterChain = append(filterChain, func(object compiler.Object) bool {
			return re.MatchString(object.ModuleURL.PointerRef())
		})
	}
	return filterChain
}
