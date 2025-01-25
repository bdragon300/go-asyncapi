package context

import (
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"strings"
	"unicode"
)

// ErrNotDefined is returned when the package location where the object is defined is not known yet.
var ErrNotDefined = errors.New("not defined")

type definable interface {
	ObjectHasDefinition() bool
}

// TODO: add object path?
type RenderContextImpl struct {
	RenderOpts     common.RenderOpts
	CurrentSelectionConfig common.ConfigSelectionItem
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

// QualifiedTypeGeneratedPackage adds the package where the obj is defined to the import list of the current module
// (if it's not there yet). Returns the import alias.
// If import is not needed (obj is defined in current package), returns empty string. If obj was not defined anywhere
// yet, returns ErrNotDefined.
func (c *RenderContextImpl) QualifiedTypeGeneratedPackage(obj common.GolangType) (string, error) {
	defInfo, defined := c.PackageNamespace.FindType(obj)
	if !defined {
		if v, ok := obj.(definable); ok && v.ObjectHasDefinition() { // TODO: replace to Selectable?
			return "", ErrNotDefined
		}
		return "", nil // Type is not supposed to be defined in the generated code (e.g. Go built-in types)
	}

	// Use the package path from reuse config if it is defined
	if defInfo.Selection.ReusePackagePath != "" {
		return c.Imports.addImport(defInfo.Selection.ReusePackagePath, ""), nil
	}

	// Check if the object is defined in the same directory (assuming the directory is equal to package)
	fileDir := path.Dir(defInfo.Selection.Render.File)
	if fileDir == path.Dir(c.CurrentSelectionConfig.Render.File) {
		return "", nil // Object is defined in the current package, its name doesn't require a package name
	}

	pkgPath := path.Join(c.RenderOpts.ImportBase, fileDir)
	return c.Imports.addImport(pkgPath, defInfo.Selection.Render.Package), nil
}

func (c *RenderContextImpl) QualifiedImplementationGeneratedPackage(obj common.ImplementationObject) (string, error) {
	defInfo, found := lo.Find(c.PackageNamespace.Implementations(), func(def RenderImplementationDefinition) bool {
		return def.Object == obj
	})
	if !found {
		return "", ErrNotDefined
	}

	// Use the package path from reuse config if it is defined
	if defInfo.Object.Config.ReusePackagePath != "" {
		return c.Imports.addImport(defInfo.Object.Config.ReusePackagePath, ""), nil
	}

	if defInfo.Directory == path.Dir(c.CurrentSelectionConfig.Render.File) {
		return "", nil // Object is defined in the current package, its name doesn't require a package name
	}

	pkgPath := path.Join(c.RenderOpts.ImportBase, defInfo.Directory)
	return c.Imports.addImport(pkgPath, defInfo.Object.Config.Package), nil
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

func (c *RenderContextImpl) CurrentSelection() common.ConfigSelectionItem {
	return c.CurrentSelectionConfig
}

func (c *RenderContextImpl) GetObject() common.CompileObject {
	return c.Object
}

func (c *RenderContextImpl) Package() string {
	return c.PackageName
}

func (c *RenderContextImpl) DefineTypeInNamespace(obj common.GolangType, selection common.ConfigSelectionItem, actual bool) {
	c.PackageNamespace.AddType(obj, selection, actual)
}

func (c *RenderContextImpl) TypeDefinedInNamespace(obj common.GolangType) bool {
	v, found := c.PackageNamespace.FindType(obj)
	return found && v.Actual
}

func (c *RenderContextImpl) DefineNameInNamespace(name string) {
	c.PackageNamespace.AddName(name)
}

func (c *RenderContextImpl) NameDefinedInNamespace(name string) bool {
	return c.PackageNamespace.FindName(name)
}

func (c *RenderContextImpl) FindImplementationInNamespace(protocol string) (common.ImplementationObject, bool) {
	implDef, found := lo.Find(c.PackageNamespace.Implementations(), func(def RenderImplementationDefinition) bool {
		return def.Object.Manifest.Protocol == protocol
	})
	if !found {
		return common.ImplementationObject{}, false
	}

	return implDef.Object, true
}

// qualifiedToImport converts the qual* template function parameters to qualified name and import package path.
// And also it returns the package name (the last part of the package path).
func qualifiedToImport(exprParts []string) (pkgPath string, pkgName string, name string) {
	// exprParts["a"] -> ["a", "a", ""]
	// exprParts["", "a"] -> ["", "", "a"]
	// exprParts["a.x"] -> ["a", "a", "x"]
	// exprParts["a/b/c"] -> ["a/b/c", "c", ""]
	// exprParts["a", "x"] -> ["a", "a", "x"]
	// exprParts["a/b.c", "x"] -> ["a/b.c", "bc", "x"]
	// exprParts["n", "d", "a/b.c", "x"] -> ["n/d/a/b.c-e", "b.c-e", "x"]
	switch len(exprParts) {
	case 0:
		panic("Empty parameters, at least one is required")
	case 1:
		pkgPath = exprParts[0]
	default:
		pkgPath = path.Join(exprParts[:len(exprParts)-1]...) + "." + exprParts[len(exprParts)-1]
	}
	// Split the whole expression into package path and name.
	// The name is the sequence after the last dot (package path can contain dots in last part).
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