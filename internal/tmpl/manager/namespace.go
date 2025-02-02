package manager

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"slices"
	"strings"
)

type NamespaceManager struct {
	types               []RenderTypeDefinition
	names               []string
}

func (s *NamespaceManager) AddType(obj common.GolangType, selection common.ConfigSelectionItem, actual bool) {
	found := lo.ContainsBy(s.types, func(def RenderTypeDefinition) bool {
		return def.Object == obj && def.Actual == actual
	})
	if !found {
		s.types = append(s.types, RenderTypeDefinition{Object: obj, Selection: selection, Actual: actual})
	}
}

func (s *NamespaceManager) AddName(name string) {
	if !lo.Contains(s.names, name) {
		s.names = append(s.names, name)
	}
}

func (s *NamespaceManager) FindType(obj common.GolangType) (RenderTypeDefinition, bool) {
	found := lo.Filter(s.types, func(def RenderTypeDefinition, _ int) bool {
		return def.Object == obj
	})
	// Return the "actual" definition first, if any
	slices.SortFunc(found, func(a, b RenderTypeDefinition) int {
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

func (s *NamespaceManager) FindName(name string) bool {
	return lo.Contains(s.names, name)
}

func (s *NamespaceManager) Clone() *NamespaceManager {
	return &NamespaceManager{
		types:               slices.Clone(s.types),
		names:               slices.Clone(s.names),
	}
}

func (s *NamespaceManager) String() string {
	defs := strings.Join(lo.Map(s.types, func(item RenderTypeDefinition, _ int) string {
		return fmt.Sprintf("[%[1]p] %[1]s", item.Object)
	}), "; ")
	return fmt.Sprintf("names: %s | defs: %s", strings.Join(s.names, "; "), defs)
}

type RenderTypeDefinition struct {
	Object common.GolangType
	Selection common.ConfigSelectionItem
	// Actual is true when this definition is the actual definition of the object, and false when it is a deferred definition.
	// E.g. defined by `def` template function
	Actual bool
}
