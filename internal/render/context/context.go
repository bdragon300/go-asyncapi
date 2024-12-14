package context

import (
	"cmp"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"go/token"
	"path"
	"slices"
	"strings"
	"unicode"
)

// TODO: add object path?
type RenderContextImpl struct {
	RenderOpts     common.RenderOpts
	CurrentSelectionConfig common.RenderSelectionConfig
	FileHeader *RenderFileHeader
	Object     common.CompileObject
}

func (c *RenderContextImpl) RuntimeModule(subPackage string) string {
	return path.Join(c.RenderOpts.RuntimeModule, subPackage)
}

func (c *RenderContextImpl) QualifiedName(parts ...string) string {
	pkgPath, pkgName, n := qualifiedToImport(parts)
	var name string
	if n != "" {
		name = utils.ToGolangName(n, unicode.IsUpper(rune(n[0])))
	}
	return fmt.Sprintf("%s.%s", c.FileHeader.addImport(pkgPath, pkgName), name)
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
	// Check if the object is defined in the same directory (assuming the directory is equal to package)
	fileDir := path.Dir(defInfo.Selection.File)
	if fileDir == path.Dir(c.CurrentSelectionConfig.File) {
		return "", nil // Object is defined in the current package, its name doesn't require a package name
	}

	parentDir := path.Dir(fileDir)
	pkgPath := path.Join(c.RenderOpts.ImportBase, parentDir, defInfo.Selection.Package)
	return c.FileHeader.addImport(pkgPath, defInfo.Selection.Package), nil
}

func (c *RenderContextImpl) QualifiedRuntimeName(parts ...string) string {
	p := append([]string{c.RenderOpts.RuntimeModule}, parts...)
	pkgPath, pkgName, n := qualifiedToImport(p)
	var name string
	if n != "" {
		name = utils.ToGolangName(n, unicode.IsUpper(rune(n[0])))
	}
	return fmt.Sprintf("%s.%s", c.FileHeader.addImport(pkgPath, pkgName), name)
}

func (c *RenderContextImpl) CurrentDefinitionInfo() *common.GolangTypeDefinitionInfo {
	return &common.GolangTypeDefinitionInfo{Selection: c.CurrentSelectionConfig}
}

func (c *RenderContextImpl) CurrentSelection() common.RenderSelectionConfig {
	return c.CurrentSelectionConfig
}

func (c *RenderContextImpl) CurrentObject() common.CompileObject {
	return c.Object
}


func NewRenderFileHeader(packageName string) *RenderFileHeader {
	return &RenderFileHeader{packageName: packageName}
}

type RenderFileHeader struct {
	packageName string
	imports map[string]common.ImportItem
}

func (s *RenderFileHeader) Imports() []common.ImportItem {
	res := lo.Values(s.imports)
	slices.SortFunc(res, func(a, b common.ImportItem) int {
		return cmp.Compare(a.PackagePath, b.PackagePath)
	})
	return res
}

func (s *RenderFileHeader) PackageName() string {
	return s.packageName
}

func (s *RenderFileHeader) addImport(pkgPath string, pkgName string) string {
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