package context

import (
	"cmp"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"go/token"
	"slices"
	"strings"
)

type ImportsList struct { //TODO: rename
	imports map[string]common.ImportItem
}

func (s *ImportsList) Imports() []common.ImportItem {
	res := lo.Values(s.imports)
	slices.SortFunc(res, func(a, b common.ImportItem) int {
		return cmp.Compare(a.PackagePath, b.PackagePath)
	})
	return res
}

func (s *ImportsList) addImport(importPath string, pkgName string) string {
	if s.imports == nil {
		s.imports = make(map[string]common.ImportItem)
	}

	// Suppose that the package name by default is the last part of the import path. But if it's specified, the import
	// path remains the same, but the package name is going to be used in the code.
	// This is because Go treats the import as directory path, but uses the package in namespace.
	// https://stackoverflow.com/questions/43579838/relationship-between-a-package-statement-and-the-directory-of-a-go-file
	if pkgName == "" {
		pkgName = utils.GetPackageName(importPath)
	}

	if _, ok := s.imports[importPath]; !ok {
		res := common.ImportItem{PackageName: pkgName, PackagePath: importPath}
		// Generate a new alias if the package with the same name already imported, or it's not a valid Go identifier (e.g. "go-asyncapi")
		conflicts := lo.Filter(lo.Entries(s.imports), func(item lo.Entry[string, common.ImportItem], _ int) bool {
			return item.Key != importPath && item.Value.PackageName == pkgName
		})
		if len(conflicts) > 0 || !token.IsIdentifier(pkgName) {
			res.Alias = fmt.Sprintf("%s%d", utils.ToGolangName(pkgName, false), len(conflicts)+1)
		}
		s.imports[importPath] = res
	}

	if v := s.imports[importPath]; v.Alias != "" {
		return v.Alias // Return alias
	}
	return pkgName
}

func (s *ImportsList) Clone() ImportsList {
	return ImportsList{imports: lo.Assign(s.imports)}
}

func (s *ImportsList) String() string {
	return strings.Join(lo.Map(s.Imports(), func(item common.ImportItem, _ int) string {
		if item.Alias != "" {
			return fmt.Sprintf("%s %s", item.Alias, item.PackagePath)
		}
		return item.PackagePath
	}), "; ")
}
