package common

import (
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/samber/lo"
)

type GolangType interface {
	Assembler
	TypeName() string
}

const nameWordSep = "_"

type ContextStackItem struct {
	Path        string
	Flags       map[SchemaTag]string
	PackageName string
	ObjName     string
}

type CompileContext struct {
	Packages map[string]*Package
	Stack    SimpleStack[ContextStackItem]
	Linker   Linker
}

func (c *CompileContext) PutToCurrentPkg(obj Assembler) {
	pkgName := c.Stack.Top().PackageName
	if pkgName == "" {
		panic("Package name has not been set")
	}
	pkg, ok := c.Packages[pkgName]
	if !ok {
		pkg = &Package{}
		c.Packages[c.Stack.Top().PackageName] = pkg
	}
	pkg.Put(obj, c.PathStack())
}

func (c *CompileContext) PathStack() []string {
	return lo.Map(c.Stack.Items(), func(item ContextStackItem, _ int) string { return item.Path })
}

func (c *CompileContext) PathRef() string {
	return "#/" + path.Join(c.PathStack()...)
}

func (c *CompileContext) SetTopObjName(n string) {
	t := c.Stack.Top()
	t.ObjName = n
	c.Stack.replaceTop(t)
}

func (c *CompileContext) TopPackageName() string {
	return c.Stack.Top().PackageName
}

func (c *CompileContext) RuntimePackage(protoName string) string {
	return path.Join(RunPackagePath, protoName)
}

func (c *CompileContext) GenerateObjName(name, suffix string) string {
	if name == "" {
		// Join name from user names, set earlier by SetTopObjName (if any)
		items := lo.FilterMap(c.Stack.Items(), func(item ContextStackItem, index int) (string, bool) {
			return item.ObjName, item.ObjName != ""
		})
		// Otherwise join name from current spec path
		if len(items) == 0 {
			items = c.PathStack()
		}
		name = strings.Join(items, nameWordSep)
	}
	if suffix != "" {
		name += nameWordSep + suffix
	}
	return utils.ToGolangName(name, true)
}
