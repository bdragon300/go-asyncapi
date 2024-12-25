package common

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"path"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/specurl"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

const nameWordSep = "_"

type CompileObject struct {
	Renderable
	ObjectURL specurl.URL
}

type GolangType interface {
	Renderable
	GoTemplate() string
	// Addressable returns true if this type can have pointer in its definition
	Addressable() bool
	// IsPointer returns true if this type is a pointer
	IsPointer() bool
}

type CompilationStorage interface {
	AddObject(obj CompileObject)
	RegisterProtocol(protoName string)
	AddExternalSpecPath(specPath *specurl.URL)
	AddPromise(p ObjectPromise)
	AddListPromise(p ObjectListPromise)
	SpecObjectURL() specurl.URL
}

type CompileOpts struct {
	AllowRemoteRefs     bool
	RuntimeModule       string
	GeneratePublishers  bool
	GenerateSubscribers bool
}

type ContextStackItem struct {
	PathItem       string
	Flags          map[SchemaTag]string
	RegisteredName string
}

func NewCompileContext(specPath *specurl.URL, compileOpts CompileOpts) *CompileContext {
	res := CompileContext{specRef: specPath, CompileOpts: compileOpts}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: log.GetLogger(log.LoggerPrefixCompilation),
	}
	return &res
}

type CompileContext struct {
	Storage     CompilationStorage
	Stack       types.SimpleStack[ContextStackItem]
	Logger      *CompilerLogger
	CompileOpts CompileOpts
	specRef     *specurl.URL
}

func (c *CompileContext) PutObject(obj Renderable) {
	c.Logger.Debug("Built", "object", obj.String(), "addr", fmt.Sprintf("%p", obj), "type", fmt.Sprintf("%T", obj))
	c.Storage.AddObject(CompileObject{Renderable: obj, ObjectURL: c.CurrentObjectURL()})
}

func (c *CompileContext) PutPromise(p ObjectPromise) {
	ref := specurl.Parse(p.Ref())
	if ref.IsExternal() {
		c.Storage.AddExternalSpecPath(ref)
	}
	c.Storage.AddPromise(p)
}

func (c *CompileContext) PutListPromise(p ObjectListPromise) {
	c.Storage.AddListPromise(p)
}

// PathStackRef returns a path to the current stack as a reference. NOTE: the reference is URL-encoded.
func (c *CompileContext) PathStackRef(joinParts ...string) string {
	parts := c.pathStack()
	if len(joinParts) > 0 {
		parts = append(parts, joinParts...)
	}
	return specurl.BuildRef(parts...)
}

func (c *CompileContext) pathStack() []string {
	return lo.Map(c.Stack.Items(), func(item ContextStackItem, _ int) string { return item.PathItem })
}

func (c *CompileContext) CurrentObjectURL() specurl.URL {
	u := c.Storage.SpecObjectURL()
	u.Pointer = c.pathStack()
	return u
}

func (c *CompileContext) RegisterNameTop(n string) {  // TODO: rework and remove this method?
	t := c.Stack.Top()
	t.RegisteredName = n
	c.Stack.ReplaceTop(t)
}

func (c *CompileContext) RuntimeModule(subPackage string) string {
	return path.Join(c.CompileOpts.RuntimeModule, subPackage)
}

func (c *CompileContext) GenerateObjName(name, suffix string) string {
	if name == "" {
		// Use names of registered object from current stack (that were set by RegisterNameTop call)
		items := lo.FilterMap(c.Stack.Items(), func(item ContextStackItem, _ int) (string, bool) {
			return item.RegisteredName, item.RegisteredName != ""
		})
		// Otherwise if no registered objects in stack, just use path
		if len(items) == 0 {
			items = c.pathStack()
		}
		name = strings.Join(items, nameWordSep)
	}
	return utils.ToGolangName(name, true) + suffix
}

func (c *CompileContext) WithResultsStore(store CompilationStorage) *CompileContext {
	res := CompileContext{
		Storage:     store,
		Stack:       types.SimpleStack[ContextStackItem]{},
		specRef:     c.specRef,
		Logger:      c.Logger,
		CompileOpts: c.CompileOpts,
	}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: log.GetLogger(log.LoggerPrefixCompilation),
	}
	return &res
}
