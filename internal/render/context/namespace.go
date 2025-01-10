package context

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"path"
	"slices"
	"strings"
)

type RenderNameDefinition struct {
	Object common.GolangType
	Selection common.ConfigSelectionItem
	// Actual is true when this definition is the actual definition of the object, and false when it is a deferred definition.
	Actual bool
}

type RenderNamespace struct {
	definitions []RenderNameDefinition
	names []string
}

func (s *RenderNamespace) AddObject(obj common.GolangType, selection common.ConfigSelectionItem, actual bool) {
	found := lo.ContainsBy(s.definitions, func(def RenderNameDefinition) bool {
		return def.Object == obj && def.Actual == actual
	})
	if !found {
		s.definitions = append(s.definitions, RenderNameDefinition{Object: obj, Selection: selection, Actual: actual})
	}
}

func (s *RenderNamespace) AddName(name string) {
	if !lo.Contains(s.names, name) {
		s.names = append(s.names, name)
	}
}

func (s *RenderNamespace) FindObject(obj common.GolangType) (RenderNameDefinition, bool) {
	found := lo.Filter(s.definitions, func(def RenderNameDefinition, _ int) bool {
		return def.Object == obj
	})
	// Return the "actual" definition first, if any
	slices.SortFunc(found, func(a, b RenderNameDefinition) int {
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

func (s *RenderNamespace) FindName(name string) bool {
	return lo.Contains(s.names, name)
}

func (s *RenderNamespace) Clone() RenderNamespace {
	return RenderNamespace{
		definitions: append([]RenderNameDefinition(nil), s.definitions...),
		names: append([]string(nil), s.names...),
	}
}

func (s *RenderNamespace) String() string {
	defs := strings.Join(lo.Map(s.definitions, func(item RenderNameDefinition, _ int) string {
		return fmt.Sprintf("[%[1]p] %[1]s", item.Object)
	}), "; ")
	return fmt.Sprintf("names: %s | defs: %s", strings.Join(s.names, "; "), defs)
}

// qualifiedToImport converts the qual* template function parameters to qualified name and import package path.
// And also it returns the package name (the last part of the package path).
func qualifiedToImport(parts []string) (pkgPath string, pkgName string, name string) {
	// parts["a"] -> ["a", "a", ""]
	// parts["", "a"] -> ["", "", "a"]
	// parts["a.x"] -> ["a", "a", "x"]
	// parts["a/b/c"] -> ["a/b/c", "c", ""]
	// parts["a", "x"] -> ["a", "a", "x"]
	// parts["a/b.c", "x"] -> ["a/b.c", "bc", "x"]
	// parts["n", "d", "a/b.c", "x"] -> ["n/d/a/b.c-e", "b.c-e", "x"]
	switch len(parts) {
	case 0:
		panic("Empty parameters, at least one is required")
	case 1:
		pkgPath = parts[0]
	default:
		pkgPath = path.Join(parts[:len(parts)-1]...) + "." + parts[len(parts)-1]
	}
	if pos := strings.LastIndex(pkgPath, "."); pos >= 0 {
		name = pkgPath[pos+1:]
		pkgPath = pkgPath[:pos]
	}
	pkgName = pkgPath
	if pos := strings.LastIndex(pkgPath, "/"); pos >= 0 {
		pkgName = pkgPath[pos+1:]
	}
	return
}