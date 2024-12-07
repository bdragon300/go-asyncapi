package selector

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"regexp"
)

type protoSelectable interface {
	ProtoObjects() []common.Renderable
}

func SelectObjects(objects []common.CompileObject, selection common.RenderSelectionConfig) []common.CompileObject {
	var allObjects []common.CompileObject
	filterChain := getFiltersChain(selection)

	// Enumerate all objects and replace (unwind) them with their proto objects if they have any.
	// This is needed because a channel, message, server are presented in spec as one object, therefore any $ref points
	// to the whole object. But in templates we pass every proto part of the object separately.
	for _, object := range objects {
		if selectable, ok := object.Renderable.(protoSelectable); ok {
			allObjects = append(allObjects, lo.Map(selectable.ProtoObjects(), func(obj common.Renderable, _ int) common.CompileObject {
				return common.CompileObject{
					Renderable: obj,
					ObjectURL:  object.ObjectURL,
				}
			})...)
		} else {
			allObjects = append(allObjects, object)
		}
	}

	return lo.Filter(objects, func(object common.CompileObject, _ int) bool {
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

type filterFunc func(common.CompileObject) bool

func getFiltersChain(selection common.RenderSelectionConfig) []filterFunc {
	var filterChain []filterFunc
	filterChain = append(filterChain, func(object common.CompileObject) bool {
		return object.Selectable()
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
