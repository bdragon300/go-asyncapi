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

func (s *ImportsList) addImport(pkgPath string, pkgName string) string {
	if s.imports == nil {
		s.imports = make(map[string]common.ImportItem)
	}
	if _, ok := s.imports[pkgPath]; !ok {
		res := common.ImportItem{PackageName: pkgName, PackagePath: pkgPath}
		// Generate alias if the package with the same name already imported, or its name is not a valid Go identifier (e.g. "go-asyncapi")
		namesakes := lo.Filter(lo.Entries(s.imports), func(item lo.Entry[string, common.ImportItem], _ int) bool {
			return item.Key != pkgPath && item.Value.PackageName == pkgName
		})
		if len(namesakes) > 0 || !token.IsIdentifier(pkgName) {
			// Generate a new alias to avoid package name conflict
			res.Alias = fmt.Sprintf("%s%d", utils.ToGolangName(pkgName, false), len(namesakes)+1)
		}
		s.imports[pkgPath] = res
	}

	if v := s.imports[pkgPath]; v.Alias != "" {
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
