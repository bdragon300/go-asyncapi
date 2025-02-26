package manager

import (
	"cmp"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"go/token"
	"maps"
	"slices"
	"strings"
)

// ImportItem represents an import in the generated Go file.
type ImportItem struct {
	Alias       string
	PackageName string
	PackagePath string
}

// ImportsManager manages the imports in the generated Go file.
type ImportsManager struct {
	items map[string]ImportItem
}

// Imports returns the sorted list of imports.
func (s *ImportsManager) Imports() []ImportItem {
	res := lo.Values(s.items)
	slices.SortFunc(res, func(a, b ImportItem) int {
		return cmp.Compare(a.PackagePath, b.PackagePath)
	})
	return res
}

// AddImport adds a new import path to list. Prevents appearing the duplicated, conflicted or invalid imports in list.
//
// The import path to add is passed in the first argument, the package name is extracted from the last part of import
// path. But if the package name argument is set, then it will be used instead.
//
// If import path was already added, the function does nothing.
//
// If the package name conflicts with the package name of the already added import, the function generates a new alias
// for this import. For example, the import [text/template] gets the alias ``template2'' if we already added the import
// [html/template] before.
//
// If the package name is not a valid Go identifier, the function generates a new alias for this import as well.
// For example, the import [github.com/bdragon300/go-asyncapi] gets the alias ``goAsyncapi'', because ``go-asyncapi''
// is not a valid Go identifier.
//
// Function returns the imported name, used to access that package in Go code. It can be the package name
// (for ``import "net/url"'' returns ``url''), or the alias, if any
// (for ``import goAsyncapi "github.com/bdragon300/go-asyncapi"'' returns ``goAsyncapi'').
func (s *ImportsManager) AddImport(importPath string, pkgName string) string {
	if s.items == nil {
		s.items = make(map[string]ImportItem)
	}

	// Suppose that the package name by default is the last part of the import path. But if it's specified, the import
	// path remains the same, but the package name is going to be used in the code.
	// This is because Go treats the import as directory path, but uses the package in namespace.
	// https://stackoverflow.com/questions/43579838/relationship-between-a-package-statement-and-the-directory-of-a-go-file
	if pkgName == "" {
		pkgName = utils.GetPackageName(importPath)
	}

	if _, ok := s.items[importPath]; !ok {
		res := ImportItem{PackageName: pkgName, PackagePath: importPath}
		// Generate a new alias if the package with the same name already imported, or it's not a valid Go identifier (e.g. "go-asyncapi")
		conflicts := lo.Filter(lo.Entries(s.items), func(item lo.Entry[string, ImportItem], _ int) bool {
			return item.Key != importPath && item.Value.PackageName == pkgName
		})
		if len(conflicts) > 0 || !token.IsIdentifier(pkgName) {
			res.Alias = fmt.Sprintf("%s%d", utils.ToGolangName(pkgName, false), len(conflicts)+1)
		}
		s.items[importPath] = res
	}

	if v := s.items[importPath]; v.Alias != "" {
		return v.Alias
	}
	return pkgName
}

func (s *ImportsManager) Clone() *ImportsManager {
	return &ImportsManager{items: maps.Clone(s.items)}
}

func (s *ImportsManager) String() string {
	return strings.Join(lo.Map(s.Imports(), func(item ImportItem, _ int) string {
		if item.Alias != "" {
			return fmt.Sprintf("%s %s", item.Alias, item.PackagePath)
		}
		return item.PackagePath
	}), "; ")
}
