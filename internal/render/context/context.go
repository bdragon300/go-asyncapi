package context

import (
	"cmp"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"slices"
	"strings"
	"unicode"
)

// TODO: add object path?
type RenderContextImpl struct {
	RenderOpts     common.RenderOpts
	CurrentSelectionConfig common.RenderSelectionConfig
	imports map[string]common.ImportItem // Key: package path
}

func (c *RenderContextImpl) RuntimeModule(subPackage string) string {
	return path.Join(c.RenderOpts.RuntimeModule, subPackage)
}

func (c *RenderContextImpl) QualifiedName(parts ...string) string {
	pkgPath, pkgName, name := qualifiedToImport(parts)
	return fmt.Sprintf("%s.%s", c.importPackage(pkgPath, pkgName), utils.ToGolangName(name, unicode.IsUpper(rune(name[0]))))
}

// QualifiedGeneratedPackage checks if the object is in the generated package of CurrentSelectionConfig and returns
// the package name if it is. If the object is not in the generated package, it returns an empty string.
func (c *RenderContextImpl) QualifiedGeneratedPackage(obj common.GolangType) (string, error) {
	defInfo, err := obj.DefinitionInfo()
	if err != nil {
		return "", fmt.Errorf("object %s: %w", obj, err)
	}
	if defInfo == nil {
		return "", nil // Object has no definition (e.g. Go built-in types)
	}
	d := path.Dir(defInfo.Selection.File)
	if d == path.Dir(c.CurrentSelectionConfig.File) {
		return "", nil // Object is defined in the current package, it name doesn't require a package name
	}

	b, _ := path.Split(d)
	pkgPath := path.Join(c.RenderOpts.ImportBase, b, defInfo.Selection.Package)
	return c.importPackage(pkgPath, defInfo.Selection.Package), nil
}

func (c *RenderContextImpl) QualifiedRuntimeName(parts ...string) string {
	p := append([]string{c.RenderOpts.ImportBase}, parts...)
	pkgPath, pkgName, name := qualifiedToImport(p)
	return fmt.Sprintf("%s.%s", c.importPackage(pkgPath, pkgName), utils.ToGolangName(name, unicode.IsUpper(rune(name[0]))))
}

func (c *RenderContextImpl) CurrentDefinitionInfo() *common.GolangTypeDefinitionInfo {
	return &common.GolangTypeDefinitionInfo{Selection: c.CurrentSelectionConfig}
}

func (c *RenderContextImpl) CurrentSelection() common.RenderSelectionConfig {
	return c.CurrentSelectionConfig
}

func (c *RenderContextImpl) Imports() []common.ImportItem {
	res := lo.Values(c.imports)
	slices.SortFunc(res, func(a, b common.ImportItem) int {
		return cmp.Compare(a.PackagePath, b.PackagePath)
	})
	return res
}

func (c *RenderContextImpl) importPackage(pkgPath string, pkgName string) string {
	if c.imports == nil {
		c.imports = make(map[string]common.ImportItem)
	}
	if _, ok := c.imports[pkgPath]; !ok {
		res := common.ImportItem{PackageName: pkgName, PackagePath: pkgPath}
		// Find imports with the same package name
		namesakes := lo.Filter(lo.Entries(c.imports), func(item lo.Entry[string, common.ImportItem], _ int) bool {
			return item.Key != pkgPath && item.Value.PackageName == pkgName
		})
		if len(namesakes) > 0 {
			res.Alias = fmt.Sprintf("%s%d", pkgName, len(namesakes)+1)  // Generate a new alias to avoid package name conflict
		}
		c.imports[pkgPath] = res
	}

	if v := c.imports[pkgPath]; v.Alias != "" {
		return v.Alias // Return alias
	}
	return pkgName
}

//// LogStartRender is typically called at the beginning of a D or U method and logs that the
//// object is started to be rendered. It also logs the object's name, type, and the current package.
//// Every call to LogStartRender should be followed by a call to LogFinishRender which logs that the object is finished to be
//// rendered.
//func (c *RenderContext) LogStartRender(kind, pkg, name, mode string, directRendering bool, args ...any) {
//	l := c.Logger
//	args = append(args, "pkg", c.CurrentPackage, "mode", mode)
//	if pkg != "" {
//		name = pkg + "." + name
//	}
//	name = kind + " " + name
//	if c.logCallLvl > 0 {
//		name = fmt.Sprintf("%s> %s", strings.Repeat("-", c.logCallLvl), name) // Ex: prefix: --> Message...
//	}
//	if directRendering && mode == "definition" {
//		l.Debug(name, args...)
//	} else {
//		l.Trace(name, args...)
//	}
//	c.logCallLvl++
//}
//
//func (c *RenderContext) LogFinishRender() {
//	if c.logCallLvl > 0 {
//		c.logCallLvl--
//	}
//}

// qualifiedToImport converts the qual* template function parameters to qualified name and import package path.
// And also it returns the package name (the last part of the package path).
func qualifiedToImport(parts []string) (pkgPath string, pkgName string, name string) {
	// parts["a"] -> ["a", "a", ""]
	// parts["", "a"] -> ["", "", "a"]
	// parts["a.x"] -> ["a", "a", "x"]
	// parts["a/b/c"] -> ["a/b/c", "c", ""]
	// parts["a", "x"] -> ["a", "a", "x"]
	// parts["a/b.c", "x"] -> ["a/b.c", "bc", "x"]
	// parts["n", "d", "a/b.c", "x"] -> ["n/d/a/b.c", "bc", "x"]
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
	pkgName = strings.ReplaceAll(pkgName, ".", "")
	return
}