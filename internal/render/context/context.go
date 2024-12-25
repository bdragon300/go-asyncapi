package context

import (
	"cmp"
	"errors"
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

// ErrNotDefined is returned when we try to get a package in the generated code for an object, but the
// definition of this object has not been rendered, therefore its location is unknown yet.
var ErrNotDefined = errors.New("not defined")

type definable interface {
	ObjectHasDefinition() bool
}

// TODO: add object path?
type RenderContextImpl struct {
	RenderOpts     common.RenderOpts
	CurrentSelectionConfig common.RenderSelectionConfig
	PackageName      string
	Imports          *ImportsList
	PackageNamespace *RenderNamespace
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
	return fmt.Sprintf("%s.%s", c.Imports.addImport(pkgPath, pkgName), name)
}

// QualifiedGeneratedPackage checks if the object is in the generated package of CurrentSelectionConfig and returns
// the package name if it is. If the object is not in the generated package, it returns an empty string.
func (c *RenderContextImpl) QualifiedGeneratedPackage(obj common.GolangType) (string, error) {
	defInfo, defined := c.PackageNamespace.FindObject(obj)
	if !defined {
		if v, ok := obj.(definable); ok && v.ObjectHasDefinition() { // TODO: replace to Selectable?
			return "", ErrNotDefined
		}
		return "", nil // Type is not supposed to be defined in the generated code (including Go built-in types)
	}

	// Use the package path from reuse config if it is defined
	if defInfo.Selection.ReusePackagePath != "" {
		_, packageName := path.Split(defInfo.Selection.ReusePackagePath)
		return c.Imports.addImport(defInfo.Selection.ReusePackagePath, packageName), nil
	}

	// Check if the object is defined in the same directory (assuming the directory is equal to package)
	fileDir := path.Dir(defInfo.Selection.Render.File)
	if fileDir == path.Dir(c.CurrentSelectionConfig.Render.File) {
		return "", nil // Object is defined in the current package, its name doesn't require a package name
	}

	parentDir := path.Dir(fileDir)
	pkgPath := path.Join(c.RenderOpts.ImportBase, parentDir, defInfo.Selection.Render.Package)
	return c.Imports.addImport(pkgPath, defInfo.Selection.Render.Package), nil
}

func (c *RenderContextImpl) QualifiedRuntimeName(parts ...string) string {
	p := append([]string{c.RenderOpts.RuntimeModule}, parts...)
	pkgPath, pkgName, n := qualifiedToImport(p)
	var name string
	if n != "" {
		name = utils.ToGolangName(n, unicode.IsUpper(rune(n[0])))
	}
	return fmt.Sprintf("%s.%s", c.Imports.addImport(pkgPath, pkgName), name)
}

func (c *RenderContextImpl) CurrentSelection() common.RenderSelectionConfig {
	return c.CurrentSelectionConfig
}

func (c *RenderContextImpl) GetObjectName(obj common.Renderable) string {
	type renderableWrapper interface {
		UnwrapRenderable() common.Renderable
	}

	res := obj.Name()
	// Take the alternate name from CurrentObject (if any), if the CurrentObject is the RenderablePromise,
	// and it points to the same object as the one was passed as argument.
	currentObj := c.Object.Renderable
	if p, ok := currentObj.(renderableWrapper); ok {
		currentObj = p.UnwrapRenderable()
	}
	if currentObj == obj {
		res = c.Object.Name()
	}

	return res
}

func (c *RenderContextImpl) Package() string {
	return c.PackageName
}

func (c *RenderContextImpl) DefineTypeInNamespace(obj common.GolangType, selection common.RenderSelectionConfig, actual bool) {
	c.PackageNamespace.AddObject(obj, selection, actual)
}

func (c *RenderContextImpl) TypeDefinedInNamespace(obj common.GolangType) bool {
	v, found := c.PackageNamespace.FindObject(obj)
	return found && v.Actual
}

func (c *RenderContextImpl) DefineNameInNamespace(name string) {
	c.PackageNamespace.AddName(name)
}

func (c *RenderContextImpl) NameDefinedInNamespace(name string) bool {
	return c.PackageNamespace.FindName(name)
}

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

type RenderNameDefinition struct {
	Object common.GolangType
	Selection common.RenderSelectionConfig
	// Actual is true when this definition is the actual definition of the object, and false when it is a deferred definition.
	Actual bool
}


type RenderNamespace struct {
	definitions []RenderNameDefinition
	names []string
}

func (s *RenderNamespace) AddObject(obj common.GolangType, selection common.RenderSelectionConfig, actual bool) {
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