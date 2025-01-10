package context

import (
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"path"
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
