package common

import (
	"fmt"
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/samber/lo"
)

type GolangType interface {
	Renderer
	TypeName() string
}

const nameWordSep = "_"

type CompilationResultsStore interface {
	Add(pkgName string, stack []string, obj Renderer)
	AddProtocol(protoName string)

	SetDefaultContentType(contentType string)
	DefaultContentType() string
}

type Linker interface {
	Add(query LinkQuerier)
	AddMany(query ListQuerier)
}

type ContextStackItem struct {
	Path        string
	Flags       map[SchemaTag]string
	PackageName string
	ObjName     string
}

func NewCompileContext(linker Linker) *CompileContext {
	res := CompileContext{
		Linker: linker,
	}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: NewLogger("Compilation 🔨"),
	}
	return &res
}

type CompileContext struct {
	ResultsStore CompilationResultsStore
	Stack        SimpleStack[ContextStackItem]
	Linker       Linker
	Logger       *CompilerLogger
}

func (c *CompileContext) PutToCurrentPkg(obj Renderer) {
	pkgName := c.Stack.Top().PackageName
	if pkgName == "" {
		panic("Package name has not been set")
	}
	c.ResultsStore.Add(pkgName, c.PathStack(), obj)
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
	c.Stack.replaceTop(t)
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

func (c *CompileContext) WithResultsStore(store CompilationResultsStore) *CompileContext {
	return &CompileContext{
		ResultsStore: store,
		Stack:        SimpleStack[ContextStackItem]{},
		Linker:       c.Linker,
		Logger:       c.Logger,
	}
}

type CompilerLogger struct {
	ctx       *CompileContext
	logger    *Logger
	callLevel int
}

func (c *CompilerLogger) Fatal(msg string, err error) {
	if err != nil {
		c.logger.Error(msg, "err", err, "path", c.ctx.PathRef())
	}
	c.logger.Error(msg, "path", c.ctx.PathRef())
}

func (c *CompilerLogger) Warn(msg string, args ...any) {
	args = append(args, "path", c.ctx.PathRef())
	c.logger.Warn(msg, args...)
}

func (c *CompilerLogger) Info(msg string, args ...any) {
	args = append(args, "path", c.ctx.PathRef())
	c.logger.Info(msg, args...)
}

func (c *CompilerLogger) Debug(msg string, args ...any) {
	l := c.logger
	if c.callLevel > 0 {
		msg = fmt.Sprintf("%s> %s", strings.Repeat("-", c.callLevel), msg) // Ex: prefix: --> Message...
	}
	args = append(args, "path", c.ctx.PathRef())
	l.Debug(msg, args...)
}

func (c *CompilerLogger) Trace(msg string, args ...any) {
	l := c.logger
	if c.callLevel > 0 {
		msg = fmt.Sprintf("%s> %s", strings.Repeat("-", c.callLevel), msg) // Ex: prefix: --> Message...
	}
	args = append(args, "path", c.ctx.PathRef())
	l.Trace(msg, args...)
}

func (c *CompilerLogger) NextCallLevel() {
	c.callLevel++
}

func (c *CompilerLogger) PrevCallLevel() {
	if c.callLevel > 0 {
		c.callLevel--
	}
}
