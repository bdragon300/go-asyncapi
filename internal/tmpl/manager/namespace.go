package manager

import (
	"cmp"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"slices"
	"strings"
)

// NamespaceManager manages the template namespace, that is used for conditional rendering functionality in the templates.
// It keeps the rendered definitions of rendered Go types and names that was explicitly defined.
// This functionality could remind the "preprocessor" in C/C++, but for Go templates.
type NamespaceManager struct {
	types               []NamespaceTypeItem
	names               []string
}

// DefineType adds the [common.GolangType] object's definition to the namespace, remembering the current file and package
// from the render manager. The priority sets the definition priority for the same object -- the higher priority wins.
func (s *NamespaceManager) DefineType(obj common.GolangType, renderManager *TemplateRenderManager, priority int) {
	found := lo.ContainsBy(s.types, func(item NamespaceTypeItem) bool {
		return item.Object == obj && item.Priority >= priority
	})
	if !found {
		s.types = append(s.types, NamespaceTypeItem{
			Object:      obj,
			Selection:   renderManager.CurrentSelection,
			Priority:    priority,
			FileName:    renderManager.FileName,
			PackageName: renderManager.PackageName,
		})
	}
}

// DefineName adds the name to the namespace.
func (s *NamespaceManager) DefineName(name string) {
	if !lo.Contains(s.names, name) {
		s.names = append(s.names, name)
	}
}

// FindType searches for definition of the object in the namespace. The function returns the definition with the highest
// priority, if found. Otherwise, returns false.
func (s *NamespaceManager) FindType(obj common.GolangType) (NamespaceTypeItem, bool) {
	found := lo.Filter(s.types, func(def NamespaceTypeItem, _ int) bool {
		return def.Object == obj
	})
	slices.SortFunc(found, func(a, b NamespaceTypeItem) int { return cmp.Compare(a.Priority, b.Priority)})

	return lo.Last(found)
}

// IsNameDefined checks if the name is defined in the namespace.
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
	Object    common.GolangType
	Selection common.ConfigSelectionItem
	FileName string
	PackageName string
	Priority int
}
