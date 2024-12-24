package selector

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"regexp"
)

func SelectObjects(objects []common.CompileObject, selection common.RenderSelectionConfig) []common.CompileObject {
	filterChain := getFiltersChain(selection)
	// TODO: logging

	allObjects := lo.Map(objects, func(object common.CompileObject, _ int) common.CompileObject {
		return common.CompileObject{Renderable: object.Renderable, ObjectURL: object.ObjectURL}
	})

	res := lo.Filter(allObjects, func(object common.CompileObject, _ int) bool {
		for _, filter := range filterChain {
			if !filter(object) {
				return false
			}
		}
		return true
	})
	return res
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

type filterFunc func(common.CompileObject) bool

func getFiltersChain(selection common.RenderSelectionConfig) []filterFunc {
	var filterChain []filterFunc
	filterChain = append(filterChain, func(object common.CompileObject) bool {
		return object.Selectable() && object.Visible()
	})
	if selection.ObjectKindRe != "" {
		re := regexp.MustCompile(selection.ObjectKindRe) // TODO: compile 1 time (and below)
		filterChain = append(filterChain, func(object common.CompileObject) bool {
			return re.MatchString(string(object.Kind()))
		})
	}
	if selection.ModuleURLRe != "" {
		re := regexp.MustCompile(selection.ModuleURLRe)
		filterChain = append(filterChain, func(object common.CompileObject) bool {
			return re.MatchString(object.ObjectURL.SpecID)
		})
	}
	if selection.PathRe != "" {
		re := regexp.MustCompile(selection.PathRe)
		filterChain = append(filterChain, func(object common.CompileObject) bool {
			return re.MatchString(object.ObjectURL.PointerRef())
		})
	}
	return filterChain
}
