package common

import (
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/samber/lo"
)

type GolangType interface {
	Renderer
	TypeName() string
}

const nameWordSep = "_"

type CompilationObjectsStore interface {
	Add(pkgName string, stack []string, obj Renderer)
	AddProtocol(protoName string)
	AddRemoteSpecID(specID string)

	SetDefaultContentType(contentType string)
	DefaultContentType() string
}

type Linker interface {
	AddPromise(p ObjectPromise, specID string)
	AddListPromise(p ObjectListPromise, specID string)
}

type ContextStackItem struct {
	Path        string // TODO: rename to Key or smth. This is a path item actually
	Flags       map[SchemaTag]string
	PackageName string
	ObjName     string
}

func NewCompileContext(linker Linker, specID string) *CompileContext {
	res := CompileContext{
		specID: specID,
		linker: linker,
	}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: types.NewLogger("Compilation 🔨"),
	}
	return &res
}

type CompileContext struct {
	ObjectsStore CompilationObjectsStore
	Stack        types.SimpleStack[ContextStackItem]
	Logger       *CompilerLogger
	specID       string
	linker       Linker
}

func (c *CompileContext) PutPromise(p ObjectPromise) {
	refSpecID, _ := utils.SplitSpecPath(p.Ref())
	if utils.IsRemoteSpecID(refSpecID) {
		c.ObjectsStore.AddRemoteSpecID(refSpecID)
	}
	c.linker.AddPromise(p, c.specID)
}

func (c *CompileContext) PutListPromise(p ObjectListPromise) {
	c.linker.AddListPromise(p, c.specID)
}

func (c *CompileContext) PutObject(obj Renderer) {
	pkgName := c.Stack.Top().PackageName
	if pkgName == "" {
		panic("Package name has not been set")
	}
	c.ObjectsStore.Add(pkgName, c.PathStack(), obj)
}

func (c *CompileContext) PathRef() string {
	return "#/" + path.Join(c.PathStack()...)
}

func (c *CompileContext) PathStack() []string {
	return lo.Map(c.Stack.Items(), func(item ContextStackItem, _ int) string { return item.Path })
}

func (c *CompileContext) SetTopObjName(n string) {
	t := c.Stack.Top()
	t.ObjName = n
	c.Stack.ReplaceTop(t)
}

func (c *CompileContext) TopPackageName() string {
	return c.Stack.Top().PackageName
}

func (c *CompileContext) RuntimePackage(subPackage string) string {
	return path.Join(RunPackagePath, subPackage)
}

func (c *CompileContext) GenerateObjName(name, suffix string) string {
	if name == "" {
		// Join name from usernames, set earlier by SetTopObjName (if any)
		items := lo.FilterMap(c.Stack.Items(), func(item ContextStackItem, index int) (string, bool) {
			return item.ObjName, item.ObjName != ""
		})
		// Otherwise join name from current spec path
		if len(items) == 0 {
			items = c.PathStack()
		}
		name = strings.Join(items, nameWordSep)
	}
	return utils.ToGolangName(name, true) + suffix
}

func (c *CompileContext) WithResultsStore(store CompilationObjectsStore) *CompileContext {
	res := CompileContext{
		ObjectsStore: store,
		Stack:        types.SimpleStack[ContextStackItem]{},
		specID:       c.specID,
		linker:       c.linker,
		Logger:       c.Logger,
	}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: types.NewLogger("Compilation 🔨"),
	}
	return &res
}
