package selector

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/samber/lo"
	"regexp"
)

func SelectObjects(objects []compiler.Object, filters common.RenderSelectionFilterConfig) []compiler.Object {
	var filterChain []func(compiler.Object) bool
	filterChain = append(filterChain, func(object compiler.Object) bool {
		return object.Object.Selectable()
	})
	if filters.ObjectKindRe != "" {
		re := regexp.MustCompile(filters.ObjectKindRe)
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

	return lo.Filter(objects, func(object compiler.Object, _ int) bool {
		for _, filter := range filterChain {
			if !filter(object) {
				return false
			}
		}
		return true
	})
}
