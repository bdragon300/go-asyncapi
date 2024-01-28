package common

import (
	"path"
	"regexp"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

const nameWordSep = "_"

type GolangType interface {
	Renderer
	TypeName() string
}

type CompilationStorage interface {
	AddObject(pkgName string, stack []string, obj Renderer)
	RegisterProtocol(protoName string)
	AddExternalSpecID(specID string)
	AddPromise(p ObjectPromise)
	AddListPromise(p ObjectListPromise)

	SetDefaultContentType(contentType string)
	DefaultContentType() string

	SetActiveServers(servers []string)
	ActiveServers() []string

	SetActiveChannels(channels []string)
	ActiveChannels() []string
}

type CompileOpts struct {
	ChannelOpts         ObjectCompileOpts
	MessageOpts         ObjectCompileOpts
	ModelOpts           ObjectCompileOpts
	ServerOpts          ObjectCompileOpts
	ReusePackages       map[string]string
	NoEncodingPackage   bool
	EnableRemoteRefs    bool
	RuntimeModule       string
	GeneratePublishers  bool
	GenerateSubscribers bool
}

type ObjectCompileOpts struct {
	Enable       bool
	IncludeRegex *regexp.Regexp
	ExcludeRegex *regexp.Regexp
}

func (o ObjectCompileOpts) IsAllowedName(name string) bool {
	switch {
	case !o.Enable:
		return false
	case o.ExcludeRegex != nil && o.ExcludeRegex.MatchString(name):
		return false
	case o.IncludeRegex != nil:
		return o.IncludeRegex.MatchString(name)
	}
	return true
}

type ContextStackItem struct {
	PathItem    string
	Flags       map[SchemaTag]string
	PackageName string
	ObjName     string
}

func NewCompileContext(specID string, compileOpts CompileOpts) *CompileContext {
	res := CompileContext{specID: specID, CompileOpts: compileOpts}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: types.NewLogger("Compilation 🔨"),
	}
	return &res
}

type CompileContext struct {
	Storage     CompilationStorage
	Stack       types.SimpleStack[ContextStackItem]
	Logger      *CompilerLogger
	CompileOpts CompileOpts
	specID      string
}

func (c *CompileContext) PutObject(obj Renderer) {
	pkgName := c.Stack.Top().PackageName
	if pkgName == "" {
		panic("Package name has not been set")
	}
	c.Storage.AddObject(pkgName, c.PathStack(), obj)
}

func (c *CompileContext) PutPromise(p ObjectPromise) {
	refSpecID, _, _ := utils.SplitRefToPathPointer(p.Ref())
	if refSpecID != "" {
		c.Storage.AddExternalSpecID(refSpecID)
	}
	c.Storage.AddPromise(p)
}

func (c *CompileContext) PutListPromise(p ObjectListPromise) {
	c.Storage.AddListPromise(p)
}

func (c *CompileContext) PathRef() string {
	return "#/" + path.Join(c.PathStack()...)
}

func (c *CompileContext) PathStack() []string {
	return lo.Map(c.Stack.Items(), func(item ContextStackItem, _ int) string { return item.PathItem })
}

func (c *CompileContext) SetTopObjName(n string) {
	t := c.Stack.Top()
	t.ObjName = n
	c.Stack.ReplaceTop(t)
}

func (c *CompileContext) CurrentPackage() string {
	return c.Stack.Top().PackageName
}

func (c *CompileContext) RuntimeModule(subPackage string) string {
	return path.Join(c.CompileOpts.RuntimeModule, subPackage)
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

func (c *CompileContext) WithResultsStore(store CompilationStorage) *CompileContext {
	res := CompileContext{
		Storage:     store,
		Stack:       types.SimpleStack[ContextStackItem]{},
		specID:      c.specID,
		Logger:      c.Logger,
		CompileOpts: c.CompileOpts,
	}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: types.NewLogger("Compilation 🔨"),
	}
	return &res
}
