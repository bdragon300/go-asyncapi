package manager

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"slices"
	"strings"
)

type NamespaceManager struct {
	types               []NamespaceTypeItem
	names               []string
}

func (s *NamespaceManager) DefineType(obj common.GolangType, renderManager *TemplateRenderManager, actual bool) {
	found := lo.ContainsBy(s.types, func(def NamespaceTypeItem) bool {
		return def.Object == obj && def.Actual == actual
	})
	if !found {
		s.types = append(s.types, NamespaceTypeItem{
			Object: obj,
			Selection: renderManager.CurrentSelection,
			Actual: actual,
			FileName: renderManager.FileName,
			PackageName: renderManager.PackageName,
		})
	}
}

func (s *NamespaceManager) DefineName(name string) {
	if !lo.Contains(s.names, name) {
		s.names = append(s.names, name)
	}
}

func (s *NamespaceManager) FindType(obj common.GolangType) (NamespaceTypeItem, bool) {
	found := lo.Filter(s.types, func(def NamespaceTypeItem, _ int) bool {
		return def.Object == obj
	})
	// Return the "actual" definition first, if any
	slices.SortFunc(found, func(a, b NamespaceTypeItem) int {
		switch {
		case a.Actual && !b.Actual:
			return 1
		case !a.Actual && b.Actual:
			return -1
		}
		return 0
	})

	return lo.Last(found)
}

func (s *NamespaceManager) IsNameDefined(name string) bool {
	return lo.Contains(s.names, name)
}

func (s *NamespaceManager) Clone() *NamespaceManager {
	return &NamespaceManager{
		types:               slices.Clone(s.types),
		names:               slices.Clone(s.names),
	}
}

func (s *NamespaceManager) String() string {
	defs := strings.Join(lo.Map(s.types, func(item NamespaceTypeItem, _ int) string {
		return fmt.Sprintf("[%[1]p] %[1]s", item.Object)
	}), "; ")
	return fmt.Sprintf("names: %s | defs: %s", strings.Join(s.names, "; "), defs)
}

type NamespaceTypeItem struct {
	Object common.GolangType
	Selection common.ConfigSelectionItem
	FileName string
	PackageName string
	// Actual is true when this definition is the actual definition of the object, and false when it is a deferred definition.
	// E.g. defined by `def` template function
	Actual bool
}
